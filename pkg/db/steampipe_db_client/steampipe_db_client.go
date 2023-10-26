package steampipe_db_client

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/turbot/steampipe/pkg/db/steampipe_db_common"
	"github.com/turbot/steampipe/pkg/serversettings"
	"log"
	"strings"

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
	sessions map[uint32]*steampipe_db_common.DatabaseSession

	// allows locked access to the 'sessions' map
	sessionsMutex *sync.Mutex

	ServerSettings *steampipe_db_common.ServerSettings

	// TODO KAI POPULATE THIS
	// this flag is set if the service that this client
	// is connected to is running in the same physical system
	isLocalService bool
}

func NewSteampipeDbClient(ctx context.Context, dbClient *db_client.DbClient, onConnectionCallback DbConnectionCallback) (*SteampipeDbClient, error) {

	steampipeClient := &SteampipeDbClient{
		DbClient:             *dbClient,
		onConnectionCallback: onConnectionCallback,
		sessions:             make(map[uint32]*steampipe_db_common.DatabaseSession),
		sessionsMutex:        &sync.Mutex{},
	}

	// TODO KAI FIGURE THIS OUT
	// set pre execute hook to re-read ArgTiming from viper (in case the .timing command has been run)
	// (this will refetch ScanMetadataMaxId if timing has just been enabled)
	//dbClient.BeforeExecuteHook = steampipeClient.setShouldShowTiming

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

	//	load up the server settings
	if err := steampipeClient.loadServerSettings(ctx); err != nil {
		return nil, err
	}

	return steampipeClient, nil
}

// TODO KAI session keying mechanism - re-add session map
// ScanMetadataMaxId broken

// TODO KAI isLocalService
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

func (c *SteampipeDbClient) setShouldShowTiming(ctx context.Context, session *steampipe_db_common.DatabaseSession) error {
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

func (c *SteampipeDbClient) getQueryTiming(ctx context.Context, startTime time.Time, session *steampipe_db_common.DatabaseSession, resultChannel chan *queryresult.TimingResult) {
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

func (c *SteampipeDbClient) updateScanMetadataMaxId(ctx context.Context, session *steampipe_db_common.DatabaseSession) error {
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
	// clear active sessions
	c.sessions = nil

	return c.DbClient.Close(ctx)
}

func (c *SteampipeDbClient) loadServerSettings(ctx context.Context) error {
	serverSettings, err := serversettings.Load(ctx, c.ManagementPool)
	if err != nil {
		if notFound := db_common.IsRelationNotFoundError(err); notFound {
			// when connecting to pre-0.21.0 services, the steampipe_server_settings table will not be available.
			// this is expected and not an error
			// code which uses steampipe_server_settings should handle this
			log.Printf("[TRACE] could not find %s.%s table. skipping\n", constants.InternalSchema, constants.ServerSettingsTable)
			return nil
		}
		return err
	}
	c.ServerSettings = serverSettings
	log.Println("[TRACE] loaded server settings:", serverSettings)
	return nil
}

// ensure the search path for the database session is as required
func (c *SteampipeDbClient) ensureSessionSearchPath(ctx context.Context, session *steampipe_db_common.DatabaseSession) error {
	log.Printf("[TRACE] ensureSessionSearchPath")

	// update the stored value of user search path
	// this might have changed if a connection has been added/removed
	if err := c.LoadUserSearchPathForConnection(ctx, session.Connection); err != nil {
		return err
	}

	// get the required search path which is either a custom search path (if present) or the user search path
	requiredSearchPath := c.GetRequiredSessionSearchPath()

	// now determine whether the session search path is the same as the required search path
	// if so, return
	if strings.Join(session.SearchPath, ",") == strings.Join(requiredSearchPath, ",") {
		log.Printf("[TRACE] session search path is already correct - nothing to do")
		return nil
	}

	// so we need to set the search path
	log.Printf("[TRACE] session search path will be updated to  %s", strings.Join(c.CustomSearchPath, ","))

	err := db_common.ExecuteSystemClientCall(ctx, session.Connection, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, fmt.Sprintf("set search_path to %s", strings.Join(db_common.PgEscapeSearchPath(requiredSearchPath), ",")))
		return err
	})

	if err == nil {
		// update the session search path property
		session.SearchPath = requiredSearchPath
	}
	return err
}
