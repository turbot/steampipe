package autocomplete

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/steampipeconfig"
)

// GetTableAutoCompleteSuggestions :: derives and returns tables for typeahead
func GetTableAutoCompleteSuggestions(schema *schema.Metadata, connectionMap *steampipeconfig.ConnectionMap) []prompt.Suggest {
	s := []prompt.Suggest{}

	// schema names
	schemasToAdd := []string{}
	// unqualified table names
	unqualifiedTablesToAdd := []string{}
	// fully qualified table names
	qualifiedTablesToAdd := []string{}

	unqualifiedTableMap := map[string]bool{}

	for schemaName, schemaDetails := range schema.Schemas {
		schemasToAdd = append(schemasToAdd, schemaName)

		// decide whether we need to include this schema in unqualified table list as well
		schemaConnection, found := (*connectionMap)[schemaName]
		if found {
			pluginOfThisSchema := stripVersionFromPluginName(schemaConnection.Plugin)
			isIncluded := unqualifiedTableMap[pluginOfThisSchema]
			for tableName := range schemaDetails {
				qualifiedTablesToAdd = append(qualifiedTablesToAdd, fmt.Sprintf("%s.%s", schemaName, tableName))
				if !isIncluded {
					unqualifiedTablesToAdd = append(unqualifiedTablesToAdd, tableName)
					unqualifiedTableMap[pluginOfThisSchema] = true
				}
			}
		} else if schemaName == "public" {
			for tableName := range schemaDetails {
				qualifiedTablesToAdd = append(qualifiedTablesToAdd, fmt.Sprintf("%s.%s", schemaName, tableName))
				unqualifiedTablesToAdd = append(unqualifiedTablesToAdd, tableName)
			}
		}
	}

	sort.Strings(schemasToAdd)
	sort.Strings(unqualifiedTablesToAdd)
	sort.Strings(qualifiedTablesToAdd)

	for _, schema := range schemasToAdd {
		s = append(s, prompt.Suggest{Text: schema, Description: "Schema"})
	}

	for _, table := range unqualifiedTablesToAdd {
		log.Println(fmt.Sprintf("%s %s", "[TRACE]", table))
		s = append(s, prompt.Suggest{Text: table, Description: "Table"})
	}

	for _, table := range qualifiedTablesToAdd {
		s = append(s, prompt.Suggest{Text: table, Description: "Table"})
	}

	return s
}

func stripVersionFromPluginName(pluginName string) string {
	return strings.Split(pluginName, "@")[0]
}
