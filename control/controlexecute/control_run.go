package controlexecute

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/grpc"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/statushooks"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
)

const controlQueryTimeout = 240 * time.Second

type ControlRunStatus uint32

const (
	ControlRunReady ControlRunStatus = 1 << iota
	ControlRunStarted
	ControlRunComplete
	ControlRunError
)

// ControlRun is a struct representing the execution of a control run. It will contain one or more result items (i.e. for one or more resources).
type ControlRun struct {
	// the control being run
	Control *modconfig.Control `json:"-"`
	// control summary
	Summary StatusSummary `json:"-"`
	// result rows
	Rows []*ResultRow `json:"results"`
	// a list of distinct dimension keys from the results of this control
	DimensionKeys []string `json:"-"`
	// execution duration
	Duration time.Duration `json:"-"`

	// properties from control
	ControlId   string            `json:"control_id"`
	Description string            `json:"description"`
	Severity    string            `json:"severity"`
	Tags        map[string]string `json:"tags"`
	Title       string            `json:"title"`

	// parent result group
	Group *ResultGroup `json:"-"`
	// execution tree
	Tree *ExecutionTree `json:"-"`
	// used to trace the events within the duration of a control execution
	Lifecycle *utils.LifecycleTimer `json:"-"`

	// the query result stream
	queryResult *queryresult.Result
	rowMap      map[string][]*ResultRow `json:"-"`
	runStatus   ControlRunStatus
	runError    error
	stateLock   sync.Mutex
	doneChan    chan bool
	attempts    int
}

func NewControlRun(control *modconfig.Control, group *ResultGroup, executionTree *ExecutionTree) *ControlRun {
	controlId := control.Name()
	// only show qualified control names for controls from dependent mods
	if control.Mod.Name() == executionTree.workspace.Mod.Name() {
		controlId = control.UnqualifiedName
	}

	res := &ControlRun{
		Control:     control,
		ControlId:   controlId,
		Description: typehelpers.SafeString(control.Description),
		Severity:    typehelpers.SafeString(control.Severity),
		Title:       typehelpers.SafeString(control.Title),
		Tags:        control.GetTags(),
		rowMap:      make(map[string][]*ResultRow),

		Lifecycle: utils.NewLifecycleTimer(),

		Tree:      executionTree,
		runStatus: ControlRunReady,

		Group:    group,
		doneChan: make(chan bool, 1),
	}
	res.Lifecycle.Add("constructed")
	return res
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

// MatchTag returns the value corresponding to the input key. Returns 'false' if not found
func (r *ControlRun) MatchTag(key string, value string) bool {
	val, found := r.Tags[key]
	return found && (val == value)
}

func (r *ControlRun) GetError() error {
	return r.runError
}

func (r *ControlRun) setError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	r.runError = utils.TransformErrorToSteampipe(err)

	// update error count
	r.Summary.Error++
	r.setRunStatus(ctx, ControlRunError)
}

func (r *ControlRun) skip(ctx context.Context) {
	r.setRunStatus(ctx, ControlRunComplete)
}

// set search path for this control run
func (r *ControlRun) setSearchPath(ctx context.Context, session *db_common.DatabaseSession, client db_common.Client) error {
	utils.LogTime("ControlRun.setSearchPath start")
	defer utils.LogTime("ControlRun.setSearchPath end")

	searchPath := []string{}
	searchPathPrefix := []string{}

	if r.Control.SearchPath == nil && r.Control.SearchPathPrefix == nil {
		return nil
	}
	if r.Control.SearchPath != nil {
		searchPath = strings.Split(*r.Control.SearchPath, ",")
	}
	if r.Control.SearchPathPrefix != nil {
		searchPathPrefix = strings.Split(*r.Control.SearchPathPrefix, ",")
	}

	currentPath, err := r.getCurrentSearchPath(ctx, session)
	if err != nil {
		return err
	}

	newSearchPath, err := client.ContructSearchPath(ctx, searchPath, searchPathPrefix, currentPath)
	if err != nil {
		return err
	}

	// no execute the SQL to actuall set the search path
	q := fmt.Sprintf("set search_path to %s", strings.Join(newSearchPath, ","))
	_, err = session.Connection.ExecContext(ctx, q)
	return err
}

func (r *ControlRun) getCurrentSearchPath(ctx context.Context, session *db_common.DatabaseSession) ([]string, error) {
	utils.LogTime("ControlRun.getCurrentSearchPath start")
	defer utils.LogTime("ControlRun.getCurrentSearchPath end")

	row := session.Connection.QueryRowContext(ctx, "show search_path")
	pathAsString := ""
	err := row.Scan(&pathAsString)
	if err != nil {
		return nil, err
	}
	currentSearchPath := strings.Split(pathAsString, ",")
	// unescape search path
	for idx, p := range currentSearchPath {
		p = strings.Join(strings.Split(p, "\""), "")
		p = strings.TrimSpace(p)
		currentSearchPath[idx] = p
	}
	return currentSearchPath, nil
}

func (r *ControlRun) execute(ctx context.Context, client db_common.Client) {
	log.Printf("[TRACE] begin ControlRun.Start: %s\n", r.Control.Name())
	defer log.Printf("[TRACE] end ControlRun.Start: %s\n", r.Control.Name())

	r.Lifecycle.Add("execute_start")

	control := r.Control
	log.Printf("[TRACE] control start, %s\n", control.Name())

	startTime := time.Now()

	// function to cleanup and update status after control run completion
	defer func() {
		// update the result group status with our status - this will be passed all the way up the execution tree
		r.Group.updateSummary(r.Summary)
		if len(r.Severity) != 0 {
			r.Group.updateSeverityCounts(r.Severity, r.Summary)
		}
		r.Lifecycle.Add("execute_end")
		r.Duration = time.Since(startTime)
		if r.Group != nil {
			r.Group.addDuration(r.Duration)
		}
		log.Printf("[TRACE] finishing with concurrency, %s, , %d\n", r.Control.Name(), r.Tree.Progress.Executing)
	}()

	// get a db connection
	r.Lifecycle.Add("queued_for_session")
	sessionResult := client.AcquireSession(ctx)
	if sessionResult.Error != nil {
		if !utils.IsCancelledError(sessionResult.Error) {
			sessionResult.Error = fmt.Errorf("error acquiring database connection, %s", sessionResult.Error.Error())
		}
		return
	}
	r.Lifecycle.Add("got_session")
	dbSession := sessionResult.Session
	defer func() {
		// do this in a closure, otherwise the argument will not get evaluated during calltime
		dbSession.Close(utils.IsContextCancelled(ctx))
	}()

	// set our status
	r.runStatus = ControlRunStarted

	// update the current running control in the Progress renderer
	r.Tree.Progress.OnControlStart(ctx, control)
	defer r.Tree.Progress.OnControlFinish(ctx)

	// resolve the control query
	r.Lifecycle.Add("query_resolution_start")
	query, err := r.resolveControlQuery(control)
	if err != nil {
		r.setError(ctx, err)
		return
	}
	r.Lifecycle.Add("query_resolution_finish")

	log.Printf("[TRACE] setting search path %s\n", control.Name())
	r.Lifecycle.Add("set_search_path_start")
	if err := r.setSearchPath(ctx, dbSession, client); err != nil {
		r.setError(ctx, err)
		return
	}
	r.Lifecycle.Add("set_search_path_finish")

	// get a context with a timeout for the control to execute within
	// we don't use the cancelFn from this timeout context, since usage will lead to 'pgx'
	// prematurely closing the database connection that this query executed in
	controlExecutionCtx := r.getControlQueryContext(ctx)

	// execute the control query
	// NOTE no need to pass an OnComplete callback - we are already closing our session after waiting for results
	log.Printf("[TRACE] execute start for, %s\n", control.Name())
	r.Lifecycle.Add("query_start")
	queryResult, err := client.ExecuteInSession(controlExecutionCtx, dbSession, query, nil)
	r.Lifecycle.Add("query_finish")
	log.Printf("[TRACE] execute finish for, %s\n", control.Name())

	if err != nil {
		r.attempts++

		// is this an rpc EOF error - meaning that the plugin somehow crashed
		if grpc.IsGRPCConnectivityError(err) {
			if r.attempts < constants.MaxControlRunAttempts {
				log.Printf("[TRACE] control %s query failed with plugin connectivity error %s - retrying...", r.Control.Name(), err)
				// recurse into this function to retry using the original context - which Execute will use to create it's own timeout context
				r.execute(ctx, client)
				return
			} else {
				log.Printf("[TRACE] control %s query failed again with plugin connectivity error %s - NOT retrying...", r.Control.Name(), err)
			}
		}
		r.setError(ctx, err)
		return
	}

	r.queryResult = queryResult

	// now wait for control completion
	log.Printf("[TRACE] wait result for, %s\n", control.Name())
	r.waitForResults(ctx)
	log.Printf("[TRACE] finish result for, %s\n", control.Name())
}

// create a context with a deadline, and with status updates disabled (we do not want to show 'loading' results)
func (r *ControlRun) getControlQueryContext(ctx context.Context) context.Context {
	// create a context with a deadline
	shouldBeDoneBy := time.Now().Add(controlQueryTimeout)
	// we don't use this cancel fn because, pgx prematurely cancels the PG connection when this cancel gets called in 'defer'
	newCtx, _ := context.WithDeadline(ctx, shouldBeDoneBy)

	// disable the status spinner to hide 'loading' results)
	newCtx = statushooks.DisableStatusHooks(newCtx)

	return newCtx
}

func (r *ControlRun) resolveControlQuery(control *modconfig.Control) (string, error) {
	query, err := r.Tree.workspace.ResolveControlQuery(control, nil)
	if err != nil {
		return "", fmt.Errorf(`cannot run %s - failed to resolve query "%s": %s`, control.Name(), typehelpers.SafeString(control.SQL), err.Error())
	}
	if query == "" {
		return "", fmt.Errorf(`cannot run %s - failed to resolve query "%s"`, control.Name(), typehelpers.SafeString(control.SQL))
	}
	return query, nil
}

func (r *ControlRun) waitForResults(ctx context.Context) {
	// create a channel to which will be closed when gathering has been done
	gatherDoneChan := make(chan string)
	go func() {
		r.gatherResults(ctx)
		close(gatherDoneChan)
	}()

	select {
	// check for cancellation
	case <-ctx.Done():
		r.setError(ctx, ctx.Err())
	case <-gatherDoneChan:
		// do nothing
	}
}

func (r *ControlRun) gatherResults(ctx context.Context) {
	r.Lifecycle.Add("gather_start")
	defer func() { r.Lifecycle.Add("gather_finish") }()

	defer func() {
		for _, row := range r.Rows {
			for _, dim := range row.Dimensions {
				r.DimensionKeys = append(r.DimensionKeys, dim.Key)
			}
		}
		r.DimensionKeys = utils.StringSliceDistinct(r.DimensionKeys)
		r.Group.addDimensionKeys(r.DimensionKeys...)
	}()

	for {
		select {
		case row := <-*r.queryResult.RowChan:
			// nil row means control run is complete
			if row == nil {
				// nil row means we are done
				r.setRunStatus(ctx, ControlRunComplete)
				r.createdOrderedResultRows()
				return
			}
			// if the row is in error then we terminate the run
			if row.Error != nil {
				// set error status and summary
				r.setError(ctx, row.Error)
				// update the result group status with our status - this will be passed all the way up the execution tree
				r.Group.updateSummary(r.Summary)
				return
			}

			// so all is ok - create another result row
			result, err := NewResultRow(r, row, r.queryResult.ColTypes)
			if err != nil {
				r.setError(ctx, err)
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
	r.rowMap[row.Status] = append(r.rowMap[row.Status], row)

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
		r.Rows = append(r.Rows, r.rowMap[status]...)
	}
}

func (r *ControlRun) setRunStatus(ctx context.Context, status ControlRunStatus) {
	r.stateLock.Lock()
	r.runStatus = status
	r.stateLock.Unlock()

	if r.Finished() {
		// update Progress
		if status == ControlRunError {
			r.Tree.Progress.OnControlError(ctx)
		} else {
			r.Tree.Progress.OnControlComplete(ctx)
		}

		r.doneChan <- true
	}
}
