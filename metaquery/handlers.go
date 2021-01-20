package metaquery

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/c-bata/go-prompt"

	"github.com/olekukonko/tablewriter"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/connection_config"

	typeHelpers "github.com/turbot/go-kit/types"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/utils"
)

// HandlerInput :: input interface for the metaquery handler
type HandlerInput struct {
	Query       string
	Schema      *schema.Metadata
	Connections *connection_config.ConnectionMap
	Prompt      *prompt.Prompt
}

func (h *HandlerInput) args() []string {
	return getArguments(h.Query)
}

type metaquery struct {
	title       string
	description string
	args        []string
	handler     handler
	validator   validator
}

type handler func(input *HandlerInput) error

var metaQueryHandlers map[string]metaquery

func init() {
	metaQueryHandlers = map[string]metaquery{
		cmdExit: {
			title:       cmdExit,
			handler:     doExit,
			validator:   noArgs,
			description: "Exit from steampipe terminal",
		},
		cmdQuit: {
			title:       cmdQuit,
			handler:     doExit,
			validator:   noArgs,
			description: "Exit from steampipe terminal",
		},
		cmdTableList: {
			title:       cmdTableList,
			handler:     listTables,
			validator:   atMostNArgs(1),
			description: "List or describe tables",
		},
		cmdSeparator: {
			title:       cmdSeparator,
			handler:     setViperConfigFromArg(constants.ArgSeparator),
			validator:   exactlyNArgs(1),
			description: "Set csv output separator",
		},
		cmdHeaders: {
			title:       "headers",
			handler:     setHeader,
			validator:   booleanValidator(cmdHeaders, allowedArgValues(false, constants.ValOn, constants.ValOff)),
			description: "Enable or disable column headers",
			args:        []string{constants.ValOn, constants.ValOff},
		},
		cmdMulti: {
			title:       "multi-line",
			handler:     setMultiLine,
			validator:   booleanValidator(cmdMulti, allowedArgValues(false, constants.ValOn, constants.ValOff)),
			description: "Enable or disable multiline mode",
			args:        []string{constants.ValOn, constants.ValOff},
		},
		cmdTiming: {
			title:       "timing",
			handler:     setTiming,
			validator:   booleanValidator(cmdTiming, allowedArgValues(false, constants.ValOn, constants.ValOff)),
			description: "Enable or disable query execution timing",
			args:        []string{constants.ValOn, constants.ValOff},
		},
		cmdOutput: {
			title:       cmdOutput,
			handler:     setViperConfigFromArg(constants.ArgOutput),
			validator:   composeValidator(exactlyNArgs(1), allowedArgValues(false, constants.ValJSON, constants.ValCSV, constants.ValTable)),
			description: "Set output format",
			args:        []string{constants.ValJSON, constants.ValCSV, constants.ValTable},
		},
		cmdInspect: {
			title:       cmdInspect,
			handler:     inspect,
			validator:   atMostNArgs(1),
			description: "View connections, tables & column information",
		},
		cmdConnections: {
			title:       cmdConnections,
			handler:     listConnections,
			validator:   noArgs,
			description: "List active connections",
		},
		cmdClear: {
			title:       cmdClear,
			handler:     clearScreen,
			validator:   noArgs,
			description: "Clear the console",
		},
	}
}

// Handle :: handle metaquery.
func Handle(input *HandlerInput) error {
	input.Query = strings.TrimSuffix(input.Query, ";")
	var s = strings.Fields(input.Query)

	var handlerFunction handler
	metaQueryObj, found := metaQueryHandlers[s[0]]
	if !found {
		return fmt.Errorf("not sure how to handle '%s'", s[0])
	}

	handlerFunction = metaQueryObj.handler
	return handlerFunction(input)
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
	// this get's caught at the handleExit function in steampipe/osquery_client/interactive.go
	panic(utils.InteractiveExitStatus{Restart: false})
}

// list all the tables in the schema
func listTables(input *HandlerInput) error {

	if len(input.args()) == 0 {
		schemas := input.Schema.GetSchemas()
		for _, schema := range schemas {
			fmt.Printf(" ==> %s\n", schema)
			inspectConnection(schema, input)
		}

		fmt.Printf(`
To get information about the columns in a table, run '.inspect {connection}.{table}'
	
`)
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
		writeTable(header, rows, true)
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
		// only a connection name
		err := inspectConnection(tableOrConnection, input)

		if err != nil {
			return err
		}

		fmt.Printf(`
To get information about the columns in a table, run '.inspect {connection}.{table}'
	
`)
		return nil
	}

	return inspectTable(split[0], split[1], input)
}

func listConnections(input *HandlerInput) error {
	header := []string{"Connection", "Plugin"}
	rows := [][]string{}

	for _, schema := range input.Schema.GetSchemas() {
		plugin := (*input.Connections)[schema]
		rows = append(rows, []string{schema, plugin.Plugin})
	}

	// sort by connection name
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	writeTable(header, rows, false)

	fmt.Printf(`
To get information about the tables in a connection, run '.inspect {connection}'
To get information about the columns in a table, run '.inspect {connection}.{table}'

`)

	return nil
}

func inspectConnection(connectionName string, input *HandlerInput) error {
	header := []string{"Table", "Description"}
	rows := [][]string{}

	schema, found := input.Schema.Schemas[connectionName]

	if !found {
		return fmt.Errorf("Could not find connection called '%s'", connectionName)
	}

	for _, tableSchema := range schema {
		rows = append(rows, []string{tableSchema.Name, tableSchema.Description})
	}

	// sort by table name
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	writeTable(header, rows, false)

	return nil
}

func clearScreen(input *HandlerInput) error {
	input.Prompt.ClearScreen()
	return nil
}

func inspectTable(connectionName string, tableName string, input *HandlerInput) error {
	header := []string{"Column", "Type", "Description"}
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

	writeTable(header, rows, false)

	return nil
}

func writeTable(header []string, rows [][]string, autoMerge bool) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader(header)
	table.SetBorder(true)
	table.SetAutoMergeCells(autoMerge)
	for _, row := range rows {
		table.Append(row)
	}
	table.Render()
}
