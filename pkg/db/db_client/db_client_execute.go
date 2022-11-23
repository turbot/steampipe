package db_client

import (
	"context"
	"database/sql"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/query/queryresult"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/utils"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// ExecuteSync implements Client
// execute a query against this client and wait for the result
func (c *DbClient) ExecuteSync(ctx context.Context, query string, args ...any) (*queryresult.SyncQueryResult, error) {
	// acquire a session
	sessionResult := c.AcquireSession(ctx)
	if sessionResult.Error != nil {
		return nil, sessionResult.Error
	}

	// set setShouldShowTiming flag
	// (this will refetch ScanMetadataMaxId if timing has just been enabled)
	c.setShouldShowTiming(ctx, sessionResult.Session)

	defer func() {
		// we need to do this in a closure, otherwise the ctx will be evaluated immediately
		// and not in call-time
		sessionResult.Session.Close(utils.IsContextCancelled(ctx))
	}()
	return c.ExecuteSyncInSession(ctx, sessionResult.Session, query, args...)
}

// ExecuteSyncInSession implements Client
// execute a query against this client and wait for the result
func (c *DbClient) ExecuteSyncInSession(ctx context.Context, session *db_common.DatabaseSession, query string, args ...any) (*queryresult.SyncQueryResult, error) {
	if query == "" {
		return &queryresult.SyncQueryResult{}, nil
	}

	result, err := c.ExecuteInSession(ctx, session, nil, query, args...)
	if err != nil {
		return nil, error_helpers.WrapError(err)
	}

	syncResult := &queryresult.SyncQueryResult{Cols: result.Cols}
	for row := range *result.RowChan {
		select {
		case <-ctx.Done():
		default:
			// save the first row error to return
			if row.Error != nil && err == nil {
				err = error_helpers.WrapError(row.Error)
			}
			syncResult.Rows = append(syncResult.Rows, row)
		}
	}
	if c.shouldShowTiming() {
		syncResult.TimingResult = <-result.TimingResult
	}

	return syncResult, err
}

// Execute implements Client
// execute the query in the given Context
// NOTE: The returned Result MUST be fully read - otherwise the connection will block and will prevent further communication
func (c *DbClient) Execute(ctx context.Context, query string, args ...any) (*queryresult.Result, error) {
	// acquire a session
	sessionResult := c.AcquireSession(ctx)
	if sessionResult.Error != nil {
		return nil, sessionResult.Error
	}
	// re-read ArgTiming from viper (in case the .timing command has been run)
	// (this will refetch ScanMetadataMaxId if timing has just been enabled)
	c.setShouldShowTiming(ctx, sessionResult.Session)

	// define callback to close session when the async execution is complete
	closeSessionCallback := func() { sessionResult.Session.Close(utils.IsContextCancelled(ctx)) }
	return c.ExecuteInSession(ctx, sessionResult.Session, closeSessionCallback, query, args...)
}

// ExecuteInSession implements Client
// execute the query in the given Context using the provided DatabaseSession
// ExecuteInSession assumes no responsibility over the lifecycle of the DatabaseSession - that is the responsibility of the caller
// NOTE: The returned Result MUST be fully read - otherwise the connection will block and will prevent further communication
func (c *DbClient) ExecuteInSession(ctx context.Context, session *db_common.DatabaseSession, onComplete func(), query string, args ...any) (res *queryresult.Result, err error) {
	if query == "" {
		return queryresult.NewResult(nil), nil
	}

	// fail-safes
	if session == nil {
		return nil, fmt.Errorf("nil session passed to ExecuteInSession")
	}
	if session.Connection == nil {
		return nil, fmt.Errorf("nil database connection passed to ExecuteInSession")
	}
	startTime := time.Now()
	// get a context with a timeout for the query to execute within
	// we don't use the cancelFn from this timeout context, since usage will lead to 'pgx'
	// prematurely closing the database connection that this query executed in
	ctxExecute := c.getExecuteContext(ctx)

	var tx *sql.Tx

	defer func() {
		if err != nil {
			err = error_helpers.HandleQueryTimeoutError(err)
			// stop spinner in case of error
			statushooks.Done(ctxExecute)
			// error - rollback transaction if we have one
			if tx != nil {
				tx.Rollback()
			}
			// in case of error call the onComplete callback
			if onComplete != nil {
				onComplete()
			}
		}
	}()

	statushooks.SetStatus(ctxExecute, "Loading results...")

	// start query
	var rows pgx.Rows
	rows, err = c.startQuery(ctxExecute, session.Connection, query, args...)
	if err != nil {
		return
	}

	colDefs := fieldDescriptionsToColumns(rows.FieldDescriptions(), session.Connection.Conn())

	result := queryresult.NewResult(colDefs)

	// read the rows in a go routine
	go func() {
		// define a callback which fetches the timing information
		// this will be invoked after reading rows is complete but BEFORE closing the rows object (which closes the connection)
		timingCallback := func() {
			c.getQueryTiming(ctxExecute, startTime, session, result.TimingResult)
		}

		// read in the rows and stream to the query result object
		c.readRows(ctxExecute, rows, result, timingCallback)

		// call the completion callback - if one was provided
		if onComplete != nil {
			onComplete()
		}
	}()

	return result, nil
}

func (c *DbClient) getExecuteContext(ctx context.Context) context.Context {
	queryTimeout := time.Duration(viper.GetInt(constants.ArgDatabaseQueryTimeout)) * time.Second
	// if timeout is zero, do not set a timeout
	if queryTimeout == 0 {
		return ctx
	}
	// create a context with a deadline
	shouldBeDoneBy := time.Now().Add(queryTimeout)
	// we don't use this cancel fn because, pgx prematurely cancels the PG connection when this cancel gets called in 'defer'
	newCtx, _ := context.WithDeadline(ctx, shouldBeDoneBy)

	return newCtx
}

func (c *DbClient) getQueryTiming(ctx context.Context, startTime time.Time, session *db_common.DatabaseSession, resultChannel chan *queryresult.TimingResult) {
	if !c.shouldShowTiming() {
		return
	}

	var timingResult = &queryresult.TimingResult{
		Duration: time.Since(startTime),
	}
	// disable fetching timing information to avoid recursion
	c.disableTiming = true

	// whatever happens, we need to reenable timing, and send the result back with at least the duration
	defer func() {
		c.disableTiming = false
		resultChannel <- timingResult
	}()

	res, err := c.ExecuteSyncInSession(ctx, session, fmt.Sprintf("select id, rows_fetched, cache_hit, hydrate_calls from steampipe_command.scan_metadata where id > %d", session.ScanMetadataMaxId))
	// if we failed to read scan metadata (either because the query failed or the plugin does not support it)
	// just return
	if err != nil || len(res.Rows) == 0 {
		return
	}

	// so we have scan metadata - create the metadata struct
	timingResult.Metadata = &queryresult.TimingMetadata{}
	var id int64
	for _, r := range res.Rows {
		rw := r.(*queryresult.RowResult)
		id = rw.Data[0].(int64)
		rowsFetched := rw.Data[1].(int64)
		cacheHit := rw.Data[2].(bool)
		hydrateCalls := rw.Data[3].(int64)

		timingResult.Metadata.HydrateCalls += hydrateCalls
		if cacheHit {
			timingResult.Metadata.CachedRowsFetched += rowsFetched
		} else {
			timingResult.Metadata.RowsFetched += rowsFetched
		}

	}
	// update the max id for this session
	session.ScanMetadataMaxId = id

	return
}

func (c *DbClient) updateScanMetadataMaxId(ctx context.Context, session *db_common.DatabaseSession) error {
	res, err := c.ExecuteSyncInSession(ctx, session, "select max(id) from steampipe_command.scan_metadata")
	if err != nil {
		return err
	}

	for _, r := range res.Rows {
		rw := r.(*queryresult.RowResult)
		id, ok := rw.Data[0].(int64)
		if ok {
			// update the max id for this session
			session.ScanMetadataMaxId = id
		}

		break
	}
	return nil
}

// run query in a goroutine, so we can check for cancellation
// in case the client becomes unresponsive and does not respect context cancellation
func (c *DbClient) startQuery(ctx context.Context, conn *pgxpool.Conn, query string, args ...any) (rows pgx.Rows, err error) {
	doneChan := make(chan bool)

	go func() {
		// start asynchronous query
		rows, err = conn.Query(ctx, query, args...)
		close(doneChan)
	}()

	select {
	case <-doneChan:
	case <-ctx.Done():
		err = ctx.Err()
	}
	return
}

func (c *DbClient) readRows(ctx context.Context, rows pgx.Rows, result *queryresult.Result, timingCallback func()) {
	// defer this, so that these get cleaned up even if there is an unforeseen error
	defer func() {
		// we are done fetching results. time for display. clear the status indication
		statushooks.Done(ctx)
		// call the timing callback BEFORE closing the rows
		timingCallback()
		// close the sql rows object
		rows.Close()
		if err := rows.Err(); err != nil {
			result.StreamError(err)
		}
		// close the channels in the result object
		result.Close()

	}()

	rowCount := 0
Loop:
	for rows.Next() {
		select {
		case <-ctx.Done():
			statushooks.SetStatus(ctx, "Cancelling query")
			break Loop
		default:
			rowResult, err := readRow(rows, result.Cols)
			if err != nil {
				// the error will be streamed in the defer
				break Loop
			}

			// TACTICAL
			// determine whether to stop the spinner as soon as we stream a row or to wait for completion
			if isStreamingOutput() {
				statushooks.Done(ctx)
			}

			result.StreamRow(rowResult)

			// update the status message with the count of rows that have already been fetched
			// this will not show if the spinner is not active
			statushooks.SetStatus(ctx, fmt.Sprintf("Loading results: %3s", humanizeRowCount(rowCount)))
			rowCount++
		}
	}
}

func readRow(rows pgx.Rows, cols []*queryresult.ColumnDef) ([]interface{}, error) {
	columnValues, err := rows.Values()
	if err != nil {
		return nil, error_helpers.WrapError(err)
	}
	return populateRow(columnValues, cols)
}

func populateRow(columnValues []interface{}, cols []*queryresult.ColumnDef) ([]interface{}, error) {
	result := make([]interface{}, len(columnValues))
	for i, columnValue := range columnValues {
		if columnValue != nil {
			result[i] = columnValue
			switch cols[i].DataType {
			case "_TEXT":
				if arr, ok := columnValue.([]interface{}); ok {
					elements := utils.Map(arr, func(e interface{}) string { return e.(string) })
					result[i] = strings.Join(elements, ",")
				}
			case "INET":
				if inet, ok := columnValue.(netip.Prefix); ok {
					result[i] = strings.TrimSuffix(inet.String(), "/32")
				}
			case "UUID":
				if bytes, ok := columnValue.([16]uint8); ok {
					if u, err := uuid.FromBytes(bytes[:]); err == nil {
						result[i] = u
					}
				}
			case "TIME":
				if t, ok := columnValue.(pgtype.Time); ok {
					result[i] = time.UnixMicro(t.Microseconds).UTC().Format("15:04:05")
				}
			case "INTERVAL":
				if interval, ok := columnValue.(pgtype.Interval); ok {
					var sb strings.Builder
					years := interval.Months / 12
					months := interval.Months % 12
					if years > 0 {
						sb.WriteString(fmt.Sprintf("%d %s ", years, utils.Pluralize("year", int(years))))
					}
					if months > 0 {
						sb.WriteString(fmt.Sprintf("%d %s ", months, utils.Pluralize("mon", int(months))))
					}
					if interval.Days > 0 {
						sb.WriteString(fmt.Sprintf("%d %s ", interval.Days, utils.Pluralize("day", int(interval.Days))))
					}
					if interval.Microseconds > 0 {
						d := time.Duration(interval.Microseconds) * time.Microsecond
						formatStr := time.Unix(0, 0).UTC().Add(d).Format("15:04:05")
						sb.WriteString(formatStr)
					}
					result[i] = sb.String()
				}

			case "NUMERIC":
				if numeric, ok := columnValue.(pgtype.Numeric); ok {
					if f, err := numeric.Float64Value(); err == nil {
						result[i] = f.Float64
					}
				}
			}
		}
	}
	return result, nil
}

func isStreamingOutput() bool {
	outputFormat := viper.GetString(constants.ArgOutput)

	return helpers.StringSliceContains([]string{constants.OutputFormatCSV, constants.OutputFormatLine}, outputFormat)
}

func humanizeRowCount(count int) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d", count)
}
