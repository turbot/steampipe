package db_common

import (
	"fmt"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"strings"
)

func GetCommentsQueryForPlugin(connectionName string, p map[string]*proto.TableSchema) string {
	var statements strings.Builder
	for t, schema := range p {
		table := PgEscapeName(t)
		schemaName := PgEscapeName(connectionName)
		if schema.Description != "" {
			tableDescription := PgEscapeString(schema.Description)
			statements.WriteString(fmt.Sprintf("COMMENT ON FOREIGN TABLE %s.%s is %s;\n", schemaName, table, tableDescription))
		}
		for _, c := range schema.Columns {
			if c.Description != "" {
				column := PgEscapeName(c.Name)
				columnDescription := PgEscapeString(c.Description)
				statements.WriteString(fmt.Sprintf("COMMENT ON COLUMN %s.%s.%s is %s;\n", schemaName, table, column, columnDescription))
			}
		}
	}
	return statements.String()
}

func GetUpdateConnectionQuery(connectionName, pluginSchemaName string) string {
	// escape the name
	connectionName = PgEscapeName(connectionName)

	var statements strings.Builder

	// Each connection has a unique schema. The schema, and all objects inside it,
	// are owned by the root user.
	statements.WriteString(fmt.Sprintf("drop schema if exists %s cascade;\n", connectionName))
	statements.WriteString(fmt.Sprintf("create schema %s;\n", connectionName))
	statements.WriteString(fmt.Sprintf("comment on schema %s is 'steampipe plugin: %s';\n", connectionName, pluginSchemaName))

	// Steampipe users are allowed to use the new schema
	statements.WriteString(fmt.Sprintf("grant usage on schema %s to steampipe_users;\n", connectionName))

	// Permissions are limited to select only, and should be granted for all new
	// objects. Steampipe users cannot create tables or modify data in the
	// connection schema - they need to use the public schema for that.  These
	// commands alter the defaults for any objects created in the future.
	// See https://www.postgresql.org/docs/12/ddl-priv.html
	statements.WriteString(fmt.Sprintf("alter default privileges in schema %s grant select on tables to steampipe_users;\n", connectionName))

	// If there are any objects already then grant their permissions now. (This
	// should not actually do anything at this point.)
	statements.WriteString(fmt.Sprintf("grant select on all tables in schema %s to steampipe_users;\n", connectionName))

	// Import the foreign schema into this connection.
	statements.WriteString(fmt.Sprintf("import foreign schema \"%s\" from server steampipe into %s;\n", pluginSchemaName, connectionName))

	return statements.String()
}

func GetDeleteConnectionQuery(name string) string {
	return fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;\n", PgEscapeName(name))
}
