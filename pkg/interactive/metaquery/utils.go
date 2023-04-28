package metaquery

import (
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/turbot/go-kit/helpers"
)

// IsMetaQuery returns whether the query is a metaquery
func IsMetaQuery(query string) bool {
	if !strings.HasPrefix(query, ".") {
		return false
	}

	// try to look for the validator
	cmd, _ := getCmdAndArgs(query)
	_, foundHandler := metaQueryDefinitions[cmd]

	return foundHandler
}

// extract the command and arguments from the query string
func getCmdAndArgs(query string) (string, []string) {
	query = strings.TrimSuffix(query, ";")
	split := helpers.SplitByWhitespace(query)
	cmd := split[0]
	args := []string{}
	if len(split) > 1 {
		args = split[1:]
	}
	return cmd, args
}

// extract the arguments from the query string
func getArguments(query string) []string {
	_, args := getCmdAndArgs(query)
	return args
}

// build a table from the provided row data
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
