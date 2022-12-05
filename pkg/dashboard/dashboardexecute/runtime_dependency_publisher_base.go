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
	Params         []*modconfig.ParamDef `json:"params,omitempty"`
	subscriptions  map[string][]*RuntimeDependencyPublishTarget
	withValueMutex sync.Mutex
	withRuns       []*LeafRun
	parent         modconfig.ModTreeItem
}

func NewRuntimeDependencyPublisherBase() *RuntimeDependencyPublisherBase {
	return &RuntimeDependencyPublisherBase{subscriptions: make(map[string][]*RuntimeDependencyPublishTarget)}
}

func (r *RuntimeDependencyPublisherBase) SubscribeToRuntimeDependency(name string, opts ...RuntimeDependencyPublishOption) chan *dashboardtypes.ResolvedRuntimeDependencyValue {
	target := &RuntimeDependencyPublishTarget{
		// make a channel (buffer to avoid potential sync issues)
		channel: make(chan *dashboardtypes.ResolvedRuntimeDependencyValue, 1),
	}
	for _, o := range opts {
		o(target)
	}
	log.Printf("[TRACE] SubscribeToRuntimeDependency %s", name)

	// subscribe, passing a function which invokes getWithValue to resolve the required with value
	r.subscriptions[name] = append(r.subscriptions[name], target)
	return target.channel
}

func (r *RuntimeDependencyPublisherBase) PublishRuntimeDependencyValue(name string, result *dashboardtypes.ResolvedRuntimeDependencyValue) {
	for _, target := range r.subscriptions[name] {
		if target.transform != nil {
			result = target.transform(result)
		}
		target.channel <- (result)
		close(target.channel)
	}
	// clear subscriptions
	delete(r.subscriptions, name)
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

// getWithValue accepts the raw with result (LeafData) and the property path, and extracts the appropriate data
func (r *RuntimeDependencyPublisherBase) getWithValue(name string, result *dashboardtypes.LeafData, path *modconfig.ParsedPropertyPath) (any, error) {
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

// if this leaf run has with runs), execute them
func (r *RuntimeDependencyPublisherBase) executeWithRuns(ctx context.Context, childCompleteChan chan dashboardtypes.DashboardNodeRun) {
	if len(r.withRuns) == 0 {
		return
	}

	// asynchronously execute all with runs
	for _, w := range r.withRuns {
		go w.Execute(ctx)
	}

	// wait for withs to complete
	for !r.allWithsComplete() {
		completeChild := <-childCompleteChan
		// set the with value (this will set error value for the 'with' if execute failed)
		r.setWithValue(completeChild.(*LeafRun))
		// fall through to recheck ChildrenComplete
	}

	log.Printf("[TRACE] run %s ALL with runs complete")

}

func (r *RuntimeDependencyPublisherBase) setWithValue(w *LeafRun) {
	r.withValueMutex.Lock()
	defer r.withValueMutex.Unlock()

	name := w.DashboardNode.GetUnqualifiedName()
	// if there was an error, w.Data will be nil and w.error will be non-nil
	result := &dashboardtypes.ResolvedRuntimeDependencyValue{Error: w.error}

	if w.error == nil {
		populateData(w.Data, result)
	}
	r.PublishRuntimeDependencyValue(name, result)
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

func (r *RuntimeDependencyPublisherBase) allWithsComplete() bool {
	for _, w := range r.withRuns {
		if !w.RunComplete() {
			return false
		}
	}
	return true
}

func (r *RuntimeDependencyPublisherBase) findRuntimeDependencyPublisher(runtimeDependency *modconfig.RuntimeDependency) RuntimeDependencyPublisher {
	return r
}

//
//func (b *RuntimeDependencyPublisherBase) WalkParentPublishers(parentFunc func(RuntimeDependencyPublisher) (bool, error)) error {
//	for continueWalking := true; continueWalking; {
//		if parent := b.GetParentPublisher(); parent != nil {
//			var err error
//			continueWalking, err = parentFunc(parent)
//			if err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}
//
//func (b *RuntimeDependencyPublisherBase) ResolveWithFromTree(name string) (*DashboardWith, bool) {
//
//	b.WalkParentPublishers(func(RuntimeDependencyPublisher) (bool, error)){
//
//	}
//	w, ok := b.withs[name]
//	if !ok {
//		parent := b.GetParentPublisher()
//		if parent != nil {
//			return parent.ResolveWithFromTree(name)
//		}
//	}
//	return w, ok
//}
//
//func (b *RuntimeDependencyPublisherBase) ResolveParamFromTree(name string) (any, bool) {
//	// TODO
//	return nil, false
//}
//
//func (b *RuntimeDependencyPublisherBase) GetParentPublisher() RuntimeDependencyPublisher {
//	parent := b.parent
//	for parent != nil {
//		if res, ok := parent.(RuntimeDependencyPublisher); ok {
//			return res
//		}
//		if grandparents := parent.GetParents(); len(grandparents) > 0 {
//			parent = grandparents[0]
//		}
//	}
//	return nil
//}
