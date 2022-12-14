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

type RuntimeDependencyPublisherImpl struct {
	DashboardParentImpl
	Args   []any                 `json:"args,omitempty"`
	Params []*modconfig.ParamDef `json:"params,omitempty"`

	// map of runtime dependencies, keyed by dependency long name
	runtimeDependencies map[string]*dashboardtypes.ResolvedRuntimeDependency
	subscriptions       map[string][]*RuntimeDependencyPublishTarget
	withValueMutex      sync.Mutex
	withRuns            map[string]*LeafRun
	inputs              map[string]*modconfig.DashboardInput
}

func NewRuntimeDependencyPublisherImpl(resource modconfig.DashboardLeafNode, parent dashboardtypes.DashboardParent, executionTree *DashboardExecutionTree) *RuntimeDependencyPublisherImpl {
	b := &RuntimeDependencyPublisherImpl{
		DashboardParentImpl: DashboardParentImpl{
			DashboardTreeRunImpl: NewDashboardTreeRunImpl(resource, parent, executionTree),
		},
		subscriptions:       make(map[string][]*RuntimeDependencyPublishTarget),
		runtimeDependencies: make(map[string]*dashboardtypes.ResolvedRuntimeDependency),
		inputs:              make(map[string]*modconfig.DashboardInput),
		withRuns:            make(map[string]*LeafRun),
	}
	// if the resource is a query provider, get params and set status
	if queryProvider, ok := resource.(modconfig.QueryProvider); ok {
		// get params
		b.Params = queryProvider.GetParams()
		if queryProvider.RequiresExecution(queryProvider) || len(queryProvider.GetChildren()) > 0 {
			b.Status = dashboardtypes.DashboardRunReady
		}
	}

	return b
}

func (b *RuntimeDependencyPublisherImpl) initRuntimeDependencies() error {
	// if the resource is a runtime dependency provider, create with runs and resolve dependencies
	if rdp, ok := b.resource.(modconfig.RuntimeDependencyProvider); ok {
		// if we have with blocks, create runs for them
		// BEFORE creating child runs, and before adding runtime dependencies
		err := b.createWithRuns(rdp.GetWiths(), b.executionTree)
		if err != nil {
			return err
		}
		// resolve any runtime dependencies
		if err := b.resolveRuntimeDependencies(rdp); err != nil {
			return err
		}
	}
	return nil
}

func (b *RuntimeDependencyPublisherImpl) Initialise(context.Context) {}

func (b *RuntimeDependencyPublisherImpl) Execute(context.Context) {
	panic("must be implemented by child struct")
}

func (b *RuntimeDependencyPublisherImpl) SetError(context.Context, error) {
	panic("must be implemented by child struct")
}

func (b *RuntimeDependencyPublisherImpl) SetComplete(context.Context) {
	panic("must be implemented by child struct")
}

func (b *RuntimeDependencyPublisherImpl) AsTreeNode() *dashboardtypes.SnapshotTreeNode {
	panic("must be implemented by child struct")
}

func (b *RuntimeDependencyPublisherImpl) GetName() string {
	return b.Name
}

func (b *RuntimeDependencyPublisherImpl) ProvidesRuntimeDependency(dependency *modconfig.RuntimeDependency) bool {
	resourceName := dependency.SourceResourceName()
	switch dependency.PropertyPath.ItemType {
	case modconfig.BlockTypeWith:
		return b.withRuns[resourceName] != nil
	case modconfig.BlockTypeInput:
		return b.inputs[resourceName] != nil
	case modconfig.BlockTypeParam:
		for _, p := range b.Params {
			// check short name not resource name (which is unqualified name)
			if p.ShortName == dependency.PropertyPath.Name {
				return true
			}
		}
	}
	return false
}

func (b *RuntimeDependencyPublisherImpl) SubscribeToRuntimeDependency(name string, opts ...RuntimeDependencyPublishOption) chan *dashboardtypes.ResolvedRuntimeDependencyValue {
	target := &RuntimeDependencyPublishTarget{
		// make a channel (buffer to avoid potential sync issues)
		channel: make(chan *dashboardtypes.ResolvedRuntimeDependencyValue, 1),
	}
	for _, o := range opts {
		o(target)
	}
	log.Printf("[TRACE] SubscribeToRuntimeDependency %s", name)

	// subscribe, passing a function which invokes getWithValue to resolve the required with value
	b.subscriptions[name] = append(b.subscriptions[name], target)
	return target.channel
}

func (b *RuntimeDependencyPublisherImpl) PublishRuntimeDependencyValue(name string, result *dashboardtypes.ResolvedRuntimeDependencyValue) {
	for _, target := range b.subscriptions[name] {
		if target.transform != nil {
			// careful not to mutate result which may be reused
			target.channel <- target.transform(result)
		} else {
			target.channel <- result
		}
		close(target.channel)
	}
	// clear subscriptions
	delete(b.subscriptions, name)
}

func (b *RuntimeDependencyPublisherImpl) FindRuntimeDependenciesForParent(parentProperty string) []*dashboardtypes.ResolvedRuntimeDependency {
	var res []*dashboardtypes.ResolvedRuntimeDependency
	for _, dep := range b.runtimeDependencies {
		if dep.Dependency.ParentPropertyName == parentProperty {
			res = append(res, dep)
		}
	}
	return res
}

func (b *RuntimeDependencyPublisherImpl) FindRuntimeDependencyForParent(parentProperty string) *dashboardtypes.ResolvedRuntimeDependency {
	res := b.FindRuntimeDependenciesForParent(parentProperty)
	if len(res) > 1 {
		panic(fmt.Sprintf("FindRuntimeDependencyForParent for %s, parent property %s, returned more that 1 result", b.Name, parentProperty))
	}
	if res == nil {
		return nil
	}
	// return first result
	return res[0]
}

func (b *RuntimeDependencyPublisherImpl) GetWithRuns() map[string]*LeafRun {
	return b.withRuns
}

// if this node has runtime dependencies, find the publisher of the dependency and create a dashboardtypes.ResolvedRuntimeDependency
// which  we use to resolve the values
func (b *RuntimeDependencyPublisherImpl) resolveRuntimeDependencies(rdp modconfig.RuntimeDependencyProvider) error {
	runtimeDependencies := rdp.GetRuntimeDependencies()
	for n, d := range runtimeDependencies {
		// find a runtime dependency publisher who can provider this runtime dependency
		publisher := b.findRuntimeDependencyPublisher(d)
		if publisher == nil {
			// should never happen as validation should have caught this
			return fmt.Errorf("cannot resolve runtime dependency %s", d.String())
		}

		// read name and dep into local loop vars to ensure correct value used when transform func is invoked
		name := n
		dep := d

		// determine the function to use to retrieve the runtime dependency value
		var opts []RuntimeDependencyPublishOption

		switch dep.PropertyPath.ItemType {
		case modconfig.BlockTypeWith:
			// set a transform function to extract the requested with data
			opts = append(opts, WithTransform(func(resolvedVal *dashboardtypes.ResolvedRuntimeDependencyValue) *dashboardtypes.ResolvedRuntimeDependencyValue {
				transformedResolvedVal := &dashboardtypes.ResolvedRuntimeDependencyValue{Error: resolvedVal.Error}
				if resolvedVal.Error == nil {
					// the runtime dependency value for a 'with' is *dashboardtypes.LeafData
					withValue, err := b.getWithValue(name, resolvedVal.Value.(*dashboardtypes.LeafData), dep.PropertyPath)
					if err != nil {
						transformedResolvedVal.Error = fmt.Errorf("failed to resolve with value '%s' for %s: %s", dep.PropertyPath.Original, name, err.Error())
					} else {
						transformedResolvedVal.Value = withValue
					}
				}
				return transformedResolvedVal
			}))
		}
		// subscribe, passing a function which invokes getWithValue to resolve the required with value
		valueChannel := publisher.SubscribeToRuntimeDependency(d.SourceResourceName(), opts...)

		publisherName := publisher.GetName()
		b.runtimeDependencies[name] = dashboardtypes.NewResolvedRuntimeDependency(dep, valueChannel, publisherName)
	}
	return nil
}

// getWithValue accepts the raw with result (dashboardtypes.LeafData) and the property path, and extracts the appropriate data
func (b *RuntimeDependencyPublisherImpl) getWithValue(name string, result *dashboardtypes.LeafData, path *modconfig.ParsedPropertyPath) (any, error) {
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

func (b *RuntimeDependencyPublisherImpl) setWithValue(w *LeafRun) {
	b.withValueMutex.Lock()
	defer b.withValueMutex.Unlock()

	name := w.Resource.GetUnqualifiedName()
	// if there was an error, w.Data will be nil and w.error will be non-nil
	result := &dashboardtypes.ResolvedRuntimeDependencyValue{Error: w.err}

	if w.err == nil {
		populateData(w.Data, result)
	}
	b.PublishRuntimeDependencyValue(name, result)
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

func (b *RuntimeDependencyPublisherImpl) withsComplete() bool {
	for _, w := range b.withRuns {
		if !w.RunComplete() {
			return false
		}
	}
	return true
}

func (b *RuntimeDependencyPublisherImpl) findRuntimeDependencyPublisher(runtimeDependency *modconfig.RuntimeDependency) RuntimeDependencyPublisher {
	// the runtime dependency publisher is usually the root node of the execution tree
	// the exception to this is if the node is a LeafRun(?) for a base node which has a with block,
	// in which case it may provide its own runtime depdency

	// try ourselves
	if b.ProvidesRuntimeDependency(runtimeDependency) {
		return b
	}

	// try root node
	// NOTE: we cannot just use b.executionTree.Root as this function is called at init time before Root is assigned
	rootPublisher := b.getRoot().(RuntimeDependencyPublisher)

	if rootPublisher.ProvidesRuntimeDependency(runtimeDependency) {
		return rootPublisher
	}

	return nil
}

// get the root of the tree by searching up the parents
func (b *RuntimeDependencyPublisherImpl) getRoot() dashboardtypes.DashboardTreeRun {
	var root dashboardtypes.DashboardTreeRun = b
	for {
		parent := root.GetParent()
		if parent == b.executionTree {
			break
		}
		root = parent.(dashboardtypes.DashboardTreeRun)
	}
	return root
}

func (b *RuntimeDependencyPublisherImpl) createWithRuns(withs []*modconfig.DashboardWith, executionTree *DashboardExecutionTree) error {
	for _, w := range withs {
		withRun, err := NewLeafRun(w, b, executionTree)
		if err != nil {
			return err
		}
		// NOTE: set the name of the run toe be the scoped name
		withRun.Name = fmt.Sprintf("%s.%s", withRun.GetParent().GetName(), w.UnqualifiedName)
		// set an onComplete function to populate 'with' data
		withRun.onComplete = func() { b.setWithValue(withRun) }

		b.withRuns[w.UnqualifiedName] = withRun
		b.children = append(b.children, withRun)
	}
	return nil
}
