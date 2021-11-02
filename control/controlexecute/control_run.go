package controlexecute

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/turbot/steampipe/db/db_common"

	"github.com/spf13/viper"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/constants"
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
	runError error `json:"-"`
	// the number of attempts this control made to run
	attempts int `json:"-"`
	// the parent control
	Control *modconfig.Control `json:"-"`
	Summary StatusSummary      `json:"-"`

	// execution duration
	Duration time.Duration `json:"-"`

	// the result
	ControlId   string                  `json:"control_id"`
	Description string                  `json:"description"`
	Severity    string                  `json:"severity"`
	Tags        map[string]string       `json:"tags"`
	Title       string                  `json:"title"`
	RowMap      map[string][]*ResultRow `json:"-"`
	Rows        []*ResultRow            `json:"results"`

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
		Control: control,

		ControlId:   control.Name(),
		Description: typehelpers.SafeString(control.Description),
		Severity:    typehelpers.SafeString(control.Severity),
		Title:       typehelpers.SafeString(control.Title),
		Tags:        control.GetTags(),
		RowMap:      make(map[string][]*ResultRow),

		executionTree: executionTree,
		runStatus:     ControlRunReady,

		group:    group,
		doneChan: make(chan bool, 1),
	}
}

func (r *ControlRun) Skip() {
	r.setRunStatus(ControlRunComplete)
}

func (r *ControlRun) Execute(ctx context.Context, client db_common.Client) {
	log.Printf("[TRACE] begin ControlRun.Start: %s\n", r.Control.Name())
	defer log.Printf("[TRACE] end ControlRun.Start: %s\n", r.Control.Name())

	startTime := time.Now()

	r.runStatus = ControlRunStarted

	control := r.Control

	// update the current running control in the Progress renderer
	r.executionTree.progress.OnControlStart(control)

	// resolve the control query
	query, err := r.executionTree.workspace.ResolveControlQuery(control)
	if err != nil {
		r.SetError(err)
		return
	}
	if query == "" {
		r.SetError(fmt.Errorf(`cannot run %s - failed to resolve query "%s"`, control.Name(), typehelpers.SafeString(control.SQL)))
		return
	}

	// set a log line in the database logs for convenience
	// pass 'true' to disable spinner
	_, _ = client.ExecuteSync(ctx, fmt.Sprintf("--- Executing %s", control.GetTitle()), true)

	var originalConfiguredSearchPath []string
	var originalConfiguredSearchPathPrefix []string

	if control.SearchPath != nil || control.SearchPathPrefix != nil {
		originalConfiguredSearchPath = viper.GetViper().GetStringSlice(constants.ArgSearchPath)
		originalConfiguredSearchPathPrefix = viper.GetViper().GetStringSlice(constants.ArgSearchPathPrefix)

		if control.SearchPath != nil {
			viper.Set(constants.ArgSearchPath, strings.Split(*control.SearchPath, ","))
		}
		if control.SearchPathPrefix != nil {
			viper.Set(constants.ArgSearchPathPrefix, strings.Split(*control.SearchPathPrefix, ","))
		}

		fmt.Println("\n##### Setting search path:", viper.GetStringSlice(constants.ArgSearchPath), viper.GetStringSlice(constants.ArgSearchPathPrefix), ":::", r.ControlId)

		client.SetSessionSearchPath()
	}

	shouldBeDoneBy := time.Now().Add(240 * time.Second)
	ctx, cancel := context.WithDeadline(ctx, shouldBeDoneBy)

	// Even though ctx will expire, it is good practice to call its
	// cancellation function in any case. Not doing so may keep the
	// context and its parent alive longer than necessary.
	defer cancel()

	r.executionTree.progress.OnControlExecuteStart()
	queryResult, err := client.Execute(ctx, query, false)
	if err != nil {
		r.executionTree.progress.OnControlExecuteFinish()

		// is this an rpc EOF error - meaning that the plugin somehow crashed
		if constants.IsGRPCConnectivityError(err) {
			if r.attempts > constants.MaxControlRunAttempts {
				// if exceeded max retries, give up
				r.SetError(err)
				return
			}
			// set a log line in the database logs for convenience - pass 'true' to disable spinner
			_, _ = client.ExecuteSync(ctx, "-- Retrying...", true)

			// the control errored
			r.executionTree.progress.OnControlError()

			// recurse into this function to retry
			// use the same context, so that we respect the timeout
			r.attempts++
			r.Execute(ctx, client)
			return
		}
		r.SetError(err)
		return
	}
	// validate required columns
	r.queryResult = queryResult

	// create a channel to which will be closed when gathering has been done
	gatherDoneChan := make(chan string)
	go func() {
		r.gatherResults()
		close(gatherDoneChan)
	}()

	select {
	case <-ctx.Done():
		r.SetError(ctx.Err())
	case <-gatherDoneChan:
		// do nothing
	}

	// reset the search path
	if control.SearchPath != nil || control.SearchPathPrefix != nil {
		fmt.Println("\n##### Resetting search path:", originalConfiguredSearchPath, originalConfiguredSearchPathPrefix, ":::", r.ControlId)
		// the search path was modified. Reset it!
		viper.Set(constants.ArgSearchPath, originalConfiguredSearchPath)
		viper.Set(constants.ArgSearchPathPrefix, originalConfiguredSearchPathPrefix)
		client.SetSessionSearchPath()
	}

	r.Duration = time.Since(startTime)
	r.executionTree.progress.OnControlExecuteFinish()
}

func (r *ControlRun) gatherResults() {
	for {
		select {
		case row := <-*r.queryResult.RowChan:
			if row == nil {
				// update the result group status with our status - this will be passed all the way up the execution tree
				r.group.updateSummary(r.Summary)
				if len(r.Severity) != 0 {
					r.group.updateSeverityCounts(r.Severity, r.Summary)
				}

				// nil row means we are done
				r.setRunStatus(ControlRunComplete)
				r.createdOrderedResultRows()
				return

			}
			// got a result - send a ping over the channel so that the
			// loop can check against the timeout
			if row.Error != nil {
				r.SetError(row.Error)
				return
			}
			result, err := NewResultRow(r.Control, row, r.queryResult.ColTypes)
			if err != nil {
				r.SetError(err)
				return
			}
			r.addResultRow(result)
		case <-r.doneChan:
			return
		}
	}

}

// add the result row to our results and update the summary with the row status
func (r *ControlRun) addResultRow(row *ResultRow) {
	// update results
	r.RowMap[row.Status] = append(r.RowMap[row.Status], row)

	// update summary
	switch row.Status {
	case constants.ControlOk:
		r.Summary.Ok++
	case constants.ControlAlarm:
		r.Summary.Alarm++
	case constants.ControlSkip:
		r.Summary.Skip++
	case constants.ControlInfo:
		r.Summary.Info++
	case constants.ControlError:
		r.Summary.Error++
	}
}

// populate ordered list of rows
func (r *ControlRun) createdOrderedResultRows() {
	statusOrder := []string{constants.ControlError, constants.ControlAlarm, constants.ControlInfo, constants.ControlOk, constants.ControlSkip}
	for _, status := range statusOrder {
		r.Rows = append(r.Rows, r.RowMap[status]...)
	}
}

func (r *ControlRun) SetError(err error) {
	r.runError = utils.TransformErrorToSteampipe(err)

	// update error count
	r.Summary.Error++
	r.setRunStatus(ControlRunError)
}

func (r *ControlRun) GetError() error {
	return r.runError
}

func (r *ControlRun) setRunStatus(status ControlRunStatus) {
	r.stateLock.Lock()
	r.runStatus = status
	r.stateLock.Unlock()

	if r.Finished() {

		// update Progress
		if status == ControlRunError {
			r.executionTree.progress.OnControlError()
		} else {
			r.executionTree.progress.OnControlComplete()
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
