package introspection

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
)

func GetPluginColumnTableCreateSql() db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
				plugin TEXT NOT NULL,
				table_name TEXT NOT NULL,
				name TEXT NOT NULL,
				type TEXT NOT NULL,
				description TEXT NULL,
				list_config jsonb NULL,
				get_config jsonb NULL,
				hydrate_name TEXT NULL,
				default_value jsonb NULL
		);`, constants.InternalSchema, constants.PluginColumnTable),
	}
}

func GetPluginColumnTablePopulateSqlForPlugin(pluginName string, schema map[string]*proto.TableSchema) ([]db_common.QueryWithArgs, error) {
	var res []db_common.QueryWithArgs
	for tableName, tableSchema := range schema {
		getKeyColumns := tableSchema.GetKeyColumnMap()
		listKeyColumns := tableSchema.ListKeyColumnMap()
		for _, columnSchema := range tableSchema.Columns {
			getKeyColumn := getKeyColumns[columnSchema.Name]
			listKeyColumn := listKeyColumns[columnSchema.Name]
			q, err := GetPluginColumnTablePopulateSql(pluginName, tableName, columnSchema, getKeyColumn, listKeyColumn)
			if err != nil {
				return nil, err
			}
			res = append(res, q)
		}
	}
	return res, nil
}

func GetPluginColumnTablePopulateSql(
	pluginName, tableName string,
	columnSchema *proto.ColumnDefinition,
	getKeyColumn, listKeyColumn *proto.KeyColumn) (db_common.QueryWithArgs, error) {

	var description, defaultValue any
	if columnSchema.Description != "" {
		description = columnSchema.Description
	}
	if columnSchema.Default != nil {
		var err error
		defaultValue, err = columnSchema.Default.ValueToInterface()
		if err != nil {
			return db_common.QueryWithArgs{}, err
		}
	}

	var listConfig, getConfig *keyColumn

	if getKeyColumn != nil {
		getConfig = newKeyColumn(getKeyColumn.Operators, getKeyColumn.Require, getKeyColumn.CacheMatch)
	}
	if listKeyColumn != nil {
		listConfig = newKeyColumn(listKeyColumn.Operators, listKeyColumn.Require, listKeyColumn.CacheMatch)
	}

	// special handling for strings
	if s, ok := defaultValue.(string); ok {
		defaultValue = fmt.Sprintf(`"%s"`, s)
	}
	var hydrate any = nil
	if columnSchema.Hydrate != "" {
		hydrate = columnSchema.Hydrate
	}

	q := db_common.QueryWithArgs{
		Query: fmt.Sprintf(`INSERT INTO %s.%s (
				plugin,
				table_name ,
				name,
				type,
				description,
				list_config,
				get_config,
				hydrate_name,
				default_value
)
	VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, constants.InternalSchema, constants.PluginColumnTable),
		Args: []any{
			pluginName,
			tableName,
			columnSchema.Name,
			proto.ColumnType_name[int32(columnSchema.Type)],
			description,
			listConfig,
			getConfig,
			hydrate,
			defaultValue,
		},
	}

	return q, nil
}

func GetPluginColumnTableDropSql() db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(
			`DROP TABLE IF EXISTS %s.%s;`,
			constants.InternalSchema,
			constants.PluginColumnTable,
		),
	}
}

func GetPluginColumnTableDeletePluginSql(plugin string) db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(
			`DELETE FROM %s.%s
WHERE plugin = $1;`,
			constants.InternalSchema,
			constants.PluginColumnTable,
		),
		Args: []any{plugin},
	}
}

func GetPluginColumnTableGrantSql() db_common.QueryWithArgs {
	return db_common.QueryWithArgs{
		Query: fmt.Sprintf(
			`GRANT SELECT ON TABLE %s.%s to %s;`,
			constants.InternalSchema,
			constants.PluginColumnTable,
			constants.DatabaseUsersRole,
		),
	}
}

type keyColumn struct {
	Operators  []string `json:"operators,omitempty"`
	Require    string   `json:"require,omitempty"`
	CacheMatch string   `json:"cache_match,omitempty"`
}

func newKeyColumn(operators []string, require string, cacheMatch string) *keyColumn {
	return &keyColumn{
		Operators:  cleanOperators(operators),
		Require:    require,
		CacheMatch: cacheMatch,
	}
}

// tactical - avoid html encoding operators
func cleanOperators(operators []string) []string {
	var res = make([]string, len(operators))
	for i, operator := range operators {

		switch operator {
		case "<>":
			operator = "!="
		case ">":
			operator = "gt"
		case "<":
			operator = "lt"
		case ">=":
			operator = "ge"
		case "<=":
			operator = "le"
		}
		res[i] = operator
	}
	return res
}

// MarshalJSON implements the json.Marshaler interface
// This method is responsible for providing the custom JSON encoding
func (s keyColumn) MarshalJSON() ([]byte, error) {
	type Alias keyColumn

	b := new(strings.Builder)
	encoder := json.NewEncoder(b)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(Alias(s))
	return []byte(b.String()), err
}
