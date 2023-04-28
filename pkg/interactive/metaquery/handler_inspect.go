package metaquery

import (
	"context"
	"fmt"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/statushooks"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/display"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
)

// inspect
func inspect(ctx context.Context, input *HandlerInput) error {
	// load connection state and put into input
	connectionState, err := getConnectionState(ctx, input.Client)
	if err != nil {
		return err
	}
	input.ConnectionState = connectionState

	// if no args were provided just list connections
	if len(input.args()) == 0 {
		return listConnections(ctx, input)
	}

	// so there were args, try to determine what the args are
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

	// here tokens could be schema.tableName
	// or tableName
	if len(tokens) > 0 {
		// only a connection name (or maybe unqualified table name)
		schemaFound := inspectConnection(tableOrConnection, input)

		// there was no schema
		if !schemaFound {
			// add the temporary schema to the search_path so that it becomes searchable
			// for the next step
			searchPath := append(input.SearchPath, input.Schema.TemporarySchemaName)

			// go through the searchPath one by one and try to find the table by this name
			for _, schema := range searchPath {
				tablesInThisSchema := input.Schema.GetTablesInSchema(schema)
				// we have a table by this name here
				if _, gotTable := tablesInThisSchema[tableOrConnection]; gotTable {
					return inspectTable(schema, tableOrConnection, input)
				}

				// check against the fully qualified name of the table
				for _, table := range input.Schema.Schemas[schema] {
					if tableOrConnection == table.FullName {
						return inspectTable(schema, table.Name, input)
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
	return inspectTable(tokens[0], tokens[1], input)
}

// list all the tables in the schema
func listTables(_ context.Context, input *HandlerInput) error {

	if len(input.args()) == 0 {
		schemas := input.Schema.GetSchemas()
		for _, schema := range schemas {
			if schema == input.Schema.TemporarySchemaName {
				continue
			}
			fmt.Printf(" ==> %s\n", schema)
			inspectConnection(schema, input)
		}

		fmt.Printf(`
To get information about the columns in a table, run %s
	
`, constants.Bold(".inspect {connection}.{table}"))
	} else {
		// could be one of connectionName and {string}*
		arg := input.args()[0]
		if !strings.HasSuffix(arg, "*") {
			inspectConnection(arg, input)
			fmt.Println()
			return nil
		}

		// treat this as a wild card
		r, err := regexp.Compile(arg)
		if err != nil {
			return fmt.Errorf("invalid search string %s", arg)
		}
		header := []string{"Table", "Schema"}
		var rows [][]string
		for schemaName, schemaDetails := range input.Schema.Schemas {
			var tables [][]string
			for tableName := range schemaDetails {
				if r.MatchString(tableName) {
					tables = append(tables, []string{tableName, schemaName})
				}
			}
			sort.SliceStable(tables, func(i, j int) bool {
				return tables[i][0] < tables[j][0]
			})
			rows = append(rows, tables...)
		}
		display.ShowWrappedTable(header, rows, &display.ShowWrappedTableOptions{AutoMerge: true})
	}

	return nil
}

func listConnections(_ context.Context, input *HandlerInput) error {
	header := []string{"connection", "plugin"}
	showState := input.ConnectionState.ConnectionsInState(
		constants.ConnectionStateUpdating,
		constants.ConnectionStateDeleting,
		constants.ConnectionStateError)
	if showState {
		header = append(header, "state")
	}
	var rows [][]string

	for connectionName, state := range input.ConnectionState {
		row := []string{connectionName, state.Plugin}
		if showState {
			row = append(row, state.State)
		}
		rows = append(rows, row)
	}

	// sort by connection name
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	display.ShowWrappedTable(header, rows, &display.ShowWrappedTableOptions{AutoMerge: false})

	if showState {

		showStateSummary(input.ConnectionState)
	}

	fmt.Printf(`
To get information about the tables in a connection, run %s
To get information about the columns in a table, run %s

`, constants.Bold(".inspect {connection}"), constants.Bold(".inspect {connection}.{table}"))

	return nil
}

func showStateSummary(connectionState steampipeconfig.ConnectionDataMap) {
	header := []string{"Connection state", "Count"}
	var rows [][]string
	stateSummary := connectionState.GetSummary()

	for _, state := range constants.ConnectionStates {
		if connectionsInState := stateSummary[state]; connectionsInState > 0 {
			rows = append(rows, []string{state, fmt.Sprintf("%d", connectionsInState)})
		}
	}
	display.ShowWrappedTable(header, rows, &display.ShowWrappedTableOptions{AutoMerge: false})
}

func inspectTable(connectionName string, tableName string, input *HandlerInput) error {
	header := []string{"column", "type", "description"}
	var rows [][]string

	schema, found := input.Schema.Schemas[connectionName]
	if !found {
		return fmt.Errorf("could not find connection called '%s'", connectionName)
	}
	tableSchema, found := schema[tableName]
	if !found {
		return fmt.Errorf("could not find table '%s' in '%s'", tableName, connectionName)
	}

	for _, columnSchema := range tableSchema.Columns {
		rows = append(rows, []string{columnSchema.Name, columnSchema.Type, columnSchema.Description})
	}

	// sort by column name
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	display.ShowWrappedTable(header, rows, &display.ShowWrappedTableOptions{AutoMerge: false})

	return nil
}

// inspect the connection with the given name
// return whether connectionName was identified as an existing connection
func inspectConnection(connectionName string, input *HandlerInput) bool {

	connectionState, connectionFoundInState := input.ConnectionState[connectionName]
	if !connectionFoundInState {
		return false
	}

	// have we loaded the schema for this connection yet?
	schema, found := input.Schema.Schemas[connectionName]

	var rows [][]string
	var header []string

	if found {
		header = []string{"table", "description"}
		for _, tableSchema := range schema {
			rows = append(rows, []string{tableSchema.Name, tableSchema.Description})
		}
	} else {
		// just display the connection state
		header = []string{"connection", "plugin", "schema mode", "state", "error", "state updated"}

		rows = [][]string{{
			connectionName,
			connectionState.Plugin,
			connectionState.SchemaMode,
			connectionState.State,
			connectionState.Error(),
			connectionState.ConnectionModTime.Format(time.RFC3339),
		},
		}
	}

	// sort by table name
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	display.ShowWrappedTable(header, rows, &display.ShowWrappedTableOptions{AutoMerge: false})

	return true
}

// helper function to acquire db connection and retrieve connection
func getConnectionState(ctx context.Context, client db_common.Client) (steampipeconfig.ConnectionDataMap, error) {
	statushooks.Show(ctx)
	defer statushooks.Done(ctx)

	statushooks.SetStatus(ctx, "Loading connection state...")

	conn, err := client.AcquireConnection(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	return steampipeconfig.LoadConnectionState(ctx, conn.Conn(), steampipeconfig.WithWaitForPending)
}
