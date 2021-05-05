package results

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/turbot/steampipe/utils"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type ControlRunStatus uint32

const (
	ControlRunReady ControlRunStatus = 1 << iota
	ControlRunStarted
	ControlRunComplete
	ControlRunError
)

// ControlResult :: the result of a control run - will contain one or more result items (oi.e. for one or more resources)
type ControlResult struct {
	status ControlRunStatus
	Error  error
	// any completed results items
	Results []*ControlResultItem
	// the parent control
	Control *modconfig.Control
	// the query result stream
	queryResult *QueryResult
	stateLock   sync.Mutex
	doneChan    chan (bool)
}

func NewControlResult(control *modconfig.Control) *ControlResult {
	return &ControlResult{
		Control:  control,
		status:   ControlRunReady,
		doneChan: make(chan bool, 1),
	}
}

func (r *ControlResult) Start(result *QueryResult) {
	// validate required columns
	if err := r.ValidateColumns(result); err != nil {
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
				result, err := NewControlResultItem(r.Control, row, result.ColTypes)
				if err != nil {
					// fail on error
					r.SetError(err)
				}
				r.Results = append(r.Results, result)
			case <-r.doneChan:
				return
			default:
				time.Sleep(25 * time.Millisecond)
			}

		}
	}()

}

func (r *ControlResult) SetError(err error) {
	r.Error = err
	r.setStatus(ControlRunError)
}

func (r *ControlResult) setStatus(status ControlRunStatus) {
	r.stateLock.Lock()
	defer r.stateLock.Unlock()
	r.status = status
	if r.Finished() {
		r.doneChan <- true
	}

}

func (r *ControlResult) GetStatus() ControlRunStatus {
	return r.status
}

func (r *ControlResult) Finished() bool {
	status := r.GetStatus()
	return status == ControlRunComplete || status == ControlRunError
}

func (r *ControlResult) ValidateColumns(result *QueryResult) error {
	// validate columns
	requiredColumns := []string{"reason", "resource", "status"}
	var missingColumns []string
	for _, col := range requiredColumns {
		if !columnTypesContainsColumn(result.ColTypes, col) {
			missingColumns = append(missingColumns, col)
		}
	}
	if len(missingColumns) > 0 {
		return fmt.Errorf("control result is missing required %s: %v", utils.Pluralize("column", len(missingColumns)), missingColumns)
	}
	return nil
}

func columnTypesContainsColumn(colTypes []*sql.ColumnType, col string) bool {
	for _, ct := range colTypes {
		if ct.Name() == col {
			return true
		}
	}
	return false

}
