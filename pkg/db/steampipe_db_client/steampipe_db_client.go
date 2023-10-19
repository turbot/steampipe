package steampipe_db_client

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/db_client"
	"github.com/turbot/pipe-fittings/db_common"
	"github.com/turbot/pipe-fittings/queryresult"
)

type DbConnectionCallback func(context.Context, *pgx.Conn) error

// DbClient wraps over `sql.DB` and gives an interface to the database
type SteampipeDbClient struct {
	db_client.DbClient
	onConnectionCallback DbConnectionCallback
	// the settings of the server that this client is connected to
	// a cached copy of (viper.GetBool(constants.ArgTiming) && viper.GetString(constants.ArgOutput) == constants.OutputFormatTable)
	// (cached to avoid concurrent access error on viper)
	showTimingFlag bool
	// disable timing - set whilst in process of querying the timing
	disableTiming bool

	// map of database sessions, keyed to the backend_pid in postgres
	// used to update session search path where necessary
	// TODO: there's no code which cleans up this map when connections get dropped by pgx
	// https://github.com/turbot/steampipe/issues/3737
	sessions map[uint32]*db_common.DatabaseSession

	// allows locked access to the 'sessions' map
	sessionsMutex *sync.Mutex

	// if a custom search path or a prefix is used, store it here
	customSearchPath []string
	searchPathPrefix []string
	// the default user search path
	userSearchPath []string
}

func NewSteampipeDbClient(ctx context.Context, connectionString string, onConnectionCallback DbConnectionCallback, opts ...db_client.ClientOption) (steampipeClient *SteampipeDbClient, err error) {
	dbClient, err := db_client.NewDbClient(ctx, connectionString)
	if err != nil {
		return nil, err
	}

	steampipeClient := &SteampipeDbClient{
		DbClient:             *dbClient,
		onConnectionCallback: onConnectionCallback,
		sessions:             make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex:        &sync.Mutex{},
	}

	// set pre execute hook to re-read ArgTiming from viper (in case the .timing command has been run)
	// (this will refetch ScanMetadataMaxId if timing has just been enabled)
	dbClient.BeforeExecuteHook = steampipeClient.setShouldShowTiming

	// wrap onConnectionCallback to use wait group
	var wrappedOnConnectionCallback DbConnectionCallback
	wg := &sync.WaitGroup{}
	if onConnectionCallback != nil {
		wrappedOnConnectionCallback = func(ctx context.Context, conn *pgx.Conn) error {
			wg.Add(1)
			defer wg.Done()
			return onConnectionCallback(ctx, conn)
		}
	}
	steampipeClient.onConnectionCallback = wrappedOnConnectionCallback

	// set user search path
	if err := steampipeClient.LoadUserSearchPath(ctx); err != nil {
		return nil, err
	}

	// populate customSearchPath
	if err := steampipeClient.SetRequiredSessionSearchPath(ctx); err != nil {
		return nil, err
	}

	return steampipeClient, nil
}

// TODO session keying mechanism - re-add session map
// ScanMetadataMaxId broken

// TODO isLocalService
//
//config, err := pgxpool.ParseConfig(c.connectionString)
//if err != nil {
//return err
//}
//
//locals := []string{
//"127.0.0.1",
//"::1",
//"localhost",
//}
//
//// when connected to a service which is running a plugin compiled with SDK pre-v5, the plugin
//// will not have the ability to turn off caching (feature introduced in SDKv5)
////
//// the 'isLocalService' is used to set the client end cache to 'false' if caching is turned off in the local service
////
//// this is a temporary workaround to make sure
//// that we can turn off caching for plugins compiled with SDK pre-V5
//// worst case scenario is that we don't switch off the cache for pre-V5 plugins
//// refer to: https://github.com/turbot/steampipe/blob/f7f983a552a07e50e526fcadf2ccbfdb7b247cc0/pkg/db/db_client/db_client_session.go#L66
//if helpers.StringSliceContains(locals, config.ConnConfig.Host) {
//c.isLocalService = true
//}

func (c *SteampipeDbClient) setShouldShowTiming(ctx context.Context, session *db_common.DatabaseSession) error {
	currentShowTimingFlag := viper.GetBool(constants.ArgTiming)

	// if we are turning timing ON, fetch the ScanMetadataMaxId
	// to ensure we only select the relevant scan metadata table entries
	if currentShowTimingFlag && !c.showTimingFlag {
		c.updateScanMetadataMaxId(ctx, session)
	}

	c.showTimingFlag = currentShowTimingFlag
	return nil
}

func (c *SteampipeDbClient) shouldShowTiming() bool {
	return c.showTimingFlag && !c.disableTiming
}

func (c *SteampipeDbClient) getQueryTiming(ctx context.Context, startTime time.Time, session *db_common.DatabaseSession, resultChannel chan *queryresult.TimingResult) {
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

	var scanRows *ScanMetadataRow
	err := db_common.ExecuteSystemClientCall(ctx, session.Connection, func(ctx context.Context, tx *sql.Tx) error {
		query := fmt.Sprintf("select id, rows_fetched, cache_hit, hydrate_calls from %s.%s where id > %d", constants.InternalSchema, constants.ForeignTableScanMetadata, session.ScanMetadataMaxId)
		rows, err := tx.QueryContext(ctx, query)
		if err != nil {
			return err
		}
		scanRows, err = db_common.CollectOneToStructByName[ScanMetadataRow](rows)
		return err
	})

	// if we failed to read scan metadata (either because the query failed or the plugin does not support it) just return
	// we don't return the error, since we don't want to error out in this case
	if err != nil || scanRows == nil {
		return
	}

	// so we have scan metadata - create the metadata struct
	timingResult.Metadata = &queryresult.TimingMetadata{}
	timingResult.Metadata.HydrateCalls += scanRows.HydrateCalls
	if scanRows.CacheHit {
		timingResult.Metadata.CachedRowsFetched += scanRows.RowsFetched
	} else {
		timingResult.Metadata.RowsFetched += scanRows.RowsFetched
	}
	// update the max id for this session
	session.ScanMetadataMaxId = scanRows.Id
}

func (c *SteampipeDbClient) updateScanMetadataMaxId(ctx context.Context, session *db_common.DatabaseSession) error {
	return db_common.ExecuteSystemClientCall(ctx, session.Connection, func(ctx context.Context, tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx, fmt.Sprintf("select max(id) from %s.%s", constants.InternalSchema, constants.ForeignTableScanMetadata))
		err := row.Scan(&session.ScanMetadataMaxId)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	})
}

func (c *SteampipeDbClient) Close(ctx context.Context) error {
	// nullify active sessions, since with the closing of the pools
	// none of the sessions will be valid anymore
	c.sessions = nil

	c.DbClient.Close(ctx)
}
