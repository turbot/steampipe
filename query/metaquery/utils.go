package metaquery

import (
	"encoding/csv"
	"sort"
	"strings"

	"github.com/c-bata/go-prompt"
)

// IsMetaQuery :: returns true if the query is a metaquery, false otherwise
func IsMetaQuery(query string) bool {
	if !strings.HasPrefix(query, ".") {
		return false
	}

	// try to look for the validator
	cmd, _ := getCmdAndArgs(query)
	_, foundHandler := metaQueryDefinitions[cmd]

	return foundHandler
}

func getCmdAndArgs(query string) (string, []string) {
	query = strings.TrimSuffix(query, ";")
	split := splitByWhitespace(query)
	cmd := split[0]
	args := []string{}
	if len(split) > 1 {
		args = split[1:]
	}
	return cmd, args
}

// splitByWhitespace uses the CSV decoder, using '\s' as the separator rune
// this enables us to parse out the tokens - even if they are quoted and/or escaped
func splitByWhitespace(str string) (s []string) {
	csvDecoder := csv.NewReader(strings.NewReader(str))
	csvDecoder.Comma = ' '
	csvDecoder.LazyQuotes = true
	csvDecoder.TrimLeadingSpace = true
	// Read can never error, because we are passing in a StringReader
	// lookup csv.Reader.Read
	split, _ := csvDecoder.Read()
	return split
}

// PromptSuggestions :: Returns a list of the suggestions for go-prompt
func PromptSuggestions() []prompt.Suggest {
	suggestions := make([]prompt.Suggest, 0, len(metaQueryDefinitions))
	for k, definition := range metaQueryDefinitions {
		suggestions = append(suggestions, prompt.Suggest{Text: k, Description: definition.description})
	}

	sort.SliceStable(suggestions[:], func(i, j int) bool {
		return suggestions[i].Text < suggestions[j].Text
	})

	return suggestions
}

func getArguments(query string) []string {
	_, args := getCmdAndArgs(query)
	return args
}
