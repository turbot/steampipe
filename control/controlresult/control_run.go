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

// ControlRun is a struct representing a  a control run - will contain one or more result items (i.e. for one or more resources)
type ControlRun struct {
	Result *Result
	Error  error
	// the parent control
	Control *modconfig.Control

	Summary StatusSummary
	// the query result stream
	queryResult *queryresult.Result
	runStatus   ControlRunStatus
	stateLock   sync.Mutex
	doneChan    chan bool
}

func NewControlRun(control *modconfig.Control) *ControlRun {
	return &ControlRun{
		Control:   control,
		Result:    NewResult(control),
		runStatus: ControlRunReady,
		doneChan:  make(chan bool, 1),
	}
}

func (r *ControlRun) Start(result *queryresult.Result) {
	// validate required columns
	r.queryResult = result
	if err := r.ValidateColumns(); err != nil {
		r.SetError(err)
		return
	}

	r.queryResult = result
	r.runStatus = ControlRunStarted
	go func() {
		for {
			select {
			case row := <-*r.queryResult.RowChan:
				if row == nil {
					// nil row means we are done
					r.setRunStatus(ControlRunComplete)
					break
				}
				result, err := NewResultRow(r.Control, row, result.ColTypes)
				if err != nil {
					// fail on error
					r.SetError(err)
					continue
				}
				r.addResultRow(result)
			case <-r.doneChan:
				return
			default:
				time.Sleep(25 * time.Millisecond)
			}
		}
	}()
}

// add the result row to our results and update the summary with the row status
func (r *ControlRun) addResultRow(row *ResultRow) {
	// update results
	r.Result.addResultRow(row)

	// update summary
	switch row.Status {
	case ControlOk:
		r.Summary.Ok++
	case ControlAlarm:
		r.Summary.Alarm++
	case ControlSkip:
		r.Summary.Skip++
	case ControlInfo:
		r.Summary.Info++
	case ControlError:
		r.Summary.Error++
	}
}

func (r *ControlRun) SetError(err error) {
	r.Error = err
	r.setRunStatus(ControlRunError)
}

func (r *ControlRun) setRunStatus(status ControlRunStatus) {
	r.stateLock.Lock()
	defer r.stateLock.Unlock()
	r.runStatus = status
	if r.Finished() {
		// TODO CANCEL QUERY IF NEEDED
		r.doneChan <- true
	}
}

func (r *ControlRun) GetRunStatus() ControlRunStatus {
	return r.runStatus
}

func (r *ControlRun) Finished() bool {
	status := r.GetRunStatus()
	return status == ControlRunComplete || status == ControlRunError
}

func (r *ControlRun) ValidateColumns() error {
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

func (r *ControlRun) ColumnTypesContainsColumn(col string) bool {
	return r.queryResult.ColumnTypesContainsColumn(col)
}
