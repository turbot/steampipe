package controlexecute

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/turbot/steampipe/db/db_common"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/query/queryresult"
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

// ControlRun is a struct representing a  a control run - will contain one or more result items (i.e. for one or more resources)
type ControlRun struct {
	runError error `json:"-"`

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

// set search path for this control run
func (r *ControlRun) setSearchPath(ctx context.Context, session *sql.Conn, client db_common.Client) error {
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

	newSearchPath, err := client.ContructSearchPath(searchPath, searchPathPrefix, currentPath)
	if err != nil {
		return err
	}

	// no execute the SQL to actuall set the search path
	q := fmt.Sprintf("set search_path to %s", strings.Join(newSearchPath, ","))
	_, err = session.ExecContext(ctx, q)
	return err
}

func (r *ControlRun) getCurrentSearchPath(ctx context.Context, session *sql.Conn) ([]string, error) {
	utils.LogTime("ControlRun.getCurrentSearchPath start")
	defer utils.LogTime("ControlRun.getCurrentSearchPath end")

	row := session.QueryRowContext(ctx, "show search_path")
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

func (r *ControlRun) Execute(ctx context.Context, client db_common.Client) {
	log.Printf("[TRACE] begin ControlRun.Start: %s\n", r.Control.Name())
	defer log.Printf("[TRACE] end ControlRun.Start: %s\n", r.Control.Name())

	startTime := time.Now()
	control := r.Control
	log.Printf("[TRACE] control start %s\n", control.Name())
	r.executionTree.progress.OnControlExecuteStart()

	// function to cleanup and update status after control run completion
	defer func() {
		r.executionTree.progress.OnControlExecuteFinish()
		// update the result group status with our status - this will be passed all the way up the execution tree
		r.group.updateSummary(r.Summary)
		if len(r.Severity) != 0 {
			r.group.updateSeverityCounts(r.Severity, r.Summary)
		}
		r.Duration = time.Since(startTime)
	}()

	// get a db connection
	dbSession, err := client.AcquireSession(ctx)
	log.Printf("[TRACE] got session for: %s\n", r.Control.Name())
	if err != nil {
		r.SetError(fmt.Errorf("error acquiring database connection: %s", err.Error()))
		return
	}
	defer dbSession.Close()

	// set our status
	r.runStatus = ControlRunStarted

	// update the current running control in the Progress renderer
	r.executionTree.progress.OnControlStart(control)

	// resolve the control query
	query, err := r.resolveControlQuery(err, control)
	if err != nil {
		r.SetError(err)
		return
	}

	log.Printf("[TRACE] setting search path %s\n", control.Name())
	if err := r.setSearchPath(ctx, dbSession, client); err != nil {
		r.SetError(err)
		return
	}

	// create a context with a deadline
	shouldBeDoneBy := time.Now().Add(controlQueryTimeout)
	ctx, cancel := context.WithDeadline(ctx, shouldBeDoneBy)
	// Even though ctx will expire, it is good practice to call its cancellation function in any case.
	defer cancel()

	// execute the control query
	queryResult, err := client.ExecuteInSession(ctx, dbSession, query, false)
	if err != nil {
		log.Printf("[TRACE] client.ExecuteInSession returned error %s", err.Error())
		r.SetError(err)
		return
	}
	r.queryResult = queryResult

	// now wait for control completion
	r.waitForResults(ctx)
}

func (r *ControlRun) resolveControlQuery(err error, control *modconfig.Control) (string, error) {
	query, err := r.executionTree.workspace.ResolveControlQuery(control)
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
		r.gatherResults()
		close(gatherDoneChan)
	}()

	select {
	// check for cancellation
	case <-ctx.Done():
		r.SetError(ctx.Err())
	case <-gatherDoneChan:
		// do nothing
	}
}

func (r *ControlRun) gatherResults() {
	for {
		select {
		case row := <-*r.queryResult.RowChan:
			// nil row means control run is complete
			if row == nil {
				// nil row means we are done
				r.setRunStatus(ControlRunComplete)
				r.createdOrderedResultRows()
				return
			}
			// if the row is in error then we terminate the run
			if row.Error != nil {
				// set error status and summary
				r.SetError(row.Error)
				// update the result group status with our status - this will be passed all the way up the execution tree
				r.group.updateSummary(r.Summary)
				return
			}

			// so all is ok - create another result row
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
	if err == nil {
		return
	}
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
