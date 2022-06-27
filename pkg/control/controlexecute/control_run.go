package controlexecute

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v3/grpc"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/control/controlstatus"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/query/queryresult"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
)

const controlQueryTimeout = 240 * time.Second

// ControlRun is a struct representing the execution of a control run. It will contain one or more result items (i.e. for one or more resources).
type ControlRun struct {
	// properties from control
	ControlId     string            `json:"name"`
	Title         string            `json:"title,omitempty"`
	Description   string            `json:"description,omitempty"`
	Documentation string            `json:"documentation,omitempty"`
	Tags          map[string]string `json:"tags,omitempty"`
	Display       string            `json:"display,omitempty"`
	Type          string            `json:"display_type,omitempty"`

	// this will be serialised under 'properties'
	Severity string `json:"-"`

	// "control"
	NodeType string `json:"panel_type"`

	// the control being run
	Control *modconfig.Control `json:"properties,omitempty"`
	// control summary
	Summary   *controlstatus.StatusSummary   `json:"summary"`
	RunStatus controlstatus.ControlRunStatus `json:"status"`
	// result rows
	Rows ResultRows `json:"-"`

	// the results in snapshot format
	Data *dashboardtypes.LeafData `json:"data"`

	// a list of distinct dimension keys from the results of this control
	DimensionKeys []string `json:"-"`

	// execution duration
	Duration time.Duration `json:"-"`
	// parent result group
	Group *ResultGroup `json:"-"`
	// execution tree
	Tree *ExecutionTree `json:"-"`
	// used to trace the events within the duration of a control execution
	Lifecycle *utils.LifecycleTimer `json:"-"`
	// save run error as string for JSON export
	RunErrorString string `json:"error,omitempty"`
	runError       error
	// the query result stream
	queryResult *queryresult.Result
	rowMap      map[string]ResultRows
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
		Control:       control,
		ControlId:     controlId,
		Description:   control.GetDescription(),
		Documentation: control.GetDocumentation(),
		Tags:          control.GetTags(),
		Display:       control.GetDisplay(),
		Type:          control.GetType(),

		Severity:  typehelpers.SafeString(control.Severity),
		Title:     typehelpers.SafeString(control.Title),
		rowMap:    make(map[string]ResultRows),
		Summary:   &controlstatus.StatusSummary{},
		Lifecycle: utils.NewLifecycleTimer(),

		Tree:      executionTree,
		RunStatus: controlstatus.ControlRunReady,

		Group:    group,
		NodeType: modconfig.BlockTypeControl,
		doneChan: make(chan bool, 1),
	}
	res.Lifecycle.Add("constructed")
	return res
}

// GetControlId implements ControlRunStatusProvider
func (r *ControlRun) GetControlId() string {
	r.stateLock.Lock()
	defer r.stateLock.Unlock()
	return r.ControlId
}

// GetRunStatus implements ControlRunStatusProvider
func (r *ControlRun) GetRunStatus() controlstatus.ControlRunStatus {
	r.stateLock.Lock()
	defer r.stateLock.Unlock()
	return r.RunStatus
}

// GetStatusSummary implements ControlRunStatusProvider
func (r *ControlRun) GetStatusSummary() *controlstatus.StatusSummary {
	r.stateLock.Lock()
	defer r.stateLock.Unlock()
	return r.Summary
}

func (r *ControlRun) Finished() bool {
	status := r.GetRunStatus()
	return status == controlstatus.ControlRunComplete || status == controlstatus.ControlRunError
}

// MatchTag returns the value corresponding to the input key. Returns 'false' if not found
func (r *ControlRun) MatchTag(key string, value string) bool {
	val, found := r.Control.GetTags()[key]
	return found && (val == value)
}

func (r *ControlRun) GetError() error {
	return r.runError
}

// IsSnapshotPanel implements SnapshotPanel
func (*ControlRun) IsSnapshotPanel() {}

// IsExecutionTreeNode implements ExecutionTreeNode
func (*ControlRun) IsExecutionTreeNode() {}

// GetChildren implements ExecutionTreeNode
func (*ControlRun) GetChildren() []ExecutionTreeNode { return nil }

// GetName implements ExecutionTreeNode
func (r *ControlRun) GetName() string { return r.ControlId }

// AsTreeNode implements ExecutionTreeNode
func (r *ControlRun) AsTreeNode() *dashboardtypes.SnapshotTreeNode {
	res := &dashboardtypes.SnapshotTreeNode{
		Name:     r.ControlId,
		NodeType: r.NodeType,
	}
	return res
}

func (r *ControlRun) setError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	if r.runError == context.DeadlineExceeded {
		r.runError = fmt.Errorf("control execution timed out")
	} else {
		r.runError = utils.TransformErrorToSteampipe(err)
	}
	r.RunErrorString = r.runError.Error()
	// update error count
	r.Summary.Error++
	r.setRunStatus(ctx, controlstatus.ControlRunError)
}

func (r *ControlRun) skip(ctx context.Context) {
	r.setRunStatus(ctx, controlstatus.ControlRunComplete)
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

	newSearchPath, err := client.ContructSearchPath(ctx, searchPath, searchPathPrefix)
	if err != nil {
		return err
	}

	// no execute the SQL to actuall set the search path
	q := fmt.Sprintf("set search_path to %s", strings.Join(newSearchPath, ","))
	_, err = session.Connection.ExecContext(ctx, q)
	return err
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
	sessionResult := r.acquireSession(ctx, client)
	if sessionResult.Error != nil {
		if !utils.IsCancelledError(sessionResult.Error) {
			log.Printf("[TRACE] controlRun %s execute failed to acquire session: %s", r.ControlId, sessionResult.Error)
			sessionResult.Error = fmt.Errorf("error acquiring database connection, %s", sessionResult.Error.Error())
			r.setError(ctx, sessionResult.Error)
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
	r.RunStatus = controlstatus.ControlRunStarted

	// update the current running control in the Progress renderer
	r.Tree.Progress.OnControlStart(ctx, r)
	defer func() {
		// update Progress
		if r.GetRunStatus() == controlstatus.ControlRunError {
			r.Tree.Progress.OnControlError(ctx, r)
		} else {
			r.Tree.Progress.OnControlComplete(ctx, r)
		}
	}()

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

// try to acquire a database session - retry up to 4 times if there is an error
func (r *ControlRun) acquireSession(ctx context.Context, client db_common.Client) *db_common.AcquireSessionResult {
	var sessionResult *db_common.AcquireSessionResult
	for attempt := 0; attempt < 4; attempt++ {
		sessionResult = client.AcquireSession(ctx)
		if sessionResult.Error == nil || utils.IsCancelledError(sessionResult.Error) {
			break
		}

		log.Printf("[TRACE] controlRun %s acquireSession failed with error: %s - retrying", r.ControlId, sessionResult.Error)
	}

	return sessionResult
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
	resolvedQuery, err := r.Tree.workspace.ResolveQueryFromQueryProvider(control, nil)
	if err != nil {
		return "", fmt.Errorf(`cannot run %s - failed to resolve query "%s": %s`, control.Name(), typehelpers.SafeString(control.SQL), err.Error())
	}
	if resolvedQuery.ExecuteSQL == "" {
		return "", fmt.Errorf(`cannot run %s - failed to resolve query "%s"`, control.Name(), typehelpers.SafeString(control.SQL))
	}
	return resolvedQuery.ExecuteSQL, nil
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
		dimensionsSchema := r.getDimensionSchema()
		// convert the data to snapshot format
		r.Data = r.Rows.ToLeafData(dimensionsSchema)
	}()

	for {
		select {
		case row := <-*r.queryResult.RowChan:
			// nil row means control run is complete
			if row == nil {
				// nil row means we are done
				r.setRunStatus(ctx, controlstatus.ControlRunComplete)
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

func (r *ControlRun) getDimensionSchema() map[string]*dashboardtypes.ColumnSchema {
	var dimensionsSchema = make(map[string]*dashboardtypes.ColumnSchema)

	for _, row := range r.Rows {
		for _, dim := range row.Dimensions {
			if _, ok := dimensionsSchema[dim.Key]; !ok {
				// add to map
				dimensionsSchema[dim.Key] = &dashboardtypes.ColumnSchema{
					Name:     dim.Key,
					DataType: dim.SqlType,
				}
				// also add to DimensionKeys
				r.DimensionKeys = append(r.DimensionKeys, dim.Key)
			}
		}
	}
	// add keys to group
	r.Group.addDimensionKeys(r.DimensionKeys...)
	return dimensionsSchema
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

func (r *ControlRun) setRunStatus(ctx context.Context, status controlstatus.ControlRunStatus) {
	r.stateLock.Lock()
	r.RunStatus = status
	r.stateLock.Unlock()

	if r.Finished() {
		r.doneChan <- true
	}
}
