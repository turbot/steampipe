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

type RuntimeDependencyPublisherBase struct {
	Args   []any                 `json:"args,omitempty"`
	Params []*modconfig.ParamDef `json:"params,omitempty"`

	runtimeDependencies map[string]*dashboardtypes.ResolvedRuntimeDependency
	subscriptions       map[string][]*RuntimeDependencyPublishTarget
	withValueMutex      sync.Mutex
	withRuns            map[string]*LeafRun
	inputs              map[string]*modconfig.DashboardInput
	parent              dashboardtypes.DashboardNodeParent
}

func NewRuntimeDependencyPublisherBase(parent dashboardtypes.DashboardNodeParent) *RuntimeDependencyPublisherBase {
	return &RuntimeDependencyPublisherBase{
		subscriptions:       make(map[string][]*RuntimeDependencyPublishTarget),
		runtimeDependencies: make(map[string]*dashboardtypes.ResolvedRuntimeDependency),
		inputs:              make(map[string]*modconfig.DashboardInput),
		parent:              parent,
	}
}

func (b *RuntimeDependencyPublisherBase) ProvidesRuntimeDependency(dependency *modconfig.RuntimeDependency) bool {
	resourceName := dependency.SourceResourceName()
	switch dependency.PropertyPath.ItemType {
	case modconfig.BlockTypeWith:
		return b.withRuns[resourceName] != nil
	case modconfig.BlockTypeInput:
		return b.inputs[resourceName] != nil
	case modconfig.BlockTypeParam:
		for _, p := range b.Params {
			// check short name not resource name (which is unqualified name)
			if p.Name == dependency.PropertyPath.Name {
				return true
			}
		}
	}
	return false
}

func (b *RuntimeDependencyPublisherBase) SubscribeToRuntimeDependency(name string, opts ...RuntimeDependencyPublishOption) chan *dashboardtypes.ResolvedRuntimeDependencyValue {
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

func (b *RuntimeDependencyPublisherBase) PublishRuntimeDependencyValue(name string, result *dashboardtypes.ResolvedRuntimeDependencyValue) {
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

// if this node has runtime dependencies, create runtime dependency instances which we use to resolve the values
func (b *RuntimeDependencyPublisherBase) addRuntimeDependencies(resource modconfig.DashboardLeafNode) error {
	// only QueryProvider resources support runtime dependencies
	queryProvider, ok := resource.(modconfig.RuntimeDependencyProvider)
	if !ok {
		return nil
	}
	runtimeDependencies := queryProvider.GetRuntimeDependencies()
	for n, d := range runtimeDependencies {
		// find a runtime depdency publisher who can provider this runtime depdency
		publisher := b.findRuntimeDependencyPublisher(d)
		if publisher == nil {
			// should never happen as validation should have caught this
			return fmt.Errorf("cannot resolve runtime dependency %s", d.String())
		}

		// read name and dep into local loop vars to ensure correct value used when getValueFunc is invoked
		name := n
		dep := d
		// determine the function to use to retrieve the runtime dependency value
		var opts []RuntimeDependencyPublishOption

		switch dep.PropertyPath.ItemType {
		case modconfig.BlockTypeWith:
			// set a transform function to extract the requested with data
			opts = append(opts, WithTransform(func(val *dashboardtypes.ResolvedRuntimeDependencyValue) *dashboardtypes.ResolvedRuntimeDependencyValue {
				res := &dashboardtypes.ResolvedRuntimeDependencyValue{Error: val.Error}
				if val.Error == nil {
					// the runtime dependency value for a 'with' is *LeafData
					res.Value, res.Error = b.getWithValue(name, val.Value.(*dashboardtypes.LeafData), dep.PropertyPath)
				}
				return res
			}))

		}
		// subscribe, passing a function which invokes getWithValue to resolve the required with value
		valueChannel := publisher.SubscribeToRuntimeDependency(d.SourceResourceName(), opts...)
		b.runtimeDependencies[name] = dashboardtypes.NewResolvedRuntimeDependency(dep, valueChannel)
	}
	return nil
}

// getWithValue accepts the raw with result (LeafData) and the property path, and extracts the appropriate data
func (b *RuntimeDependencyPublisherBase) getWithValue(name string, result *dashboardtypes.LeafData, path *modconfig.ParsedPropertyPath) (any, error) {
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

// if this leaf run has with runs), execute them
func (b *RuntimeDependencyPublisherBase) executeWithRuns(ctx context.Context, childCompleteChan chan dashboardtypes.DashboardNodeRun) {
	if len(b.withRuns) == 0 {
		return
	}

	// asynchronously execute all with runs
	for _, w := range b.withRuns {
		go w.Execute(ctx)
	}

	// wait for withs to complete
	for !b.allWithsComplete() {
		completeChild := <-childCompleteChan
		// set the with value (this will set error value for the 'with' if execute failed)
		b.setWithValue(completeChild.(*LeafRun))
		// fall through to recheck ChildrenComplete
	}

	log.Printf("[TRACE] run %s ALL with runs complete")

}

func (b *RuntimeDependencyPublisherBase) setWithValue(w *LeafRun) {
	b.withValueMutex.Lock()
	defer b.withValueMutex.Unlock()

	name := w.DashboardNode.GetUnqualifiedName()
	// if there was an error, w.Data will be nil and w.error will be non-nil
	result := &dashboardtypes.ResolvedRuntimeDependencyValue{Error: w.error}

	if w.error == nil {
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

func (b *RuntimeDependencyPublisherBase) allWithsComplete() bool {
	for _, w := range b.withRuns {
		if !w.RunComplete() {
			return false
		}
	}
	return true
}

func (b *RuntimeDependencyPublisherBase) findRuntimeDependencyPublisher(runtimeDependency *modconfig.RuntimeDependency) RuntimeDependencyPublisher {
	var res RuntimeDependencyPublisher
	b.WalkParentPublishers(func(p RuntimeDependencyPublisher) (bool, error) {
		if p.ProvidesRuntimeDependency(runtimeDependency) {
			res = p
			return false, nil
		}
		return true, nil
	})
	return res
}
func (b *RuntimeDependencyPublisherBase) WalkParentPublishers(parentFunc func(RuntimeDependencyPublisher) (bool, error)) error {
	for continueWalking := true; continueWalking; {
		if parent := b.GetParentPublisher(); parent != nil {
			var err error
			continueWalking, err = parentFunc(parent)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (b *RuntimeDependencyPublisherBase) GetParentPublisher() RuntimeDependencyPublisher {
	parent := b.parent
	for parent != nil {
		if res, ok := parent.(RuntimeDependencyPublisher); ok {
			return res
		}
		if grandparent := parent.GetParent(); grandparent != nil {
			parent = grandparent
		}
	}
	return nil
}
