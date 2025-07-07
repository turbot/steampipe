package introspection

import (
	"fmt"

	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
)

func GetPluginTableCreateSql() db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
				plugin_instance TEXT,
				plugin TEXT NOT NULL,
				version TEXT ,
				memory_max_mb INTEGER,
				limiters JSONB,
				file_name TEXT, 
				start_line_number INTEGER, 
				end_line_number INTEGER				
		);`, constants.InternalSchema, constants.PluginInstanceTable),
	}
}

func GetPluginTablePopulateSql(plugin *plugin.Plugin) db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(`INSERT INTO %s.%s (
plugin,
version,
plugin_instance,
memory_max_mb,
limiters,                
file_name,
start_line_number,
end_line_number
)
	VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, constants.InternalSchema, constants.PluginInstanceTable),
		Args: []any{
			plugin.Plugin,
			plugin.Version,
			plugin.Instance,
			plugin.MemoryMaxMb,
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
			constants.PluginInstanceTable,
		),
	}
}

func GetPluginTableGrantSql() db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(
			`GRANT SELECT ON TABLE %s.%s to %s;`,
			constants.InternalSchema,
			constants.PluginInstanceTable,
			constants.DatabaseUsersRole,
		),
	}
}
