package interactive

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"sort"
	"strings"
)

func (c *InteractiveClient) initialiseSuggestions() {
	c.initialiseTableSuggestions()
	c.initialiseTableSuggestions()
}

func (c *InteractiveClient) initialiseQuerySuggestions() {
	var res []prompt.Suggest

	workspaceModName := c.initData.Workspace.Mod.Name()
	resourceFunc := func(item modconfig.HclResource) (continueWalking bool, err error) {
		continueWalking = true

		qp, ok := item.(modconfig.QueryProvider)
		if !ok {
			return
		}
		modTreeItem, ok := item.(modconfig.ModTreeItem)
		if !ok {
			return
		}
		if qp.GetQuery() == nil && qp.GetSQL() == nil {
			return
		}
		rm := item.(modconfig.ResourceWithMetadata)
		if rm.IsAnonymous() {
			return
		}
		isLocal := modTreeItem.GetMod().Name() == workspaceModName
		itemType := item.BlockType()
		// only include global inputs
		if itemType == modconfig.BlockTypeInput {
			if _, ok := c.initData.Workspace.Mod.ResourceMaps.GlobalDashboardInputs[item.Name()]; !ok {
				return
			}
		}
		// special case for query
		if itemType == modconfig.BlockTypeQuery {
			itemType = "named query"
		}
		name := qp.Name()
		if isLocal {
			name = qp.GetUnqualifiedName()
		}

		res = append(res, c.addSuggestion(itemType, qp.GetDescription(), name))
		return
	}

	c.workspace().GetResourceMaps().WalkResources(resourceFunc)

	// sort the suggestions
	sort.Slice(res, func(i, j int) bool {
		return res[i].Text < res[j].Text
	})
	c.querySuggestions = res
}

// initialiseTableSuggestions build a list of schema and table querySuggestions
func (c *InteractiveClient) initialiseTableSuggestions() {

	if c.schemaMetadata == nil {
		return
	}

	var s []prompt.Suggest
	connectionMap := c.initData.Client.ConnectionMap()

	// schema names
	var schemasToAdd []string
	// unqualified table names - initialise to the introspection table names
	unqualifiedTablesToAdd := []string{}
	// fully qualified table names
	var qualifiedTablesToAdd []string

	// keep track of which plugins we have added unqualified tables for
	pluginSchemaMap := map[string]bool{}

	for schemaName, schemaDetails := range c.schemaMetadata.Schemas {
		isTemporarySchema := schemaName == c.schemaMetadata.TemporarySchemaName

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
		foundInSearchPath := helpers.StringSliceContains(c.schemaMetadata.SearchPath, schemaName)

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
		s = append(s, prompt.Suggest{Text: table, Description: "Table", Output: table})
	}

	c.tableSuggestions = s
}

func stripVersionFromPluginName(pluginName string) string {
	return strings.Split(pluginName, "@")[0]
}

func sanitiseTableName(strToEscape string) string {
	tokens := helpers.SplitByRune(strToEscape, '.')
	escaped := []string{}
	for _, token := range tokens {
		// if string contains spaces or special characters(-) or upper case characters, escape it,
		// as Postgres by default converts to lower case
		if strings.ContainsAny(token, " -") || utils.ContainsUpper(token) {
			token = db_common.PgEscapeName(token)
		}
		escaped = append(escaped, token)
	}
	return strings.Join(escaped, ".")
}
