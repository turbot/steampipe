package metaquery

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/querydisplay"
)

// inspect
func inspectLegacy(ctx context.Context, input *HandlerInput) error {
	if len(input.args()) == 0 {
		return listConnectionsLegacy(ctx, input)
	}
	tableOrConnection := input.args()[0]
	if len(input.args()) > 0 {
		// this should be one argument, but may have been split by the tokenizer
		// because of the escape characters that autocomplete puts in
		// join them up
		tableOrConnection = strings.Join(input.args(), " ")
	}

	// remove all double quotes (if any)
	tableOrConnection = strings.Join(
		strings.Split(tableOrConnection, "\""),
		"",
	)

	// arg can be one of <connection_name> or <connection_name>.<table_name>
	tokens := strings.SplitN(tableOrConnection, ".", 2)

	// here tokens could be schema.tablename
	// or table.name
	// or both

	if len(tokens) > 0 {
		// only a connection name (or maybe unqualified table name)
		schemaFound := inspectConnectionLegacy(tableOrConnection, input)

		// there was no schema
		if !schemaFound {
			// we couldn't find a schema with the name
			// try a prefix search with the schema name
			// for schema := range input.Schema.Schemas {
			// 	if strings.HasPrefix(tableOrConnection, schema) {
			// 		tableName := strings.TrimPrefix(tableOrConnection, fmt.Sprintf("%s.", schema))
			// 		return inspectTable(schema, tableName, input)
			// 	}
			// }

			// still here - the last sledge hammer is to go through
			// the schema names one by one
			searchPath := input.Client.GetRequiredSessionSearchPath()

			// add the temporary schema to the search_path so that it becomes searchable
			// for the next step
			searchPath = append(searchPath, input.Schema.TemporarySchemaName)

			// go through the searchPath one by one and try to find the table by this name
			for _, schema := range searchPath {
				tablesInThisSchema := input.Schema.GetTablesInSchema(schema)
				// we have a table by this name here
				if _, foundTable := tablesInThisSchema[tableOrConnection]; foundTable {
					return inspectTableLegacy(schema, tableOrConnection, input)
				}

				// check against the fully qualified name of the table
				for _, table := range input.Schema.Schemas[schema] {
					if tableOrConnection == table.FullName {
						return inspectTableLegacy(schema, table.Name, input)
					}
				}
			}

			return fmt.Errorf("could not find connection or table called '%s'. Is the plugin installed? Is the connection configured?", tableOrConnection)
		}

		fmt.Printf(`
To get information about the columns in a table, run %s
	
`, constants.Bold(".inspect {connection}.{table}"))
		return nil
	}

	// this is a fully qualified table name
	return inspectTableLegacy(tokens[0], tokens[1], input)
}

func listConnectionsLegacy(ctx context.Context, input *HandlerInput) error {
	header := []string{"connection", "plugin"}
	var rows [][]string

	for _, schema := range input.Schema.GetSchemas() {
		if schema == input.Schema.TemporarySchemaName {
			continue
		}
		rows = append(rows, []string{schema, ""})

	}

	// sort by connection name
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	querydisplay.ShowWrappedTable(header, rows, &querydisplay.ShowWrappedTableOptions{AutoMerge: false})

	fmt.Printf(`
To get information about the tables in a connection, run %s
To get information about the columns in a table, run %s

`, constants.Bold(".inspect {connection}"), constants.Bold(".inspect {connection}.{table}"))

	return nil
}

func inspectConnectionLegacy(connectionName string, input *HandlerInput) bool {
	header := []string{"table", "description"}
	var rows [][]string

	schema, found := input.Schema.Schemas[connectionName]

	if !found {
		return false
	}

	for _, tableSchema := range schema {
		rows = append(rows, []string{tableSchema.Name, tableSchema.Description})
	}

	// sort by table name
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	querydisplay.ShowWrappedTable(header, rows, &querydisplay.ShowWrappedTableOptions{AutoMerge: false})

	return true
}

func inspectTableLegacy(connectionName string, tableName string, input *HandlerInput) error {
	header := []string{"column", "type", "description"}
	rows := [][]string{}

	schema, found := input.Schema.Schemas[connectionName]
	if !found {
		return fmt.Errorf("Could not find connection called '%s'", connectionName)
	}
	tableSchema, found := schema[tableName]
	if !found {
		return fmt.Errorf("Could not find table '%s' in '%s'", tableName, connectionName)
	}

	for _, columnSchema := range tableSchema.Columns {
		rows = append(rows, []string{columnSchema.Name, columnSchema.Type, columnSchema.Description})
	}

	// sort by column name
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	querydisplay.ShowWrappedTable(header, rows, &querydisplay.ShowWrappedTableOptions{AutoMerge: false})

	return nil
}
