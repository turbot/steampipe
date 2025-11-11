package interactive

import (
	"github.com/c-bata/go-prompt"
	"sort"
)

type autoCompleteSuggestions struct {
	schemas            []prompt.Suggest
	unqualifiedTables  []prompt.Suggest
	unqualifiedQueries []prompt.Suggest
	tablesBySchema     map[string][]prompt.Suggest
	queriesByMod       map[string][]prompt.Suggest
	mods               []prompt.Suggest
}

func newAutocompleteSuggestions() *autoCompleteSuggestions {
	return &autoCompleteSuggestions{
		tablesBySchema: make(map[string][]prompt.Suggest),
		queriesByMod:   make(map[string][]prompt.Suggest),
	}
}
func (s autoCompleteSuggestions) sort() {
	sortSuggestions := func(s []prompt.Suggest) {
		sort.Slice(s, func(i, j int) bool {
			return s[i].Text < s[j].Text
		})
	}

	sortSuggestions(s.schemas)
	sortSuggestions(s.unqualifiedTables)
	sortSuggestions(s.unqualifiedQueries)

	// Defensive nil checks for maps
	// While Go handles nil map iteration safely, explicit checks improve code clarity
	if s.tablesBySchema != nil {
		for _, tables := range s.tablesBySchema {
			sortSuggestions(tables)
		}
	}
	if s.queriesByMod != nil {
		for _, queries := range s.queriesByMod {
			sortSuggestions(queries)
		}
	}
}
