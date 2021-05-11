package controlresult

import (
	"fmt"
	"sync"
	"time"

	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
)

type ControlRunStatus uint32

const (
	ControlRunReady ControlRunStatus = 1 << iota
	ControlRunStarted
	ControlRunComplete
	ControlRunError
)

// Result is the result of a control run - will contain one or more result items (i.e. for one or more resources)
type Result struct {
	status ControlRunStatus
	Error  error
	// any completed results items
	Rows []*ResultRow
	// the parent control
	Control *modconfig.Control
	// the query result stream
	queryResult *queryresult.Result
	stateLock   sync.Mutex
	doneChan    chan bool
	// parent in the result tree
	//parentResult
}

func NewControlResult(control *modconfig.Control) *Result {
	return &Result{
		Control:  control,
		status:   ControlRunReady,
		doneChan: make(chan bool, 1),
	}
}

func (r *Result) Start(result *queryresult.Result) {
	// validate required columns
	r.queryResult = result
	if err := r.ValidateColumns(); err != nil {
		r.SetError(err)
		return
	}

	r.queryResult = result
	r.status = ControlRunStarted
	go func() {
		for {
			select {
			case row := <-*r.queryResult.RowChan:
				if row == nil {
					// nil row means we are done
					r.setStatus(ControlRunComplete)
					break
				}
				result, err := NewResultRow(r.Control, row, result.ColTypes)
				if err != nil {
					// fail on error
					r.SetError(err)
					continue
				}
				r.Rows = append(r.Rows, result)
			case <-r.doneChan:
				return
			default:
				time.Sleep(25 * time.Millisecond)
			}
		}
	}()
}

func (r *Result) SetError(err error) {
	r.Error = err
	r.setStatus(ControlRunError)
}

func (r *Result) setStatus(status ControlRunStatus) {
	r.stateLock.Lock()
	defer r.stateLock.Unlock()
	r.status = status
	if r.Finished() {
		// TODO CANCEL QUERY IF NEEDED
		r.doneChan <- true
	}
}

func (r *Result) GetStatus() ControlRunStatus {
	return r.status
}

func (r *Result) Finished() bool {
	status := r.GetStatus()
	return status == ControlRunComplete || status == ControlRunError
}

func (r *Result) ValidateColumns() error {
	// validate columns
	requiredColumns := []string{"reason", "resource", "status"}
	var missingColumns []string
	for _, col := range requiredColumns {
		if !r.ColumnTypesContainsColumn(col) {
			missingColumns = append(missingColumns, col)
		}
	}
	if len(missingColumns) > 0 {
		return fmt.Errorf("control result is missing required %s: %v", utils.Pluralize("column", len(missingColumns)), missingColumns)
	}
	return nil
}

func (r *Result) ColumnTypesContainsColumn(col string) bool {
	return r.queryResult.ColumnTypesContainsColumn(col)
}
