package metaquery

import (
	"sort"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/turbot/steampipe/pkg/utils"
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
	split := utils.SplitByWhitespace(query)
	cmd := split[0]
	args := []string{}
	if len(split) > 1 {
		args = split[1:]
	}
	return cmd, args
}

// PromptSuggestions :: Returns a list of the suggestions for go-prompt
func PromptSuggestions() []prompt.Suggest {
	suggestions := make([]prompt.Suggest, 0, len(metaQueryDefinitions))
	for k, definition := range metaQueryDefinitions {
		suggestions = append(suggestions, prompt.Suggest{Text: k, Description: definition.description, Output: k})
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
