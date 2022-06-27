package interactive

import (
	"fmt"
	"sort"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/schema"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/utils"
)

// GetTableAutoCompleteSuggestions derives and returns tables for typeahead
func GetTableAutoCompleteSuggestions(schema *schema.Metadata, connectionMap *steampipeconfig.ConnectionDataMap) []prompt.Suggest {
	var s []prompt.Suggest

	// schema names
	var schemasToAdd []string
	// unqualified table names - initialise to the introspection table names
	unqualifiedTablesToAdd := []string{}
	// fully qualified table names
	var qualifiedTablesToAdd []string

	// keep track of which plugins we have added unqualified tables for
	pluginSchemaMap := map[string]bool{}

	for schemaName, schemaDetails := range schema.Schemas {
		isTemporarySchema := schemaName == schema.TemporarySchemaName

		// when the schema.Schemas map is built, it is built from the configured connections and `public`
		// all other schema are ignored.
		// therefore, the only schema which will not have a connection is `public`
		var pluginOfThisSchema string
		schemaConnection, hasConnectionForSchema := (*connectionMap)[schemaName]
		if hasConnectionForSchema {
			pluginOfThisSchema = stripVersionFromPluginName(schemaConnection.Plugin)
		}

		// add the schema into the list of schema
		if !isTemporarySchema {
			schemasToAdd = append(schemasToAdd, schemaName)
		}

		// add qualified names of all tables
		for tableName := range schemaDetails {
			if !isTemporarySchema {
				qualifiedTablesToAdd = append(qualifiedTablesToAdd, fmt.Sprintf("%s.%s", schemaName, sanitiseTableName(tableName)))
			}
		}

		// only add unqualified table name if the schema is in the search_path
		// and we have not added tables for another connection using the same plugin as this one
		schemaOfSamePluginIncluded := hasConnectionForSchema && pluginSchemaMap[pluginOfThisSchema]
		foundInSearchPath := helpers.StringSliceContains(schema.SearchPath, schemaName)

		if (foundInSearchPath || isTemporarySchema) && !schemaOfSamePluginIncluded {
			for tableName := range schemaDetails {
				unqualifiedTablesToAdd = append(unqualifiedTablesToAdd, tableName)
				if !isTemporarySchema {
					pluginSchemaMap[pluginOfThisSchema] = true
				}
			}
		}
	}

	sort.Strings(schemasToAdd)
	sort.Strings(unqualifiedTablesToAdd)
	sort.Strings(qualifiedTablesToAdd)

	for _, schema := range schemasToAdd {
		// we don't need to escape schema names, since schema names are derived from connection names
		// which are validated so that we don't end up with names which need it
		s = append(s, prompt.Suggest{Text: schema, Description: "Schema", Output: schema})
	}

	for _, table := range unqualifiedTablesToAdd {
		s = append(s, prompt.Suggest{Text: table, Description: "Table", Output: sanitiseTableName(table)})
	}

	for _, table := range qualifiedTablesToAdd {
		s = append(s, prompt.Suggest{Text: table, Description: "Table", Output: sanitiseTableName(table)})
	}

	return s
}

func stripVersionFromPluginName(pluginName string) string {
	return strings.Split(pluginName, "@")[0]
}

func sanitiseTableName(strToEscape string) string {
	tokens := utils.SplitByRune(strToEscape, '.')
	escaped := []string{}
	for _, token := range tokens {
		if strings.ContainsAny(token, " -") {
			token = db_common.PgEscapeName(token)
		}
		escaped = append(escaped, token)
	}
	return strings.Join(escaped, ".")
}
