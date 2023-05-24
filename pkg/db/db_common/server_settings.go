package db_common

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/sperr"
)

type ServerSettingKey string
type ServerSettings struct {
	StartTime        time.Time
	SteampipeVersion string
	FdwVersion       string
	CacheMaxTtl      int
	CacheMaxSizeMb   int
	CacheEnabled     bool

	// a private property
	// defaults to false - set to true after loading completes
	loaded bool
}

// StubServerSettings returns a server settings struct which is maked as unloaded
func StubServerSettings() *ServerSettings {
	return new(ServerSettings)
}

func LoadServerSettings(ctx context.Context, conn *pgx.Conn) (*ServerSettings, error) {
	rows, err := conn.Query(ctx, fmt.Sprintf("SELECT name,value FROM %s.%s", constants.InternalSchema, constants.ServerSettingsTable))
	if err != nil {
		return nil, sperr.WrapWithMessage(err, "could not load %s.%s", constants.InternalSchema, constants.ServerSettingsTable)
	}
	defer rows.Close()
	settings := new(ServerSettings)
	for rows.Next() {
		var name ServerSettingKey
		var value any
		if err := rows.Scan(&name, &value); err != nil {
			return nil, sperr.WrapWithMessage(err, "error reading row from %s.%s", constants.InternalSchema, constants.ServerSettingsTable)
		}

		switch name {
		case ServerSettingSteampipeVersion:
			settings.SteampipeVersion = value.(string)
		case ServerSettingFdwersion:
			settings.FdwVersion = value.(string)
		case ServerSettingStartTime:
			if st, err := time.Parse(time.RFC3339, value.(string)); err == nil {
				settings.StartTime = st
			}
		case ServerSettingCacheEnabled:
			settings.CacheEnabled = value.(bool)
		case ServerSettingCacheMaxTtl:
			// the value (although written as an integer), is evaluated as a
			// float64 by the driver
			// we need to cast it
			settings.CacheMaxTtl = int(value.(float64))
		case ServerSettingCacheMaxSizeMb:
			// the value (although written as an integer), is evaluated as a
			// float64 by the driver
			// we need to cast it
			settings.CacheMaxSizeMb = int(value.(float64))
		default:
			log.Printf(
				"[INFO] unknown key '%s' with value '%v(%s)' found in %s.%s - Skipping",
				name, value, reflect.TypeOf(value).Kind(),
				constants.InternalSchema, constants.ServerSettingsTable,
			)
		}
	}
	settings.loaded = true
	return settings, nil
}

// Loaded returns a bool indicating whether settings data has been loaded
func (s *ServerSettings) Loaded(ctx context.Context) bool {
	return s.loaded
}

// SetupSql returns the set of SQL statements to fully replace any existing
// settings table with a new one and populates the values
func (s *ServerSettings) SetupSql(ctx context.Context) []QueryWithArgs {
	utils.LogTime("db_local.initializeServerSettingsTable start")
	defer utils.LogTime("db_local.initializeServerSettingsTable end")

	settings := map[ServerSettingKey]any{
		ServerSettingSteampipeVersion: s.SteampipeVersion,
		ServerSettingFdwersion:        s.FdwVersion,
		ServerSettingStartTime:        time.Now().UTC().Format(time.RFC3339),
		ServerSettingCacheEnabled:     viper.GetBool(constants.ArgServiceCacheEnabled),
		ServerSettingCacheMaxTtl:      viper.GetInt(constants.ArgCacheMaxTtl),
		ServerSettingCacheMaxSizeMb:   viper.GetInt(constants.ArgMaxCacheSizeMb),
	}

	// start with a clean slate
	queries := []QueryWithArgs{
		// drop the old table (alternative is "if exists then truncate" which is more expensive)
		// this also allows us to modify the table structure without having to go through complex
		// migrations
		getServerSettingsTableDropSQL(ctx),
		// create a new one
		getServerSettingsTableCreateSQL(ctx),
		// grants
		getServerSettingsTableGrantSQL(ctx),
	}

	queries = append(queries, getServerSettingsRowSql(ctx, settings)...)

	return queries
}

func getServerSettingsRowSql(_ context.Context, settings map[ServerSettingKey]any) []QueryWithArgs {
	queries := []QueryWithArgs{}
	for name, value := range settings {
		dataType := "text"
		kind := reflect.TypeOf(value).Kind()
		switch kind {
		case reflect.Bool:
			dataType = "bool"
		case reflect.String:
			dataType = "text"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
			dataType = "integer"
		default:
			log.Printf("[INFO] Skipping unknown server setting type '%s' for '%s' (%v)", kind.String(), name, value)
			continue
		}

		queries = append(queries, QueryWithArgs{
			Query: fmt.Sprintf(
				`INSERT INTO %s.%s (name,value,vartype) VALUES ($1,TO_JSONB($2::%s),$3)`,
				constants.InternalSchema,
				constants.ServerSettingsTable,
				dataType,
			),
			Args: []any{name, value, dataType},
		})
	}
	return queries
}

func getServerSettingsTableGrantSQL(_ context.Context) QueryWithArgs {
	return QueryWithArgs{
		Query: fmt.Sprintf(
			`GRANT SELECT ON TABLE %s.%s to %s;`,
			constants.InternalSchema,
			constants.ServerSettingsTable,
			constants.DatabaseUsersRole,
		),
	}
}

func getServerSettingsTableCreateSQL(_ context.Context) QueryWithArgs {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
		name TEXT PRIMARY KEY,
		value JSONB NOT NULL,
		vartype TEXT NOT NULL
		);`, constants.InternalSchema, constants.ServerSettingsTable)

	return QueryWithArgs{Query: query}
}

func getServerSettingsTableDropSQL(_ context.Context) QueryWithArgs {
	query := fmt.Sprintf(
		`DROP TABLE IF EXISTS %s.%s;`,
		constants.InternalSchema,
		constants.ServerSettingsTable,
	)

	return QueryWithArgs{Query: query}
}

const (
	ServerSettingSteampipeVersion ServerSettingKey = "steampipe_version"
	ServerSettingFdwersion        ServerSettingKey = "fdw_version"
	ServerSettingStartTime        ServerSettingKey = "start_time"
	ServerSettingCacheEnabled     ServerSettingKey = "cache_enabled"
	ServerSettingCacheMaxTtl      ServerSettingKey = "cache_max_ttl"
	ServerSettingCacheMaxSizeMb   ServerSettingKey = "cache_max_size_mb"
)
