package metaquery

import (
	"context"
	"sort"
	"strings"

	"github.com/c-bata/go-prompt"
)

// IsMetaQuery :: returns true if the query is a metaquery, false otherwise
func IsMetaQuery(query string) bool {
	if !strings.HasPrefix(query, ".") {
		return false
	}

	query = strings.TrimSuffix(query, ";")

	// try to look for the validator
	cmd := strings.Fields(query)[0]
	_, foundHandler := metaQueryDefinitions[cmd]

	return foundHandler
}

// PromptSuggestions :: Returns a list of the suggestions for go-prompt
func PromptSuggestions(context.Context) []prompt.Suggest {
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
	return strings.Fields(strings.TrimSpace(query))[1:]
}
