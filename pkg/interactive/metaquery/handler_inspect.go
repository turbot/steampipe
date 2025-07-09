package metaquery

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/querydisplay"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

// inspect
func inspect(ctx context.Context, input *HandlerInput) error {
	connStateMap, err := input.GetConnectionStateMap(ctx)
	if err != nil {
		return err
	}
	if connStateMap == nil {
		log.Printf("[TRACE] failed to load connection state - are we connected to a server running a previous steampipe version?")
		// if there is no connection state, call legacy inspect
		return inspectLegacy(ctx, input)
	}

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

	// here tokens could be schema.tableName or tableName

	if len(tokens) == 1 {
		// only a connection name (or maybe unqualified table name)
		return inspectSchemaOrUnqualifiedTable(ctx, tableOrConnection, input)
	}

	// this is a fully qualified table name
	return inspectQualifiedTable(ctx, tokens[0], tokens[1], input)
}

func inspectSchemaOrUnqualifiedTable(ctx context.Context, tableOrConnection string, input *HandlerInput) error {
	// only a connection name (or maybe unqualified table name)
	if inspectConnection(ctx, tableOrConnection, input) {
		return nil
	}

	// there was no schema
	// add the temporary schema to the search_path so that it becomes searchable
	// for the next step
	//nolint:golint,gocritic // we don't want to modify the input value
	searchPath := append(input.SearchPath, input.Schema.TemporarySchemaName)

	// go through the searchPath one by one and try to find the table by this name
	for _, schema := range searchPath {
		tablesInThisSchema := input.Schema.GetTablesInSchema(schema)
		// we have a table by this name here
		if _, gotTable := tablesInThisSchema[tableOrConnection]; gotTable {
			return inspectQualifiedTable(ctx, schema, tableOrConnection, input)
		}

		// check against the fully qualified name of the table
		for _, table := range input.Schema.Schemas[schema] {
			if tableOrConnection == table.FullName {
				return inspectQualifiedTable(ctx, schema, table.Name, input)
			}
		}
	}

	return fmt.Errorf("could not find connection or table called '%s'. Is the plugin installed? Is the connection configured?", tableOrConnection)
}

// list all the tables in the schema
func listTables(ctx context.Context, input *HandlerInput) error {

	if len(input.args()) == 0 {
		schemas := input.Schema.GetSchemas()
		for _, schema := range schemas {
			if schema == input.Schema.TemporarySchemaName {
				continue
			}
			fmt.Printf(" ==> %s\n", schema)
			inspectConnection(ctx, schema, input)
		}

		fmt.Printf(`
To get information about the columns in a table, run %s
	
`, pconstants.Bold(".inspect {connection}.{table}"))
	} else {
		// could be one of connectionName and {string}*
		arg := input.args()[0]
		if !strings.HasSuffix(arg, "*") {
			inspectConnection(ctx, arg, input)
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
		querydisplay.ShowWrappedTable(header, rows, &querydisplay.ShowWrappedTableOptions{AutoMerge: true})
	}

	return nil
}

func listConnections(ctx context.Context, input *HandlerInput) error {
	connStateMap, err := input.GetConnectionStateMap(ctx)
	if err != nil {
		return err
	}
	// if there is no connection state in the input, call listConnectionsLegacy
	if connStateMap == nil {
		log.Printf("[TRACE] failed to load connection state - are we connected to a server running a previous steampipe version?")
		// call legacy inspect
		return listConnectionsLegacy(ctx, input)
	}

	header := []string{"connection", "plugin", "state"}

	connectionState, err := input.GetConnectionStateMap(ctx)
	if err != nil {
		return err
	}
	showStateSummary := connectionState.ConnectionsInState(
		constants.ConnectionStateUpdating,
		constants.ConnectionStateDeleting,
		constants.ConnectionStateError)

	var rows [][]string

	for connectionName, state := range connectionState {
		// skip disabled connections
		if state.Disabled() {
			continue
		}
		row := []string{connectionName, state.Plugin, state.State}
		rows = append(rows, row)
	}

	// sort by connection name
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	querydisplay.ShowWrappedTable(header, rows, &querydisplay.ShowWrappedTableOptions{AutoMerge: false})

	if showStateSummary {
		showStateSummaryTable(connectionState)
	}

	fmt.Printf(`
To get information about the tables in a connection, run %s
To get information about the columns in a table, run %s

`, pconstants.Bold(".inspect {connection}"), pconstants.Bold(".inspect {connection}.{table}"))

	return nil
}

func showStateSummaryTable(connectionState steampipeconfig.ConnectionStateMap) {
	header := []string{"Connection state", "Count"}
	var rows [][]string
	stateSummary := connectionState.GetSummary()

	for _, state := range constants.ConnectionStates {
		if connectionsInState := stateSummary[state]; connectionsInState > 0 {
			rows = append(rows, []string{state, fmt.Sprintf("%d", connectionsInState)})
		}
	}
	querydisplay.ShowWrappedTable(header, rows, &querydisplay.ShowWrappedTableOptions{AutoMerge: false})
}

func inspectQualifiedTable(ctx context.Context, connectionName string, tableName string, input *HandlerInput) error {
	header := []string{"column", "type", "description"}
	var rows [][]string

	connectionStateMap, err := input.GetConnectionStateMap(ctx)
	if err != nil {
		return err
	}
	// do we have connection state for this schema and if so is it disabled?
	if connectionState := connectionStateMap[connectionName]; connectionState != nil && connectionState.Disabled() {
		error_helpers.ShowWarning(fmt.Sprintf("connection '%s' has schema import disabled", connectionName))
		return nil
	}

	schema, found := input.Schema.Schemas[connectionName]
	if !found {
		return fmt.Errorf("could not find connection called '%s'. Is the plugin installed? Is the connection configured?\n", connectionName)
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

	querydisplay.ShowWrappedTable(header, rows, &querydisplay.ShowWrappedTableOptions{AutoMerge: false})

	return nil
}

// inspect the connection with the given name
// return whether connectionName was identified as an existing connection
func inspectConnection(ctx context.Context, connectionName string, input *HandlerInput) bool {
	connectionStateMap, err := input.GetConnectionStateMap(ctx)
	if err != nil {
		error_helpers.ShowError(ctx, sperr.WrapWithMessage(err, "connection '%s' has schema import disabled", connectionName))
		return true
	}

	connectionState, connectionFoundInState := connectionStateMap[connectionName]
	if !connectionFoundInState {
		return false
	}
	if connectionState.Disabled() {
		error_helpers.ShowWarning(fmt.Sprintf("connection '%s' has schema import disabled", connectionName))
		return true
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

	querydisplay.ShowWrappedTable(header, rows, &querydisplay.ShowWrappedTableOptions{AutoMerge: false})

	return true
}
