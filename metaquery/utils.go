package metaquery

import (
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
	_, foundHandler := metaQueryHandlers[cmd]

	return foundHandler
}

// PromptSuggestions :: Returns a list of the suggestions for go-prompt
func PromptSuggestions() []prompt.Suggest {
	suggestions := make([]prompt.Suggest, 0, len(metaQueryHandlers))
	for k := range metaQueryHandlers {
		suggestions = append(suggestions, prompt.Suggest{Text: k, Description: metaQueryHandlers[k].description})
	}

	sort.SliceStable(suggestions[:], func(i, j int) bool {
		return suggestions[i].Text < suggestions[j].Text
	})

	return suggestions
}

// PromptArgsSuggestions :: Returns a list of the suggestions for go-prompt
func PromptArgsSuggestions(query string) []prompt.Suggest {
	var suggestions []prompt.Suggest
	// if this is a metaquery, get the args

	if metaquery, ok := metaQueryHandlers[query]; ok {
		args := metaquery.args
		suggestions = make([]prompt.Suggest, len(args))
		for i, arg := range args {
			suggestions[i] = prompt.Suggest{Text: arg}
		}

		sort.SliceStable(suggestions[:], func(i, j int) bool {
			return suggestions[i].Text < suggestions[j].Text
		})
	}

	return suggestions
}
func getArguments(query string) []string {
	return strings.Fields(strings.TrimSpace(query))[1:]
}
