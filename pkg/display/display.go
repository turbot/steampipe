package display

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/karrick/gows"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/query/queryresult"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// ShowOutput displays the output using the proper formatter as applicable
func ShowOutput(ctx context.Context, result *queryresult.Result, opts ...DisplayOption) int {
	rowErrors := 0
	config := newDisplayConfiguration()
	for _, o := range opts {
		o(config)
	}

	var timingResult *queryresult.TimingResult

	outputFormat := cmdconfig.Viper().GetString(constants.ArgOutput)
	switch outputFormat {
	case constants.OutputFormatJSON:
		rowErrors, timingResult = displayJSON(ctx, result)
	case constants.OutputFormatCSV:
		rowErrors, timingResult = displayCSV(ctx, result)
	case constants.OutputFormatLine:
		rowErrors, timingResult = displayLine(ctx, result)
	case constants.OutputFormatTable:
		rowErrors, timingResult = displayTable(ctx, result)
	}

	// show timing
	if config.timing != constants.ArgOff && timingResult != nil {
		str := buildTimingString(timingResult)
		if viper.GetBool(constants.ConfigKeyInteractive) {
			fmt.Println(str)
		} else {
			fmt.Fprintln(os.Stderr, str)
		}
	}
	// return the number of rows that returned errors
	return rowErrors
}

type ShowWrappedTableOptions struct {
	AutoMerge        bool
	HideEmptyColumns bool
	Truncate         bool
	OutputMirror     io.Writer
}

func ShowWrappedTable(headers []string, rows [][]string, opts *ShowWrappedTableOptions) {
	if opts == nil {
		opts = &ShowWrappedTableOptions{}
	}
	t := table.NewWriter()

	t.SetStyle(table.StyleDefault)
	t.Style().Format.Header = text.FormatDefault
	if opts.OutputMirror == nil {
		t.SetOutputMirror(os.Stdout)
	} else {
		t.SetOutputMirror(opts.OutputMirror)
	}

	rowConfig := table.RowConfig{AutoMerge: opts.AutoMerge}
	colConfigs, headerRow := getColumnSettings(headers, rows, opts)

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

func GetMaxCols() int {
	colsAvailable, _, _ := gows.GetWinSize()
	// check if STEAMPIPE_DISPLAY_WIDTH env variable is set
	if viper.IsSet(constants.ArgDisplayWidth) {
		colsAvailable = viper.GetInt(constants.ArgDisplayWidth)
	}
	return colsAvailable
}

// calculate and returns column configuration based on header and row content
func getColumnSettings(headers []string, rows [][]string, opts *ShowWrappedTableOptions) ([]table.ColumnConfig, table.Row) {
	colConfigs := make([]table.ColumnConfig, len(headers))
	headerRow := make(table.Row, len(headers))

	sumOfAllCols := 0

	// account for the spaces around the value of a column and separators
	spaceAccounting := ((len(headers) * 3) + 1)

	for idx, colName := range headers {
		headerRow[idx] = colName

		// get the maximum len of strings in this column
		maxLen := getTerminalColumnsRequiredForString(colName)
		colHasValue := false
		for _, row := range rows {
			colVal := row[idx]
			if !colHasValue && len(colVal) > 0 {
				// the !colHasValue is necessary in the condition,
				// otherwise, even after being set, we will keep
				// evaluating the length
				colHasValue = true
			}

			// get the maximum line length of the value
			colLen := getTerminalColumnsRequiredForString(colVal)
			if colLen > maxLen {
				maxLen = colLen
			}
		}
		colConfigs[idx] = table.ColumnConfig{
			Name:     colName,
			Number:   idx + 1,
			WidthMax: maxLen,
			WidthMin: maxLen,
		}
		if opts.HideEmptyColumns && !colHasValue {
			colConfigs[idx].Hidden = true
		}
		sumOfAllCols += maxLen
	}

	// now that all columns are set to the widths that they need,
	// set the last one to occupy as much as is available - no more - no less
	sumOfRest := sumOfAllCols - colConfigs[len(colConfigs)-1].WidthMax
	// get the max cols width
	maxCols := GetMaxCols()
	if sumOfAllCols > maxCols {
		colConfigs[len(colConfigs)-1].WidthMax = (maxCols - sumOfRest - spaceAccounting)
		colConfigs[len(colConfigs)-1].WidthMin = (maxCols - sumOfRest - spaceAccounting)
		if opts.Truncate {
			colConfigs[len(colConfigs)-1].WidthMaxEnforcer = helpers.TruncateString
		}
	}

	return colConfigs, headerRow
}

// getTerminalColumnsRequiredForString returns the length of the longest line in the string
func getTerminalColumnsRequiredForString(str string) int {
	colsRequired := 0
	scanner := bufio.NewScanner(bytes.NewBufferString(str))
	for scanner.Scan() {
		line := scanner.Text()
		runeCount := utf8.RuneCountInString(line)
		if runeCount > colsRequired {
			colsRequired = runeCount
		}
	}
	return colsRequired
}

type jsonOutput struct {
	Rows     []map[string]interface{}  `json:"rows"`
	Metadata *queryresult.TimingResult `json:"metadata"`
}

func newJSONOutput() *jsonOutput {
	return &jsonOutput{
		Rows: make([]map[string]interface{}, 0),
	}

}

func displayJSON(ctx context.Context, result *queryresult.Result) (int, *queryresult.TimingResult) {
	rowErrors := 0
	jsonOutput := newJSONOutput()

	// define function to add each row to the JSON output
	rowFunc := func(row []interface{}, result *queryresult.Result) {
		record := map[string]interface{}{}
		for idx, col := range result.Cols {
			value, _ := ParseJSONOutputColumnValue(row[idx], col)
			record[col.Name] = value
		}
		jsonOutput.Rows = append(jsonOutput.Rows, record)
	}

	// call this function for each row
	count, err := iterateResults(result, rowFunc)
	if err != nil {
		error_helpers.ShowError(ctx, err)
		rowErrors++
		return rowErrors, nil
	}

	// now we have iterated the rows, get the timing
	jsonOutput.Metadata = getTiming(result, count)

	// display the JSON
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", " ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(jsonOutput); err != nil {
		fmt.Print("Error displaying result as JSON", err)
		return 0, nil
	}
	return rowErrors, jsonOutput.Metadata
}

func displayCSV(ctx context.Context, result *queryresult.Result) (int, *queryresult.TimingResult) {
	rowErrors := 0
	csvWriter := csv.NewWriter(os.Stdout)
	csvWriter.Comma = []rune(cmdconfig.Viper().GetString(constants.ArgSeparator))[0]

	if cmdconfig.Viper().GetBool(constants.ArgHeader) {
		_ = csvWriter.Write(ColumnNames(result.Cols))
	}

	// print the data as it comes
	// define function display each csv row
	rowFunc := func(row []interface{}, result *queryresult.Result) {
		rowAsString, _ := ColumnValuesAsString(row, result.Cols, WithNullString(""))
		_ = csvWriter.Write(rowAsString)
	}

	// call this function for each row
	count, err := iterateResults(result, rowFunc)
	if err != nil {
		error_helpers.ShowError(ctx, err)
		rowErrors++
		return rowErrors, nil
	}

	csvWriter.Flush()
	if csvWriter.Error() != nil {
		error_helpers.ShowErrorWithMessage(ctx, csvWriter.Error(), "unable to print csv")
	}

	// now we have iterated the rows, get the timing
	timingResult := getTiming(result, count)

	return rowErrors, timingResult
}

func displayLine(ctx context.Context, result *queryresult.Result) (int, *queryresult.TimingResult) {

	maxColNameLength, rowErrors := 0, 0
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
	count, err := iterateResults(result, rowFunc)
	if err != nil {
		error_helpers.ShowError(ctx, err)
		rowErrors++
		return rowErrors, nil
	}

	// now we have iterated the rows, get the timing
	timingResult := getTiming(result, count)
	return rowErrors, timingResult
}

func displayTable(ctx context.Context, result *queryresult.Result) (int, *queryresult.TimingResult) {
	rowErrors := 0
	// the buffer to put the output data in
	outbuf := bytes.NewBufferString("")

	// the table
	t := table.NewWriter()
	t.SetOutputMirror(outbuf)
	t.SetStyle(table.StyleDefault)
	t.Style().Format.Header = text.FormatDefault

	var colConfigs []table.ColumnConfig
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
			// trim out non-displayable code-points in string
			// exfept white-spaces
			col = strings.Map(func(r rune) rune {
				if unicode.IsSpace(r) || unicode.IsGraphic(r) {
					// return if this is a white space character
					return r
				}
				return -1
			}, col)
			rowObj = append(rowObj, col)
		}
		t.AppendRow(rowObj)
	}

	// iterate each row, adding each to the table
	count, err := iterateResults(result, rowFunc)
	if err != nil {
		// display the error
		fmt.Println()
		error_helpers.ShowError(ctx, err)
		rowErrors++
		fmt.Println()
	}
	// write out the table to the buffer
	t.Render()

	// page out the table
	ShowPaged(ctx, outbuf.String())

	// now we have iterated the rows, get the timing
	timingResult := getTiming(result, count)

	return rowErrors, timingResult
}

func getTiming(result *queryresult.Result, count int) *queryresult.TimingResult {
	// now we have iterated the rows, get the timing
	timingResult := <-result.TimingResult
	// set rows returned
	timingResult.RowsReturned = int64(count)
	return timingResult
}

func buildTimingString(timingResult *queryresult.TimingResult) string {
	var sb strings.Builder
	// large numbers should be formatted with commas
	p := message.NewPrinter(language.English)

	sb.WriteString(fmt.Sprintf("\nTime: %s.", getDurationString(timingResult.DurationMs, p)))
	sb.WriteString(p.Sprintf(" Rows returned: %d.", timingResult.RowsReturned))
	sb.WriteString(" Rows fetched: ")
	if timingResult.RowsFetched == 0 {
		sb.WriteString("0")
	} else {

		// calculate the number of cached rows fetched
		var cachedRowsFetched int64
		for _, scan := range timingResult.Scans {
			if scan.CacheHit {
				cachedRowsFetched += scan.RowsFetched
			}
		}
		sb.WriteString(p.Sprintf("%d", timingResult.RowsFetched))

		if timingResult.RowsFetched == cachedRowsFetched {
			sb.WriteString(" (cached)")
		} else if cachedRowsFetched > 0 {
			sb.WriteString(p.Sprintf(" (%d cached)", cachedRowsFetched))
		}
	}

	sb.WriteString(p.Sprintf(". Hydrate calls: %d.", timingResult.HydrateCalls))

	if viper.GetString(constants.ArgTiming) == constants.ArgVerbose && len(timingResult.Scans) > 0 {
		if err := getVerboseTimingString(&sb, p, timingResult.Scans); err != nil {
			log.Printf("[WARN] Error getting verbose timing: %v", err)
		}
	}

	return sb.String()
}

func getDurationString(durationMs int64, p *message.Printer) string {
	if durationMs < 500 {
		return p.Sprintf("%dms", durationMs)
	} else {
		seconds := float64(durationMs / 1000)
		return p.Sprintf("%.1fs", seconds)
	}
}

func getVerboseTimingString(sb *strings.Builder, p *message.Printer, scans []*queryresult.ScanMetadataRow) error {
	sb.WriteString("\n\nScans:\n")
	for i, scan := range scans {
		cacheString := ""
		if scan.CacheHit {
			cacheString = " (cached)"
		}
		qualsString := ""
		if len(scan.Quals) > 0 {
			qualsJson, err := json.Marshal(scan.Quals)
			if err != nil {
				return err
			}
			qualsString = fmt.Sprintf(" Quals: %s.", string(qualsJson))
		}
		limitString := ""
		if scan.Limit != nil {
			limitString = fmt.Sprintf(" Limit: %d.", *scan.Limit)
		}

		timeString := getDurationString(scan.DurationMs, p)
		rowsFetchedString := p.Sprintf("%d", scan.RowsFetched)

		sb.WriteString(fmt.Sprintf("  %d) Table: %s. Connection: %s. Time: %s. Rows fetched: %s%s. Hydrate calls: %d.%s%s\n", i+1, scan.Table, scan.Connection, timeString, rowsFetchedString, cacheString, scan.HydrateCalls, qualsString, limitString))
	}
	return nil
}

type displayResultsFunc func(row []interface{}, result *queryresult.Result)

// call func displayResult for each row of results
func iterateResults(result *queryresult.Result, displayResult displayResultsFunc) (int, error) {
	count := 0
	for row := range *result.RowChan {
		if row == nil {
			return count, nil
		}
		if row.Error != nil {
			return count, row.Error
		}
		displayResult(row.Data, result)
		count++
	}
	// we will not get here
	return count, nil
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
