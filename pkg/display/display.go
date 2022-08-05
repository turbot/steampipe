package display

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/turbot/go-kit/helpers"
	"io"
	"os"
	"os/exec"
	"strings"
	"unicode/utf8"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/karrick/gows"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/query/queryresult"
	"github.com/turbot/steampipe/pkg/utils"
)

// ShowOutput displays the output using the proper formatter as applicable
func ShowOutput(ctx context.Context, result *queryresult.Result) {
	output := cmdconfig.Viper().GetString(constants.ArgOutput)
	if output == constants.OutputFormatJSON {
		displayJSON(ctx, result)
	} else if output == constants.OutputFormatCSV {
		displayCSV(ctx, result)
	} else if output == constants.OutputFormatLine {
		displayLine(ctx, result)
	} else {
		// default
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
	rowFunc := func(row []interface{}, result *queryresult.Result) {
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
		utils.ShowError(ctx, err)
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
		for idx, colType := range result.ColTypes {
			value, _ := ParseJSONOutputColumnValue(row[idx], colType)
			record[colType.Name()] = value
		}
		jsonOutput = append(jsonOutput, record)
	}

	// call this function for each row
	if err := iterateResults(result, rowFunc); err != nil {
		utils.ShowError(ctx, err)
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
	fmt.Println()
}

func displayCSV(ctx context.Context, result *queryresult.Result) {
	csvWriter := csv.NewWriter(os.Stdout)
	csvWriter.Comma = []rune(cmdconfig.Viper().GetString(constants.ArgSeparator))[0]

	if cmdconfig.Viper().GetBool(constants.ArgHeader) {
		_ = csvWriter.Write(ColumnNames(result.ColTypes))
	}

	// print the data as it comes
	// define function display each csv row
	rowFunc := func(row []interface{}, result *queryresult.Result) {
		rowAsString, _ := ColumnValuesAsString(row, result.ColTypes)
		_ = csvWriter.Write(rowAsString)
	}

	// call this function for each row
	if err := iterateResults(result, rowFunc); err != nil {
		utils.ShowError(ctx, err)
		return
	}

	csvWriter.Flush()
	if csvWriter.Error() != nil {
		utils.ShowErrorWithMessage(ctx, csvWriter.Error(), "unable to print csv")
	}
}

func displayTable(ctx context.Context, result *queryresult.Result) {

	var colConfigs []table.ColumnConfig
	headers := make(table.Row, len(result.ColTypes))

	for idx, column := range result.ColTypes {
		headers[idx] = column.Name()
		colConfigs = append(colConfigs, table.ColumnConfig{
			Name:     column.Name(),
			Number:   idx + 1,
			WidthMax: constants.MaxColumnWidth,
			// use truncation to enforce max width
			WidthMaxEnforcer: helpers.TruncateString,
		})
	}

	// define a function to execute for each row
	rowFunc := func(t table.Writer, row []interface{}, result *queryresult.Result) {
		rowAsString, _ := ColumnValuesAsString(row, result.ColTypes)
		rowObj := table.Row{}
		for _, col := range rowAsString {
			rowObj = append(rowObj, col)
		}
		t.AppendRow(rowObj)
	}

	if !viper.GetBool(constants.ArgHeader) {
		headers = nil
	}
	pageSize := 9

	// render first page into a buffer
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	allDone, headerTable, err := renderTablePage(colConfigs, headers, result, rowFunc, pageSize, writer, nil)

	// TODO determine if we need a pager

	// assuming we do,
	cmd := exec.Command("less", "-SRXF")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		utils.ShowErrorWithMessage(ctx, err, "could not display results")
		return
	}

	// write first page
	stdin.Write(b.Bytes())

	// now render further pages async if necessary
	if !allDone {
		// update column configs to set the max widths as the actual column widths of the header table
		// to ensure we truncate long results
		for idx, columnWidth := range headerTable.MaxColumnLengths() {
			colConfigs[idx].WidthMax = columnWidth
		}

		go func() {
			defer stdin.Close()

			for !allDone {
				// pass nil for header
				allDone, _, err = renderTablePage(colConfigs, nil, result, rowFunc, pageSize, stdin, headerTable)

				if err != nil {
					// display the error
					fmt.Println()
					utils.ShowError(ctx, err)
					fmt.Println()
					break
				}
			}

		}()
	}

	// run the command - it will block until the pager is exited
	err = cmd.Run()
	if err != nil {
		utils.ShowErrorWithMessage(ctx, err, "could not display results")
		return
	}

	// if timer is turned on
	if cmdconfig.Viper().GetBool(constants.ArgTiming) {
		displayTiming(result)
	}
}

func renderTablePage(colConfigs []table.ColumnConfig, headers table.Row, result *queryresult.Result, rowFunc displayResultsFunc2, pageSize int, writer io.Writer, headerTable table.Writer) (bool, table.Writer, error) {
	// the buffer to put the output data in
	outbuff := bytes.NewBufferString("")

	// the table
	t := createTable(outbuff, colConfigs, headers, headerTable)

	// iterate each row, adding each to the table

	allDone, err := iterateResults2(t, result, rowFunc, pageSize)
	if err != nil {
		return false, nil, err

	}

	// if a header table was passed, this is not the first table page
	// do not render the top border
	if headerTable != nil {
		t.HideTopBorder(true)
	}
	// if we are not complete, hide the bottom border
	if !allDone {
		t.HideBottomBorder(true)
	}

	// write out the table to the buffer
	t.Render()

	tableStr := outbuff.String()
	io.WriteString(writer, tableStr)

	return allDone, t, nil
}

func createTable(outbuf *bytes.Buffer, colConfigs []table.ColumnConfig, headers table.Row, headerTable table.Writer) table.Writer {
	t := table.NewWriter()
	t.SetOutputMirror(outbuf)
	t.SetStyle(table.StyleDefault)
	t.Style().Format.Header = text.FormatDefault
	t.SetColumnConfigs(colConfigs)
	if headers != nil {
		t.AppendHeader(headers)
	}
	// if an headerTable was provided, use it to set the column widths
	if headerTable != nil {
		t.SetMaxColumnLengths(headerTable.MaxColumnLengths())

	}
	return t
}

func filePager(ctx context.Context, filename string) {
	//cmdStr := fmt.Sprintf("cat %s | less -SRXF", filename)
	//cmd := exec.Command("bash", "-c", cmdStr)
	//cmd.Dir = filepath.Dir(filename)

	cmd := exec.Command("less", "-SRXF", filename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	//cmd.Stdin = strings.NewReader(content)
	// run the command - it will block until the pager is exited
	err := cmd.Run()
	if err != nil {
		utils.ShowErrorWithMessage(ctx, err, "could not display results")
	}
}

func displayTiming(result *queryresult.Result) {
	timingResult := <-result.TimingResult
	var sb strings.Builder

	milliseconds := float64(timingResult.Duration.Microseconds()) / 1000
	sb.WriteString(fmt.Sprintf("\nTime: %.1fms.", milliseconds))
	if timingMetadata := timingResult.Metadata; timingMetadata != nil {
		totalRows := timingMetadata.RowsFetched + timingMetadata.CachedRowsFetched
		sb.WriteString(" Rows fetched: ")
		if totalRows == 0 {
			sb.WriteString("0")
		} else {
			if totalRows > 0 {
				sb.WriteString(fmt.Sprintf("%d", timingMetadata.RowsFetched+timingMetadata.CachedRowsFetched))
			}
			if timingMetadata.CachedRowsFetched > 0 {
				if timingMetadata.RowsFetched == 0 {
					sb.WriteString(" (cached)")
				} else {
					sb.WriteString(fmt.Sprintf(" (%d cached)", timingMetadata.CachedRowsFetched))
				}
			}
		}
		sb.WriteString(fmt.Sprintf(". Hydrate calls: %d.", timingMetadata.HydrateCalls))
	}

	fmt.Println(sb.String())
}

type displayResultsFunc func([]interface{}, *queryresult.Result)
type displayResultsFunc2 func(table.Writer, []interface{}, *queryresult.Result)

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

//var bufferedRow *queryresult.RowResult

func iterateResults2(t table.Writer, result *queryresult.Result, displayResult displayResultsFunc2, count int) (allDone bool, err error) {
	i := 0
	//if bufferedRow != nil {
	//	displayResult(t, bufferedRow.Data, result)
	//	bufferedRow = nil
	//}
	for row := range *result.RowChan {
		// shouldn't happen
		if row == nil {
			return true, nil
		}
		if row.Error != nil {
			return true, row.Error
		}
		i++
		displayResult(t, row.Data, result)
		if i == count {
			// buffer the last row (to ensure that we never end up trying to render an empty table)
			//bufferedRow = row
			return false, nil
		}

	}

	return true, nil
}
