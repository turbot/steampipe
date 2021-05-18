package execute

import (
	"context"
	"fmt"
	"sync"
	"time"

	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/query/execute"

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

	group         *ResultGroup
	executionTree *ExecutionTree
}

func NewControlRun(control *modconfig.Control, group *ResultGroup, executionTree *ExecutionTree) *ControlRun {
	return &ControlRun{
		Control:       control,
		Result:        NewResult(control),
		executionTree: executionTree,
		runStatus:     ControlRunReady,
		group:         group,
		doneChan:      make(chan bool, 1),
	}
}

func (r *ControlRun) Start(ctx context.Context, client *db.Client) {
	r.runStatus = ControlRunStarted

	control := r.Control

	//log.Println("[WARN]", "start", r.Control.ShortName)
	// update the current running control in the Progress renderer
	r.executionTree.progress.OnControlStart(control)

	// resolve the query parameter of the control
	query, _ := execute.GetQueryFromArg(typeHelpers.SafeString(control.SQL), r.executionTree.workspace)
	if query == "" {
		r.SetError(fmt.Errorf(`cannot run %s - failed to resolve query "%s"`, control.Name(), typeHelpers.SafeString(control.SQL)))
		return
	}

	startTime := time.Now()

	queryResult, err := client.ExecuteQuery(ctx, query, false)
	if err != nil {
		//log.Println("[WARN]", "set run error", r.Control.ShortName)
		r.SetError(err)
		return
	}
	// validate required columns
	r.queryResult = queryResult

	// set the control as started
	go r.gatherResults(queryResult)

	// TEMPORARY - we will eventually pass the streams to the renderer before completion
	// wait for control to finish
	controlCompletionTimeout := 240 * time.Second
	for {
		// if the control is finished (either successfully or with an error), return the controlRun
		if r.Finished() {
			//log.Println("[WARN]", "finished", r.Control.ShortName)

			// we cannot do this until the query has finished - otherwise we would leave the query lock locked
			// TODO we need a way to cancel
			if err := r.ValidateColumns(); err != nil {
				//log.Println("[WARN]", "set run error", r.Control.ShortName)
				r.SetError(err)
				return
			}

			break
		}
		time.Sleep(50 * time.Millisecond)
		if time.Since(startTime) > controlCompletionTimeout {
			// TODO we need a way to cancel
			r.SetError(fmt.Errorf("control %s timed out", control.Name()))
		}
	}

}

func (r *ControlRun) gatherResults(result *queryresult.Result) {
	for {

		select {
		case row := <-*r.queryResult.RowChan:
			if row == nil {
				// update the result group status with our status - this will be passed all the way up the execution tree
				r.group.updateSummary(r.Summary)
				// nil row means we are done
				//log.Println("[WARN]", "set complete", r.Control.ShortName)
				r.setRunStatus(ControlRunComplete)
				return

			}
			result, err := NewResultRow(r.Control, row, result.ColTypes)
			if err != nil {
				// fail on error
				//log.Println("[WARN]", "set error", r.Control.ShortName)
				r.SetError(err)
				continue
			}
			r.addResultRow(result)
		case <-r.doneChan:
			//log.Println("[WARN]", "close goroutine", r.Control.Title)
			return
		default:
			time.Sleep(25 * time.Millisecond)
		}
	}
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
	r.runStatus = status
	r.stateLock.Unlock()

	if r.Finished() {

		// update Progress
		if status == ControlRunError {
			r.executionTree.progress.OnError()
		} else {
			r.executionTree.progress.OnComplete()
		}

		// TODO CANCEL QUERY IF NEEDED
		r.doneChan <- true
	}
}

func (r *ControlRun) GetRunStatus() ControlRunStatus {
	r.stateLock.Lock()
	defer r.stateLock.Unlock()
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
