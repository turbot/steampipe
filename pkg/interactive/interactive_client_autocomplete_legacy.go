package interactive

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"golang.org/x/exp/maps"
	"sort"
	"strings"
)

func (c *InteractiveClient) initialiseSuggestionsLegacy() {
	c.initialiseQuerySuggestionsLegacy()
	c.initialiseTableSuggestionsLegacy()
}

func (c *InteractiveClient) initialiseQuerySuggestionsLegacy() {
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

		res = append(res, c.newSuggestion(itemType, qp.GetDescription(), name))
		return
	}

	c.workspace().GetResourceMaps().WalkResources(resourceFunc)

	// sort the suggestions
	sort.Slice(res, func(i, j int) bool {
		return res[i].Text < res[j].Text
	})
	c.suggestions.unqualifiedQueries = res
}

// initialiseTableSuggestions build a list of schema and table querySuggestions
func (c *InteractiveClient) initialiseTableSuggestionsLegacy() {

	if c.schemaMetadata == nil {
		return
	}

	// schema names
	var schemasToAdd []string
	// unqualified table names - initialise to the introspection table names
	var unqualifiedTablesToAddMap = make(map[string]struct{})
	var unqualifiedTablesToAdd []string

	// keep track of which plugins we have added unqualified tables for
	//pluginSchemaMap := map[string]bool{}

	for schemaName, schemaDetails := range c.schemaMetadata.Schemas {
		// fully qualified table names
		var qualifiedTablesToAdd []string
		isTemporarySchema := schemaName == c.schemaMetadata.TemporarySchemaName

		// add the schema into the list of schema
		if !isTemporarySchema {
			schemasToAdd = append(schemasToAdd, schemaName)
		}

		// add qualified names of all tables
		for tableName := range schemaDetails {
			if !isTemporarySchema {

				qualifiedTablesToAdd = append(qualifiedTablesToAdd, fmt.Sprintf("%s.%s", schemaName, sanitiseTableName(tableName)))

				if helpers.StringSliceContains(c.client().GetRequiredSessionSearchPath(), schemaName) {
					unqualifiedTablesToAddMap[tableName] = struct{}{}
				}
			}
		}

		sort.Strings(qualifiedTablesToAdd)
		var tableSuggestions []prompt.Suggest
		for _, t := range qualifiedTablesToAdd {
			tableSuggestions = append(tableSuggestions, prompt.Suggest{Text: t, Description: "Table", Output: sanitiseTableName(t)})
		}
		c.suggestions.tablesBySchema[schemaName] = tableSuggestions
	}

	sort.Strings(schemasToAdd)
	for _, schema := range schemasToAdd {
		// we don't need to escape schema names, since schema names are derived from connection names
		// which are validated so that we don't end up with names which need it
		c.suggestions.schemas = append(c.suggestions.schemas, prompt.Suggest{Text: schema, Description: "Schema", Output: schema})
	}

	unqualifiedTablesToAdd = maps.Keys(unqualifiedTablesToAddMap)
	sort.Strings(unqualifiedTablesToAdd)
	for _, table := range unqualifiedTablesToAdd {
		c.suggestions.unqualifiedTables = append(c.suggestions.unqualifiedTables, prompt.Suggest{Text: table, Description: "Table", Output: sanitiseTableName(table)})
	}
}

func stripVersionFromPluginName(pluginName string) string {
	return strings.Split(pluginName, "@")[0]
}
