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

	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/utils"
)

// ShowOutput :: displays the output using the proper formatter as applicable
func ShowOutput(result *db.QueryResult) {
	if cmdconfig.Viper().Get(constants.ArgOutput) == constants.ArgJSON {
		displayJSON(result)
	} else if cmdconfig.Viper().Get(constants.ArgOutput) == constants.ArgCSV {
		displayCSV(result)
	} else if cmdconfig.Viper().Get(constants.ArgOutput) == constants.ArgLine {
		displayLine(result)
	} else {
		// default
		displayTable(result)
	}
}

func displayLine(result *db.QueryResult) {
	colNames := ColumnNames(result.ColTypes)
	maxColNameLength := 0
	for _, colName := range colNames {
		thisLength := utf8.RuneCountInString(colName)
		if thisLength > maxColNameLength {
			maxColNameLength = thisLength
		}
	}
	lineFormat := fmt.Sprintf("%%-%ds | %%s\n", maxColNameLength)
	itemIdx := 0
	for item := range *result.RowChan {
		if itemIdx != 0 {
			fmt.Println()
		}
		recordAsString, _ := ColumnValuesAsString(item, result.ColTypes)
		fmt.Printf("-[ RECORD %-2d ]%s\n", (itemIdx + 1), strings.Repeat("-", 75))
		for idx, column := range recordAsString {
			fmt.Printf(lineFormat, colNames[idx], column)
		}
		itemIdx++
	}
}

func displayJSON(result *db.QueryResult) {
	var jsonOutput []map[string]interface{}
	for item := range *result.RowChan {
		record := map[string]interface{}{}
		for idx, colType := range result.ColTypes {
			value, _ := ParseJSONOutputColumnValue(item[idx], colType)
			record[colType.Name()] = value
		}
		jsonOutput = append(jsonOutput, record)
	}
	data, err := json.MarshalIndent(jsonOutput, "", " ")
	if err != nil {
		fmt.Print("Error displaying result as JSON", err)
		return
	}
	fmt.Printf("%s\n", string(data))
}

func displayCSV(result *db.QueryResult) {
	csvWriter := csv.NewWriter(os.Stdout)
	csvWriter.Comma = []rune(cmdconfig.Viper().GetString(constants.ArgSeparator))[0]

	// TODO handle errors

	if cmdconfig.Viper().GetBool(constants.ArgHeader) {
		_ = csvWriter.Write(ColumnNames(result.ColTypes))
	}

	// print the data as it comes
	for row := range *result.RowChan {
		rowAsString, _ := ColumnValuesAsString(row, result.ColTypes)
		_ = csvWriter.Write(rowAsString)
	}
	csvWriter.Flush()
	if csvWriter.Error() != nil {
		utils.ShowErrorWithMessage(csvWriter.Error(), "unable to print csv")
	}
}

func displayTable(result *db.QueryResult) {
	// the buffer to put the output data in
	outbuf := bytes.NewBufferString("")

	// the table
	t := table.NewWriter()
	t.SetOutputMirror(outbuf)
	t.SetStyle(table.StyleDefault)
	t.Style().Format.Header = text.FormatLower

	colConfigs := []table.ColumnConfig{}
	headers := table.Row{}

	for idx, column := range result.ColTypes {
		headers = append(headers, column.Name())
		colConfigs = append(colConfigs, table.ColumnConfig{
			Name:     column.Name(),
			Number:   idx + 1,
			WidthMax: constants.MaxColumnWidth,
		})
	}

	t.SetColumnConfigs(colConfigs)
	t.AppendHeader(headers)

	for {
		row := <-(*result.RowChan)
		if row == nil {
			break
		}
		// TODO how to handle error
		rowAsString, _ := ColumnValuesAsString(row, result.ColTypes)
		rowObj := table.Row{}
		for _, col := range rowAsString {
			rowObj = append(rowObj, col)
		}
		t.AppendRow(rowObj)
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
