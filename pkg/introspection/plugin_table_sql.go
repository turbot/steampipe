package introspection

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func GetPluginTableCreateSql() db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
				plugin_instance TEXT NULL,
				plugin TEXT NOT NULL,
				memory_max_mb INTEGER,
				limiters JSONB NULL,
				file_name TEXT, 
				start_line_number INTEGER, 
				end_line_number INTEGER
		);`, constants.InternalSchema, constants.PluginConfigTable),
	}
}

func GetPluginTablePopulateSql(plugin *modconfig.Plugin) db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(`INSERT INTO %s.%s (
plugin,
plugin_instance,
memory_max_mb,
limiters,                
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

func GetPluginTableDropSql() db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(
			`DROP TABLE IF EXISTS %s.%s;`,
			constants.InternalSchema,
			constants.PluginConfigTable,
		),
	}
}

func GetPluginTableGrantSql() db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(
			`GRANT SELECT ON TABLE %s.%s to %s;`,
			constants.InternalSchema,
			constants.PluginConfigTable,
			constants.DatabaseUsersRole,
		),
	}
}
