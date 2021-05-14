package execute

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/turbot/steampipe/workspace"

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

	workspace *workspace.Workspace
	group     *ResultGroup
}

func NewControlRun(control *modconfig.Control, group *ResultGroup, workspace *workspace.Workspace) *ControlRun {
	return &ControlRun{
		Control:   control,
		Result:    NewResult(control),
		workspace: workspace,
		runStatus: ControlRunReady,
		group:     group,
		doneChan:  make(chan bool, 1),
	}
}

func (r *ControlRun) Start(ctx context.Context, client *db.Client) {
	r.runStatus = ControlRunStarted

	control := r.Control
	// resolve the query parameter of the control
	query, _ := execute.GetQueryFromArg(typeHelpers.SafeString(control.SQL), r.workspace)
	if query == "" {
		r.SetError(fmt.Errorf(`cannot run %s - failed to resolve query "%s"`, control.Name(), typeHelpers.SafeString(control.SQL)))
		return
	}

	startTime := time.Now()
	queryResult, err := client.ExecuteQuery(ctx, query, false)
	if err != nil {
		r.SetError(err)
		return
	}
	// validate required columns
	r.queryResult = queryResult
	if err := r.ValidateColumns(); err != nil {
		r.SetError(err)
		return
	}

	// set the control as started
	go r.gatherResults(queryResult)

	// TEMPORARY - we will eventually pass the streams to the renderer before completion
	// wait for control to finish
	controlCompletionTimeout := 240 * time.Second
	for {
		// if the control is finished (either successfully or with an error), return the controlRun
		if r.Finished() {
			break
		}
		time.Sleep(50 * time.Millisecond)
		if time.Since(startTime) > controlCompletionTimeout {
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
