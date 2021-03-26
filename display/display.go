package display

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/karrick/gows"
	"github.com/spf13/viper"

	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/definitions/results"
	"github.com/turbot/steampipe/utils"
)

// ShowOutput :: displays the output using the proper formatter as applicable
func ShowOutput(result *results.QueryResult) {
	output := cmdconfig.Viper().GetString(constants.ArgOutput)
	if output == constants.ArgJSON {
		displayJSON(result)
	} else if output == constants.ArgCSV {
		displayCSV(result)
	} else if output == constants.ArgLine {
		displayLine(result)
	} else {
		// default
		displayTable(result)
	}
}

func ShowWrappedTable(headers []string, rows [][]string, autoMerge bool) {
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
	sumOfRest := sumOfAllCols - colConfigs[len(colConfigs)-1].WidthMax

	if sumOfAllCols > maxCols {
		colConfigs[len(colConfigs)-1].WidthMax = (maxCols - sumOfRest - spaceAccounting)
		colConfigs[len(colConfigs)-1].WidthMin = (maxCols - sumOfRest - spaceAccounting)
	}

	return colConfigs, headerRow
}

func displayLine(result *results.QueryResult) {
	colNames := ColumnNames(result.ColTypes)
	maxColNameLength := 0
	for _, colName := range colNames {
		thisLength := utf8.RuneCountInString(colName)
		if thisLength > maxColNameLength {
			maxColNameLength = thisLength
		}
	}
	itemIdx := 0

	// define a function to display each row
	rowFunc := func(row []interface{}, result *results.QueryResult) {
		recordAsString, _ := ColumnValuesAsString(row, result.ColTypes)
		requiredTerminalColumnsForValuesOfRecord := 0
		for _, colValue := range recordAsString {
			colRequired := getTerminalColumnsRequiredForString(colValue)
			if requiredTerminalColumnsForValuesOfRecord < colRequired {
				requiredTerminalColumnsForValuesOfRecord = colRequired
			}
		}

		lineFormat := fmt.Sprintf("%%-%ds | %%s\n", maxColNameLength)
		multiLineFormat := fmt.Sprintf("%%-%ds | %%-%ds", maxColNameLength, requiredTerminalColumnsForValuesOfRecord)

		fmt.Printf("-[ RECORD %-2d ]%s\n", (itemIdx + 1), strings.Repeat("-", 75))
		for idx, column := range recordAsString {
			lines := strings.Split(column, "\n")
			if len(lines) == 1 {
				fmt.Printf(lineFormat, colNames[idx], lines[0])
			} else {
				for lineIdx, line := range lines {
					if lineIdx == 0 {
						// the first line
						fmt.Printf(multiLineFormat, colNames[idx], line)
					} else {
						// next lines
						fmt.Printf(multiLineFormat, "", line)
					}

					// is this not the last line of value?
					if lineIdx < len(lines)-1 {
						fmt.Printf(" +\n")
					} else {
						fmt.Printf("\n")
					}

				}
			}
		}
		itemIdx++

	}

	// call this function for each row
	if err := iterateResults(result, rowFunc); err != nil {
		utils.ShowError(err)
		return
	}
}

func getTerminalColumnsRequiredForString(str string) int {
	colsRequired := 0
	for _, line := range strings.Split(str, "\n") {
		if colsRequired < utf8.RuneCountInString(line) {
			colsRequired = utf8.RuneCountInString(line)
		}
	}
	return colsRequired
}

func displayJSON(result *results.QueryResult) {
	var jsonOutput []map[string]interface{}

	// define function to add each row to the JSON output
	rowFunc := func(row []interface{}, result *results.QueryResult) {
		record := map[string]interface{}{}
		for idx, colType := range result.ColTypes {
			value, _ := ParseJSONOutputColumnValue(row[idx], colType)
			record[colType.Name()] = value
		}
		jsonOutput = append(jsonOutput, record)
	}

	// call this function for each row
	if err := iterateResults(result, rowFunc); err != nil {
		utils.ShowError(err)
		return
	}
	// display the JSON
	data, err := json.MarshalIndent(jsonOutput, "", " ")
	if err != nil {
		fmt.Print("Error displaying result as JSON", err)
		return
	}
	fmt.Printf("%s\n", string(data))
}

func displayCSV(result *results.QueryResult) {
	csvWriter := csv.NewWriter(os.Stdout)
	csvWriter.Comma = []rune(cmdconfig.Viper().GetString(constants.ArgSeparator))[0]

	if cmdconfig.Viper().GetBool(constants.ArgHeader) {
		_ = csvWriter.Write(ColumnNames(result.ColTypes))
	}

	// print the data as it comes
	// define function display each csv row
	rowFunc := func(row []interface{}, result *results.QueryResult) {
		rowAsString, _ := ColumnValuesAsString(row, result.ColTypes)
		_ = csvWriter.Write(rowAsString)
	}

	// call this function for each row
	if err := iterateResults(result, rowFunc); err != nil {
		utils.ShowError(err)
		return
	}

	csvWriter.Flush()
	if csvWriter.Error() != nil {
		utils.ShowErrorWithMessage(csvWriter.Error(), "unable to print csv")
	}
}

func displayTable(result *results.QueryResult) {
	// the buffer to put the output data in
	outbuf := bytes.NewBufferString("")

	// the table
	t := table.NewWriter()
	t.SetOutputMirror(outbuf)
	t.SetStyle(table.StyleDefault)
	t.Style().Format.Header = text.FormatDefault

	colConfigs := []table.ColumnConfig{}
	headers := make(table.Row, len(result.ColTypes))

	for idx, column := range result.ColTypes {
		headers[idx] = column.Name()
		colConfigs = append(colConfigs, table.ColumnConfig{
			Name:     column.Name(),
			Number:   idx + 1,
			WidthMax: constants.MaxColumnWidth,
		})
	}

	t.SetColumnConfigs(colConfigs)
	if viper.GetBool(constants.ArgHeader) {
		t.AppendHeader(headers)
	}

	// define a function to execute for each row
	rowFunc := func(row []interface{}, result *results.QueryResult) {
		rowAsString, _ := ColumnValuesAsString(row, result.ColTypes)
		rowObj := table.Row{}
		for _, col := range rowAsString {
			rowObj = append(rowObj, col)
		}
		t.AppendRow(rowObj)
	}

	// iterate each row, adding each to the table
	if err := iterateResults(result, rowFunc); err != nil {
		utils.ShowError(err)
		return
	}

	// write out the table to the buffer
	t.Render()
	// if timer is turned on
	if cmdconfig.Viper().GetBool(constants.ArgTimer) {
		// put in the time information in the buffer
		outbuf.WriteString(fmt.Sprintf("\nTime: %v\n", <-result.Duration))
	}

	// page out the table
	ShowPaged(outbuf.String())
}

type displayResultsFunc func(row []interface{}, result *results.QueryResult)

// call func displayResult for each row of results
func iterateResults(result *results.QueryResult, displayResult displayResultsFunc) error {
	for row := range *result.RowChan {
		if row == nil {
			return nil
		}
		if row.Error != nil {
			return row.Error
		}
		displayResult(row.Data, result)
	}
	// we will not get here
	return nil
}
