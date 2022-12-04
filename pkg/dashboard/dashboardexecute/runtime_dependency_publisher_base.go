package dashboardexecute

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"log"
	"strconv"
	"sync"
)

type RuntimeDependencyPublisherBase struct {
	subscriptions  map[string][]*RuntimeDependencyPublishTarget
	withValueMutex sync.Mutex
	withRuns       []*LeafRun
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

func columnValuesFromRows(column string, rows []map[string]interface{}) (any, error) {
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
func (r *RuntimeDependencyPublisherBase) executeWithRuns(ctx context.Context, childCompleteChan chan dashboardtypes.DashboardNodeRun) error {
	for _, w := range r.withRuns {
		go w.Execute(ctx)
	}
	// wait for children to complete
	var errors []error

	for !r.withComplete() {

		completeChild := <-childCompleteChan
		log.Printf("[TRACE] run %s got with complete")
		if completeChild.GetRunStatus() == dashboardtypes.DashboardRunError {
			errors = append(errors, completeChild.GetError())
		}
		// fall through to recheck ChildrenComplete
	}

	log.Printf("[TRACE] run %s ALL children complete")
	// so all with runs have completed - check for errors
	err := error_helpers.CombineErrors(errors...)
	if err == nil {
		err = r.setWithData()
	}

	// return error (is any)
	return err
}

func (r *RuntimeDependencyPublisherBase) setWithData() error {
	for _, w := range r.withRuns {
		if err := r.setWithValue(w); err != nil {
			return err
		}
	}
	return nil
}

func (r *RuntimeDependencyPublisherBase) setWithValue(w *LeafRun) error {
	r.withValueMutex.Lock()
	defer r.withValueMutex.Unlock()

	name := w.DashboardNode.GetUnqualifiedName()
	result := &dashboardtypes.ResolvedRuntimeDependencyValue{Value: w.Data, Error: w.error}

	// TACTICAL - is there are any JSON columns convert them back to a JSON string
	var jsonColumns []string
	for _, c := range w.Data.Columns {
		if c.DataType == "JSONB" || c.DataType == "JSON" {
			jsonColumns = append(jsonColumns, c.Name)
		}
	}
	// now convert any json values into a json string
	for _, c := range jsonColumns {
		for _, row := range w.Data.Rows {
			jsonBytes, err := json.Marshal(row[c])
			if err != nil {
				return err
			}
			row[c] = string(jsonBytes)
		}
	}
	r.PublishRuntimeDependencyValue(name, result)
	return nil
}

func (r *RuntimeDependencyPublisherBase) withComplete() bool {
	for _, w := range r.withRuns {
		if !w.RunComplete() {
			return false
		}
	}
	return true
}
