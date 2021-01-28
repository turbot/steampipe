package metaquery

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/karrick/gows"
	"github.com/turbot/go-kit/helpers"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/connection_config"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/utils"
)

var CommonCmds = []string{constants.CmdHelp, constants.CmdInspect, constants.CmdExit}

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

// help
func doHelp(input *HandlerInput) error {
	commonCmdRows := getMetaQueryHelpRows(CommonCmds, false)
	var advanceCmds []string
	for cmd, _ := range metaQueryDefinitions {
		if !helpers.StringSliceContains(CommonCmds, cmd) {
			advanceCmds = append(advanceCmds, cmd)
		}
	}
	advanceCmdRows := getMetaQueryHelpRows(advanceCmds, true)
	// print out
	fmt.Printf("Welcome to Steampipe shell.\n\nTo start, simply enter your SQL query at the prompt:\n\n  select * from aws_iam_user\n\nCommon commands:\n\n%s\n\nAdvanced commands:\n\n%s\n",
		buildTable(commonCmdRows, true),
		buildTable(advanceCmdRows, true))
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

func writeTable(headers []string, rows [][]string, autoMerge bool) {
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.SetOutputMirror(os.Stdout)

	rowConfig := table.RowConfig{AutoMerge: autoMerge}
	colConfigs, headerRow := getColumnSettings(headers, rows)

	t.SetColumnConfigs(colConfigs)
	t.AppendHeader(headerRow)

	for _, row := range rows {
		rowObj := table.Row{}
		for _, col := range row {
			rowObj = append(rowObj, col)
		}
		t.AppendRow(rowObj, rowConfig)
	}
	t.Render()
}

// calculate and returns column configuration based on header and row content
func getColumnSettings(headers []string, rows [][]string) ([]table.ColumnConfig, table.Row) {
	maxCols, _, _ := gows.GetWinSize()
	colConfigs := make([]table.ColumnConfig, len(headers))
	headerRow := make(table.Row, len(headers))

	sumOfAllCols := 0

	// account for the spaces around the value of a column and separators
	spaceAccounting := ((len(headers) * 3) + 1)

	for idx, colName := range headers {
		headerRow[idx] = colName

		// get the maximum len of strings in this column
		maxLen := 0
		for _, row := range rows {
			colVal := row[idx]
			if len(colVal) > maxLen {
				maxLen = len(colVal)
			}
			if len(colName) > maxLen {
				maxLen = len(colName)
			}
		}
		colConfigs[idx] = table.ColumnConfig{
			Name:     colName,
			Number:   idx + 1,
			WidthMax: maxLen,
			WidthMin: maxLen,
		}
		sumOfAllCols += maxLen
	}

	// now that all columns are set to the widths that they need,
	// set the last one to occupy as much as is available - no more - no less
	sumOfRest := 0
	for idx, c := range colConfigs {
		if idx == len(colConfigs)-1 {
			continue
		}
		sumOfRest += c.WidthMax
	}
	colConfigs[len(colConfigs)-1].WidthMax = (maxCols - sumOfRest - spaceAccounting)
	colConfigs[len(colConfigs)-1].WidthMin = (maxCols - sumOfRest - spaceAccounting)

	return colConfigs, headerRow
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
