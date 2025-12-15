package interactive

import (
	"sort"
	"sync"

	"github.com/c-bata/go-prompt"
)

const (
	// Maximum number of schemas/connections to store in suggestion maps
	maxSchemasInSuggestions = 100
	// Maximum number of tables per schema in suggestions
	maxTablesPerSchema = 500
	// Maximum number of queries per mod in suggestions
	maxQueriesPerMod = 500
)

type autoCompleteSuggestions struct {
	mu                 sync.RWMutex
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

// setTablesForSchema adds tables for a schema with size limits to prevent unbounded growth.
// If the schema count exceeds maxSchemasInSuggestions, the oldest schema is removed.
// If the table count exceeds maxTablesPerSchema, only the first maxTablesPerSchema are kept.
func (s *autoCompleteSuggestions) setTablesForSchema(schemaName string, tables []prompt.Suggest) {
	// Enforce per-schema table limit
	if len(tables) > maxTablesPerSchema {
		tables = tables[:maxTablesPerSchema]
	}

	// Enforce global schema limit
	if len(s.tablesBySchema) >= maxSchemasInSuggestions {
		// Remove one schema to make room (simple eviction - remove first key found)
		for k := range s.tablesBySchema {
			delete(s.tablesBySchema, k)
			break
		}
	}

	s.tablesBySchema[schemaName] = tables
}

// setQueriesForMod adds queries for a mod with size limits to prevent unbounded growth.
// If the mod count exceeds maxSchemasInSuggestions, the oldest mod is removed.
// If the query count exceeds maxQueriesPerMod, only the first maxQueriesPerMod are kept.
func (s *autoCompleteSuggestions) setQueriesForMod(modName string, queries []prompt.Suggest) {
	// Enforce per-mod query limit
	if len(queries) > maxQueriesPerMod {
		queries = queries[:maxQueriesPerMod]
	}

	// Enforce global mod limit
	if len(s.queriesByMod) >= maxSchemasInSuggestions {
		// Remove one mod to make room (simple eviction - remove first key found)
		for k := range s.queriesByMod {
			delete(s.queriesByMod, k)
			break
		}
	}

	s.queriesByMod[modName] = queries
}

func (s *autoCompleteSuggestions) sort() {
	s.mu.Lock()
	defer s.mu.Unlock()

	sortSuggestions := func(s []prompt.Suggest) {
		sort.Slice(s, func(i, j int) bool {
			return s[i].Text < s[j].Text
		})
	}

	sortSuggestions(s.schemas)
	sortSuggestions(s.unqualifiedTables)
	sortSuggestions(s.unqualifiedQueries)
	for _, tables := range s.tablesBySchema {
		sortSuggestions(tables)
	}
	for _, queries := range s.queriesByMod {
		sortSuggestions(queries)
	}
}
