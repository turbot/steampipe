package introspection

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func CreatePluginTable() db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
				plugin TEXT NOT NULL,
				plugin_instance TEXT NULL,
				max_memory_mb INTEGER,
				rate_limiters JSONB NULL,
				file_name TEXT NOT NULL, 
				start_line_number INTEGER NOT NULL, 
				end_line_number INTEGER NOT NULL
		);`, constants.InternalSchema, constants.PluginConfigTable),
	}
}

func GetPopulatePluginSql(plugin *modconfig.Plugin) db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(`INSERT INTO %s.%s (
plugin,
plugin_instance,
max_memory_mb,
rate_limiters,                
file_name,
start_line_number,
end_line_number
)
	VALUES($1,$2,$3,$4,$5,$6,$7)`, constants.InternalSchema, constants.PluginConfigTable),
		Args: []any{
			plugin.Plugin,
			plugin.Instance,
			plugin.MaxMemoryMb,
			plugin.Limiters,
			plugin.FileName,
			plugin.StartLineNumber,
			plugin.EndLineNumber,
		},
	}
}

func DropPluginTable() db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(
			`DROP TABLE IF EXISTS %s.%s;`,
			constants.InternalSchema,
			constants.PluginConfigTable,
		),
	}
}

func GrantsOnPluginTable() db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(
			`GRANT SELECT ON TABLE %s.%s to %s;`,
			constants.InternalSchema,
			constants.PluginConfigTable,
			constants.DatabaseUsersRole,
		),
	}
}
