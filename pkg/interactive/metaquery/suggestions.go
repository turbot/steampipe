package metaquery

import (
	"sort"

	"github.com/c-bata/go-prompt"
)

// PromptSuggestions returns a list of the metaquery suggestions for go-prompt
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
