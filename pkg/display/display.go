package display

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/turbot/steampipe/pkg/error_helpers"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/karrick/gows"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/query/queryresult"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// ShowOutput displays the output using the proper formatter as applicable
func ShowOutput(ctx context.Context, result *queryresult.Result) {
	switch cmdconfig.Viper().GetString(constants.ArgOutput) {
	case constants.OutputFormatJSON:
		displayJSON(ctx, result)
	case constants.OutputFormatCSV:
		displayCSV(ctx, result)
	case constants.OutputFormatLine:
		displayLine(ctx, result)
	case constants.OutputFormatTable:
		displayTable(ctx, result)
	}
}

func ShowWrappedTable(headers []string, rows [][]string, autoMerge bool) {
	t := table.NewWriter()

	t.SetStyle(table.StyleDefault)
	t.Style().Format.Header = text.FormatDefault
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

func displayLine(ctx context.Context, result *queryresult.Result) {

	maxColNameLength := 0
	for _, col := range result.Cols {
		thisLength := utf8.RuneCountInString(col.Name)
		if thisLength > maxColNameLength {
			maxColNameLength = thisLength
		}
	}
	itemIdx := 0

	// define a function to display each row
	rowFunc := func(row []interface{}, result *queryresult.Result) {
		recordAsString, _ := ColumnValuesAsString(row, result.Cols)
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
				fmt.Printf(lineFormat, result.Cols[idx].Name, lines[0])
			} else {
				for lineIdx, line := range lines {
					if lineIdx == 0 {
						// the first line
						fmt.Printf(multiLineFormat, result.Cols[idx].Name, line)
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
		error_helpers.ShowError(ctx, err)
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

func displayJSON(ctx context.Context, result *queryresult.Result) {
	var jsonOutput []map[string]interface{}

	// define function to add each row to the JSON output
	rowFunc := func(row []interface{}, result *queryresult.Result) {
		record := map[string]interface{}{}
		for idx, col := range result.Cols {
			value, _ := ParseJSONOutputColumnValue(row[idx], col)
			record[col.Name] = value
		}
		jsonOutput = append(jsonOutput, record)
	}

	// call this function for each row
	if err := iterateResults(result, rowFunc); err != nil {
		error_helpers.ShowError(ctx, err)
		return
	}
	// display the JSON
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", " ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(jsonOutput); err != nil {
		fmt.Print("Error displaying result as JSON", err)
		return
	}
}

func displayCSV(ctx context.Context, result *queryresult.Result) {
	csvWriter := csv.NewWriter(os.Stdout)
	csvWriter.Comma = []rune(cmdconfig.Viper().GetString(constants.ArgSeparator))[0]

	if cmdconfig.Viper().GetBool(constants.ArgHeader) {
		_ = csvWriter.Write(ColumnNames(result.Cols))
	}

	// print the data as it comes
	// define function display each csv row
	rowFunc := func(row []interface{}, result *queryresult.Result) {
		rowAsString, _ := ColumnValuesAsString(row, result.Cols)
		_ = csvWriter.Write(rowAsString)
	}

	// call this function for each row
	if err := iterateResults(result, rowFunc); err != nil {
		error_helpers.ShowError(ctx, err)
		return
	}

	csvWriter.Flush()
	if csvWriter.Error() != nil {
		error_helpers.ShowErrorWithMessage(ctx, csvWriter.Error(), "unable to print csv")
	}
}

func displayTable(ctx context.Context, result *queryresult.Result) {
	// the buffer to put the output data in
	outbuf := bytes.NewBufferString("")

	// the table
	t := table.NewWriter()
	t.SetOutputMirror(outbuf)
	t.SetStyle(table.StyleDefault)
	t.Style().Format.Header = text.FormatDefault

	colConfigs := []table.ColumnConfig{}
	headers := make(table.Row, len(result.Cols))

	for idx, column := range result.Cols {
		headers[idx] = column.Name
		colConfigs = append(colConfigs, table.ColumnConfig{
			Name:     column.Name,
			Number:   idx + 1,
			WidthMax: constants.MaxColumnWidth,
		})
	}

	t.SetColumnConfigs(colConfigs)
	if viper.GetBool(constants.ArgHeader) {
		t.AppendHeader(headers)
	}

	// define a function to execute for each row
	rowFunc := func(row []interface{}, result *queryresult.Result) {
		rowAsString, _ := ColumnValuesAsString(row, result.Cols)
		rowObj := table.Row{}
		for _, col := range rowAsString {
			rowObj = append(rowObj, col)
		}
		t.AppendRow(rowObj)
	}

	// iterate each row, adding each to the table
	err := iterateResults(result, rowFunc)
	if err != nil {
		// display the error
		fmt.Println()
		error_helpers.ShowError(ctx, err)
		fmt.Println()
	}
	// write out the table to the buffer
	t.Render()

	// page out the table
	ShowPaged(ctx, outbuf.String())

	// if timer is turned on
	if cmdconfig.Viper().GetBool(constants.ArgTiming) {
		fmt.Println(buildTimingString(result))
	}
}

func buildTimingString(result *queryresult.Result) string {
	timingResult := <-result.TimingResult
	if timingResult == nil {
		return ""
	}
	var sb strings.Builder
	// large numbers should be formatted with commas
	p := message.NewPrinter(language.English)

	milliseconds := float64(timingResult.Duration.Microseconds()) / 1000
	seconds := timingResult.Duration.Seconds()
	if seconds < 0.5 {
		sb.WriteString(p.Sprintf("\nTime: %dms.", int64(milliseconds)))
	} else {
		sb.WriteString(p.Sprintf("\nTime: %.1fs.", seconds))
	}

	if timingMetadata := timingResult.Metadata; timingMetadata != nil {
		totalRows := timingMetadata.RowsFetched + timingMetadata.CachedRowsFetched
		sb.WriteString(" Rows fetched: ")
		if totalRows == 0 {
			sb.WriteString("0")
		} else {
			if totalRows > 0 {
				sb.WriteString(p.Sprintf("%d", timingMetadata.RowsFetched+timingMetadata.CachedRowsFetched))
			}
			if timingMetadata.CachedRowsFetched > 0 {
				if timingMetadata.RowsFetched == 0 {
					sb.WriteString(" (cached)")
				} else {
					sb.WriteString(p.Sprintf(" (%d cached)", timingMetadata.CachedRowsFetched))
				}
			}
		}
		sb.WriteString(p.Sprintf(". Hydrate calls: %d.", timingMetadata.HydrateCalls))
	}

	return sb.String()
}

type displayResultsFunc func(row []interface{}, result *queryresult.Result)

// call func displayResult for each row of results
func iterateResults(result *queryresult.Result, displayResult displayResultsFunc) error {
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

// DisplayErrorTiming shows the time taken for the query to fail
func DisplayErrorTiming(t time.Time) {
	elapsed := time.Since(t)
	var sb strings.Builder
	// large numbers should be formatted with commas
	p := message.NewPrinter(language.English)

	milliseconds := float64(elapsed.Microseconds()) / 1000
	seconds := elapsed.Seconds()
	if seconds < 0.5 {
		sb.WriteString(p.Sprintf("\nTime: %dms.", int64(milliseconds)))
	} else {
		sb.WriteString(p.Sprintf("\nTime: %.1fs.", seconds))
	}
	fmt.Println(sb.String())
}
