package display

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/utils"

	"github.com/olekukonko/tablewriter"
)

// ShowOutput :: displays the output using the proper formatter as applicable
func ShowOutput(result *db.QueryResult) {
	if cmdconfig.Viper().Get(constants.ArgOutput) == constants.ValJSON {
		displayJSON(result)
	} else if cmdconfig.Viper().Get(constants.ArgOutput) == constants.ValCSV {
		displayCSV(result)
	} else {
		displayTable(result)
	}
}

func displayJSON(result *db.QueryResult) {
	var jsonOutput []map[string]interface{}
	for item := range *result.RowChan {
		record := map[string]interface{}{}
		for idx, col := range result.ColTypes {
			record[col.Name()] = item[idx]
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

	if cmdconfig.Viper().GetBool(constants.ArgHeader) {
		csvWriter.Write(ColumnNames(result.ColTypes))
	}

	// print the data as it comes
	for row := range *result.RowChan {
		// TODO how to handle error
		rowAsString, _ := ColumnValuesAsString(row, result.ColTypes)
		csvWriter.Write(rowAsString)
	}
	csvWriter.Flush()
	if csvWriter.Error() != nil {
		utils.ShowErrorWithMessage(csvWriter.Error(), "unable to print csv")
	}
}

func displayTable(result *db.QueryResult) {
	// the buffer to put the output data in
	outbuf := bytes.NewBufferString("")
	table := tablewriter.NewWriter(outbuf)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	if cmdconfig.Viper().GetBool(constants.ArgHeader) {
		table.SetHeader(ColumnNames(result.ColTypes))
	}
	table.SetBorder(true)

	for {
		row := <-(*result.RowChan)
		if row == nil {
			break
		}
		// TODO how to handle error
		rowAsString, _ := ColumnValuesAsString(row, result.ColTypes)
		table.Append(rowAsString)
	}

	table.SetAutoWrapText(false)

	// write out the table to the buffer
	table.Render()

	// if timer is turned on
	if cmdconfig.Viper().GetBool(constants.ArgTimer) {
		// put in the time information in the buffer
		outbuf.WriteString(fmt.Sprintf("\nTime: %v\n", <-result.Duration))
	}

	// spit it out!
	displayPaged(outbuf.String())
}
