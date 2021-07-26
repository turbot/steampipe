package interactive

import (
	"strings"

	"github.com/turbot/go-kit/helpers"
)

type queryCompletionInfo struct {
	Table         string
	EditingTable  bool
	EditingColumn bool
}

func getQueryInfo(text string) *queryCompletionInfo {
	table := getTable(text)
	prevWord := getPreviousWord(text)

	return &queryCompletionInfo{
		Table:         table,
		EditingTable:  isEditingTable(prevWord),
		EditingColumn: isEditingColumn(prevWord, table),
	}
}

func isEditingTable(prevWord string) bool {
	var editingTable = prevWord == "from"
	return editingTable
}

func isEditingColumn(prevWord string, table string) bool {
	var editingColumn = prevWord == "where" || prevWord == "and" || prevWord == "or" && table != ""
	return editingColumn
}

func getTable(text string) string {
	// split on space and remove empty results - they occur if there is a double space
	split := helpers.RemoveFromStringSlice(strings.Split(text, " "), "")

	for idx, word := range split {
		if word == "from" {
			if idx+1 < len(split) {
				return split[idx+1]
			}
		}
	}
	return ""
}

func getPreviousWord(text string) string {
	// create a new document up the previous space
	finalSpace := strings.LastIndex(text, " ")
	if finalSpace == -1 {
		return ""
	}
	lastNotSpace := lastIndexByteNot(text[:finalSpace], ' ')
	if lastNotSpace == -1 {
		return ""
	}
	prevSpace := strings.LastIndex(text[:lastNotSpace], " ")
	if prevSpace == -1 {
		return ""
	}
	return text[prevSpace+1 : lastNotSpace+1]
}

func lastIndexByteNot(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != c {
			return i
		}
	}
	return -1
}

// if there are no spaces this is the first word
func isFirstWord(text string) bool {
	return strings.LastIndex(text, " ") == -1
}

//
// keeping this around because we may need
// to revisit exit on non-darwin platforms.
// as per line #128
//
//
// https://github.com/c-bata/go-prompt/issues/59
// func exit(_ *prompt.Buffer) {
// 	fmt.Println("Ctrl+D :: exitCallback")
// 	panic(utils.ExitCode(0))
// }
