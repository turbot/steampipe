package dashboardexecute

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"log"
	"strconv"
	"sync"
)

type runtimeDependencyPublisherImpl struct {
	DashboardParentImpl
	Args           []any                 `json:"args,omitempty"`
	Params         []*modconfig.ParamDef `json:"params,omitempty"`
	subscriptions  map[string][]*RuntimeDependencyPublishTarget
	withValueMutex *sync.Mutex
	withRuns       map[string]*LeafRun
	inputs         map[string]*modconfig.DashboardInput
}

func newRuntimeDependencyPublisherImpl(resource modconfig.DashboardLeafNode, parent dashboardtypes.DashboardParent, run dashboardtypes.DashboardTreeRun, executionTree *DashboardExecutionTree) runtimeDependencyPublisherImpl {
	b := runtimeDependencyPublisherImpl{
		DashboardParentImpl: newDashboardParentImpl(resource, parent, run, executionTree),
		subscriptions:       make(map[string][]*RuntimeDependencyPublishTarget),
		inputs:              make(map[string]*modconfig.DashboardInput),
		withRuns:            make(map[string]*LeafRun),
		withValueMutex:      new(sync.Mutex),
	}
	// if the resource is a query provider, get params and set status
	if queryProvider, ok := resource.(modconfig.QueryProvider); ok {
		// get params
		b.Params = queryProvider.GetParams()
		if queryProvider.RequiresExecution(queryProvider) || len(queryProvider.GetChildren()) > 0 {
			b.Status = dashboardtypes.RunInitialized
		}
	}

	return b
}

func (p *runtimeDependencyPublisherImpl) Initialise(context.Context) {}

func (p *runtimeDependencyPublisherImpl) Execute(context.Context) {
	panic("must be implemented by child struct")
}

func (p *runtimeDependencyPublisherImpl) AsTreeNode() *dashboardtypes.SnapshotTreeNode {
	panic("must be implemented by child struct")
}

func (p *runtimeDependencyPublisherImpl) GetName() string {
	return p.Name
}

func (p *runtimeDependencyPublisherImpl) ProvidesRuntimeDependency(dependency *modconfig.RuntimeDependency) bool {
	resourceName := dependency.SourceResourceName()
	switch dependency.PropertyPath.ItemType {
	case modconfig.BlockTypeWith:
		// we cannot use withRuns here as if withs have dependencies on each other,
		// this function may be called before all runs have been added
		// instead, look directly at the underlying resource withs
		if wp, ok := p.resource.(modconfig.WithProvider); ok {
			for _, w := range wp.GetWiths() {
				if w.UnqualifiedName == resourceName {
					return true
				}
			}
		}
		return false
	case modconfig.BlockTypeInput:
		return p.inputs[resourceName] != nil
	case modconfig.BlockTypeParam:
		for _, p := range p.Params {
			// check short name not resource name (which is unqualified name)
			if p.ShortName == dependency.PropertyPath.Name {
				return true
			}
		}
	}
	return false
}

func (p *runtimeDependencyPublisherImpl) SubscribeToRuntimeDependency(name string, opts ...RuntimeDependencyPublishOption) chan *dashboardtypes.ResolvedRuntimeDependencyValue {
	target := &RuntimeDependencyPublishTarget{
		// make a channel (buffer to avoid potential sync issues)
		channel: make(chan *dashboardtypes.ResolvedRuntimeDependencyValue, 1),
	}
	for _, o := range opts {
		o(target)
	}
	log.Printf("[TRACE] SubscribeToRuntimeDependency %s", name)

	// subscribe, passing a function which invokes getWithValue to resolve the required with value
	p.subscriptions[name] = append(p.subscriptions[name], target)
	return target.channel
}

func (p *runtimeDependencyPublisherImpl) PublishRuntimeDependencyValue(name string, result *dashboardtypes.ResolvedRuntimeDependencyValue) {
	for _, target := range p.subscriptions[name] {
		if target.transform != nil {
			// careful not to mutate result which may be reused
			target.channel <- target.transform(result)
		} else {
			target.channel <- result
		}
		close(target.channel)
	}
	// clear subscriptions
	delete(p.subscriptions, name)
}

func (p *runtimeDependencyPublisherImpl) GetWithRuns() map[string]*LeafRun {
	return p.withRuns
}

func (p *runtimeDependencyPublisherImpl) initWiths() error {
	// if the resource is a runtime dependency provider, create with runs and resolve dependencies
	wp, ok := p.resource.(modconfig.WithProvider)
	if !ok {
		return nil
	}
	// if we have with blocks, create runs for them
	// BEFORE creating child runs, and before adding runtime dependencies
	err := p.createWithRuns(wp.GetWiths(), p.executionTree)
	if err != nil {
		return err
	}

	return nil
}

// getWithValue accepts the raw with result (dashboardtypes.LeafData) and the property path, and extracts the appropriate data
func (p *runtimeDependencyPublisherImpl) getWithValue(name string, result *dashboardtypes.LeafData, path *modconfig.ParsedPropertyPath) (any, error) {
	//  get the set of rows which will be used ot generate the return value
	rows := result.Rows
	/*
			You can
		reference the whole table with:
				with.stuff1
			this is equivalent to:
				with.stuff1.rows
			and
				with.stuff1.rows[*]

			Rows is a list, and you can index it to get a single row:
				with.stuff1.rows[0]
			or splat it to get all rows:
				with.stuff1.rows[*]
			Each row, in turn, contains all the columns, so you can get a single column of a single row:
				with.stuff1.rows[0].a
			if you splat the row, then you can get an array of a single column from all rows. This would be passed to sql as an array:
				with.stuff1.rows[*].a
	*/

	// with.stuff1 -> PropertyPath will be ""
	// with.stuff1.rows -> PropertyPath will be "rows"
	// with.stuff1.rows[*] -> PropertyPath will be "rows.*"
	// with.stuff1.rows[0] -> PropertyPath will be "rows.0"
	// with.stuff1.rows[0].a -> PropertyPath will be "rows.0.a"
	const rowsSegment = 0
	const rowsIdxSegment = 1
	const columnSegment = 2

	// second path section MUST  be "rows"
	if len(path.PropertyPath) > rowsSegment && path.PropertyPath[rowsSegment] != "rows" || len(path.PropertyPath) > (columnSegment+1) {
		return nil, fmt.Errorf("reference to with '%s' has invalid property path '%s'", name, path.Original)
	}

	// if no row is specified assume all
	rowIdxStr := "*"
	if len(path.PropertyPath) > rowsIdxSegment {
		// so there is 3rd part - this will be the row idx (or '*')
		rowIdxStr = path.PropertyPath[rowsIdxSegment]
	}
	var column string

	// is a column specified?
	if len(path.PropertyPath) > columnSegment {
		column = path.PropertyPath[columnSegment]
	} else {
		if len(result.Columns) > 1 {
			// we do not support returning all columns (yet
			return nil, fmt.Errorf("reference to with '%s' is returning more than one column - not supported", name)
		}
		column = result.Columns[0].Name
	}

	if rowIdxStr == "*" {
		return columnValuesFromRows(column, rows)
	}

	rowIdx, err := strconv.Atoi(rowIdxStr)
	if err != nil {
		return nil, fmt.Errorf("reference to with '%s' has invalid property path '%s' - cannot parse row idx '%s'", name, path.Original, rowIdxStr)
	}

	// do we have the requested row
	if rowCount := len(rows); rowIdx >= rowCount {
		return nil, fmt.Errorf("reference to with '%s' has invalid row index '%d' - %d %s were returned", name, rowIdx, rowCount, utils.Pluralize("row", rowCount))
	}
	// so we are returning a single row
	row := rows[rowIdx]
	return row[column], nil
}

func columnValuesFromRows(column string, rows []map[string]any) (any, error) {
	if column == "" {
		return nil, fmt.Errorf("columnValuesFromRows failed - no column specified")
	}
	var res = make([]any, len(rows))
	for i, row := range rows {
		var ok bool
		res[i], ok = row[column]
		if !ok {
			return nil, fmt.Errorf("column %s does not exist", column)
		}
	}
	return res, nil
}

func (p *runtimeDependencyPublisherImpl) setWithValue(w *LeafRun) {
	p.withValueMutex.Lock()
	defer p.withValueMutex.Unlock()

	name := w.resource.GetUnqualifiedName()
	// if there was an error, w.Data will be nil and w.error will be non-nil
	result := &dashboardtypes.ResolvedRuntimeDependencyValue{Error: w.err}

	if w.err == nil {
		populateData(w.Data, result)
	}
	p.PublishRuntimeDependencyValue(name, result)
	return
}

func populateData(withData *dashboardtypes.LeafData, result *dashboardtypes.ResolvedRuntimeDependencyValue) {
	result.Value = withData
	// TACTICAL - is there are any JSON columns convert them back to a JSON string
	var jsonColumns []string
	for _, c := range withData.Columns {
		if c.DataType == "JSONB" || c.DataType == "JSON" {
			jsonColumns = append(jsonColumns, c.Name)
		}
	}
	// now convert any json values into a json string
	for _, c := range jsonColumns {
		for _, row := range withData.Rows {
			jsonBytes, err := json.Marshal(row[c])
			if err != nil {
				// publish result with the error
				result.Error = err
				result.Value = nil
				return
			}
			row[c] = string(jsonBytes)
		}
	}
}

func (p *runtimeDependencyPublisherImpl) createWithRuns(withs []*modconfig.DashboardWith, executionTree *DashboardExecutionTree) error {
	for _, w := range withs {
		// NOTE: set the name of the run to be the scoped name
		withRunName := fmt.Sprintf("%s.%s", p.GetName(), w.UnqualifiedName)
		withRun, err := NewLeafRun(w, p, executionTree, setName(withRunName))
		if err != nil {
			return err
		}
		// set an onComplete function to populate 'with' data
		withRun.onComplete = func() { p.setWithValue(withRun) }

		p.withRuns[w.UnqualifiedName] = withRun
		p.children = append(p.children, withRun)
	}
	return nil
}

// called when the args are resolved - if anyone is subscribing to the args value, publish
func (p *runtimeDependencyPublisherImpl) argsResolved(args []any) {
	// use params to get param names for each arg and then look of subscriber
	for i, param := range p.Params {
		if i == len(args) {
			return
		}
		// do we have a subscription for this param
		if _, ok := p.subscriptions[param.UnqualifiedName]; ok {
			p.PublishRuntimeDependencyValue(param.UnqualifiedName, &dashboardtypes.ResolvedRuntimeDependencyValue{Value: args[i]})
		}
	}
	log.Printf("[TRACE] %s: argsResolved", p.Name)
}
