package metaquery

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/steampipeconfig"
)

var commonCmds = []string{constants.CmdHelp, constants.CmdInspect, constants.CmdExit}

// QueryExecutor :: this is a container interface which allows us to call into the db/Client object
type QueryExecutor interface {
	SetClientSearchPath() error
	GetCurrentSearchPath() ([]string, error)
	CacheOn() error
	CacheOff() error
	CacheClear() error
}

// HandlerInput :: input interface for the metaquery handler
type HandlerInput struct {
	Query       string
	Executor    QueryExecutor
	Schema      *schema.Metadata
	Connections *steampipeconfig.ConnectionMap
	Prompt      *prompt.Prompt
	ClosePrompt func()
}
type PromptControl interface {
	Clear()
	Close()
}

func (h *HandlerInput) args() []string {
	return getArguments(h.Query)
}

type handler func(input *HandlerInput) error

// Handle :: handle metaquery.
func Handle(input *HandlerInput) error {
	input.Query = strings.TrimSuffix(input.Query, ";")
	var s = strings.Fields(input.Query)

	var handlerFunction handler
	metaQueryObj, found := metaQueryDefinitions[s[0]]
	if !found {
		return fmt.Errorf("not sure how to handle '%s'", s[0])
	}

	handlerFunction = metaQueryObj.handler
	return handlerFunction(input)
}

func setOrGetSearchPath(input *HandlerInput) error {
	if len(input.args()) == 0 {
		currentPath, err := input.Executor.GetCurrentSearchPath()
		if err != nil {
			return err
		}
		currentPath = helpers.RemoveFromStringSlice(currentPath, constants.FunctionSchema)

		display.ShowWrappedTable(
			[]string{"search_path"},
			[][]string{
				{strings.Join(currentPath, ",")},
			},
			false,
		)
	} else {
		arg := input.args()[0]
		paths := []string{}
		split := strings.Split(arg, ",")
		for _, s := range split {
			s = strings.TrimSpace(s)
			paths = append(paths, s)
		}
		viper.Set(constants.ArgSearchPath, paths)

		// now that the viper is set, call back into the client (exposed via QueryExecutor) which
		// already knows how to setup the search_paths with the viper values
		return input.Executor.SetClientSearchPath()
	}
	return nil
}

func setSearchPathPrefix(input *HandlerInput) error {
	arg := input.args()[0]
	paths := []string{}
	split := strings.Split(arg, ",")
	for _, s := range split {
		s = strings.TrimSpace(s)
		paths = append(paths, s)
	}
	viper.Set(constants.ArgSearchPathPrefix, paths)

	// now that the viper is set, call back into the client (exposed via QueryExecutor) which
	// already knows how to setup the search_paths with the viper values
	return input.Executor.SetClientSearchPath()
}

// set the ArgHeader viper key with the boolean value evaluated from arg[0]
func setHeader(input *HandlerInput) error {
	cmdconfig.Viper().Set(constants.ArgHeader, typeHelpers.StringToBool(input.args()[0]))
	return nil
}

// set the ArgMulti viper key with the boolean value evaluated from arg[0]
func setMultiLine(input *HandlerInput) error {
	cmdconfig.Viper().Set(constants.ArgMultiLine, typeHelpers.StringToBool(input.args()[0]))
	return nil
}

// controls the cache in the connected FDW
func cacheControl(input *HandlerInput) error {
	command := input.args()[0]
	switch command {
	case constants.ArgOn:
		return input.Executor.CacheOn()
	case constants.ArgOff:
		return input.Executor.CacheOff()
	case constants.ArgClear:
		return input.Executor.CacheClear()
	}

	return fmt.Errorf("invalid command")
}

// set the ArgHeader viper key with the boolean value evaluated from arg[0]
func setTiming(input *HandlerInput) error {
	cmdconfig.Viper().Set(constants.ArgTimer, typeHelpers.StringToBool(input.args()[0]))
	return nil
}

// set the value of `viperKey` in `viper` with the value from `args[0]`
func setViperConfigFromArg(viperKey string) handler {
	return func(input *HandlerInput) error {
		cmdconfig.Viper().Set(viperKey, input.args()[0])
		return nil
	}
}

// set the value of `viperKey` in `viper` with a static value
func setViperConfig(viperKey string, value interface{}) handler {
	return func(input *HandlerInput) error {
		cmdconfig.Viper().Set(viperKey, value)
		return nil
	}
}

// exit
func doExit(input *HandlerInput) error {
	input.ClosePrompt()
	return nil
}

// help
func doHelp(input *HandlerInput) error {
	commonCmdRows := getMetaQueryHelpRows(commonCmds, false)
	var advanceCmds []string
	for cmd := range metaQueryDefinitions {
		if !helpers.StringSliceContains(commonCmds, cmd) {
			advanceCmds = append(advanceCmds, cmd)
		}
	}
	advanceCmdRows := getMetaQueryHelpRows(advanceCmds, true)
	// print out
	fmt.Printf("Welcome to Steampipe shell.\n\nTo start, simply enter your SQL query at the prompt:\n\n  select * from aws_iam_user\n\nCommon commands:\n\n%s\n\nAdvanced commands:\n\n%s\n\nDocumentation available at %s\n",
		buildTable(commonCmdRows, true),
		buildTable(advanceCmdRows, true),
		constants.Bold("https://steampipe.io/docs"))
	fmt.Println()
	return nil
}

func getMetaQueryHelpRows(cmds []string, arrange bool) [][]string {
	var rows [][]string
	for _, cmd := range cmds {
		metaQuery := metaQueryDefinitions[cmd]
		var argsStr []string
		if len(metaQuery.args) > 2 {
			rows = append(rows, []string{cmd + " " + "[mode]", metaQuery.description})
		} else {
			for _, v := range metaQuery.args {
				argsStr = append(argsStr, v.value)
			}
			rows = append(rows, []string{cmd + " " + strings.Join(argsStr, "|"), metaQuery.description})
		}
	}
	// sort by metacmds name
	if arrange {
		sort.SliceStable(rows, func(i, j int) bool {
			return rows[i][0] < rows[j][0]
		})
	}
	return rows
}

// list all the tables in the schema
func listTables(input *HandlerInput) error {

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
		regexp, err := regexp.Compile(arg)
		if err != nil {
			return fmt.Errorf("Invalid search string %s", arg)
		}
		header := []string{"Table", "Schema"}
		rows := [][]string{}
		for schemaName, schemaDetails := range input.Schema.Schemas {
			tables := [][]string{}
			for tableName := range schemaDetails {
				if regexp.MatchString(tableName) {
					tables = append(tables, []string{tableName, schemaName})
				}
			}
			sort.SliceStable(tables, func(i, j int) bool {
				return tables[i][0] < tables[j][0]
			})
			rows = append(rows, tables...)
		}
		display.ShowWrappedTable(header, rows, true)
	}

	return nil
}

// inspect
func inspect(input *HandlerInput) error {
	if len(input.args()) == 0 {
		return listConnections(input)
	}
	// arg can be one of <connection_name> or <connection_name>.<table_name>
	tableOrConnection := input.args()[0]
	split := strings.Split(tableOrConnection, ".")

	if len(split) == 1 {
		// only a connection name (or maybe unqualified table name)
		schemaFound := inspectConnection(tableOrConnection, input)

		// there was no schema
		if !schemaFound {
			searchPath, _ := input.Executor.GetCurrentSearchPath()

			// add the temporary schema to the search_path so that it becomes searchable
			// for the next step
			searchPath = append(searchPath, input.Schema.TemporarySchemaName)

			// go through the searchPath one by one and try to find the table by this name
			for _, schema := range searchPath {
				tablesInThisSchema := input.Schema.GetTablesInSchema(schema)
				// we have a table by this name here
				if helpers.StringSliceContains(tablesInThisSchema, tableOrConnection) {
					return inspectTable(schema, tableOrConnection, input)
				}
			}
			return fmt.Errorf("Could not find Connection or Table called %s", tableOrConnection)
		}

		fmt.Printf(`
To get information about the columns in a table, run %s
	
`, constants.Bold(".inspect {connection}.{table}"))
		return nil
	}

	// this is a fully qualified table name
	return inspectTable(split[0], split[1], input)
}

func listConnections(input *HandlerInput) error {
	header := []string{"connection", "plugin"}
	rows := [][]string{}

	for _, schema := range input.Schema.GetSchemas() {
		if schema == input.Schema.TemporarySchemaName {
			continue
		}
		plugin, found := (*input.Connections)[schema]
		if found {
			rows = append(rows, []string{schema, plugin.Plugin})
		} else {
			rows = append(rows, []string{schema, ""})
		}
	}

	// sort by connection name
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	display.ShowWrappedTable(header, rows, false)

	fmt.Printf(`
To get information about the tables in a connection, run %s
To get information about the columns in a table, run %s

`, constants.Bold(".inspect {connection}"), constants.Bold(".inspect {connection}.{table}"))

	return nil
}

func inspectConnection(connectionName string, input *HandlerInput) bool {
	header := []string{"table", "description"}
	rows := [][]string{}

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

	display.ShowWrappedTable(header, rows, false)

	return true
}

func clearScreen(input *HandlerInput) error {
	input.Prompt.ClearScreen()
	return nil
}

func inspectTable(connectionName string, tableName string, input *HandlerInput) error {
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

	display.ShowWrappedTable(header, rows, false)

	return nil
}

func buildTable(rows [][]string, autoMerge bool) string {
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.Style().Options = table.Options{
		DrawBorder:      false,
		SeparateColumns: false,
		SeparateFooter:  false,
		SeparateHeader:  false,
		SeparateRows:    false,
	}
	t.Style().Box.PaddingLeft = ""

	rowConfig := table.RowConfig{AutoMerge: autoMerge}

	for _, row := range rows {
		rowObj := table.Row{}
		for _, col := range row {
			rowObj = append(rowObj, col)
		}
		t.AppendRow(rowObj, rowConfig)
	}
	return t.Render()
}
