package interactive

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

func (c *InteractiveClient) initialiseSuggestions(ctx context.Context) error {
	log.Printf("[TRACE] initialiseSuggestions")

	conn, err := c.client().AcquireManagementConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	connectionStateMap, err := steampipeconfig.LoadConnectionState(ctx, conn.Conn(), steampipeconfig.WithWaitUntilLoading())
	if err != nil {
		log.Printf("[WARN] could not load connection state: %v", err)
		//nolint:golint,nilerr // valid condition - not an error
		return nil
	}

	// reset suggestions
	c.suggestions = newAutocompleteSuggestions()
	c.initialiseSchemaAndTableSuggestions(connectionStateMap)
	c.initialiseQuerySuggestions()
	c.suggestions.sort()
	return nil
}

// initialiseSchemaAndTableSuggestions build a list of schema and table querySuggestions
func (c *InteractiveClient) initialiseSchemaAndTableSuggestions(connectionStateMap steampipeconfig.ConnectionStateMap) {
	if c.schemaMetadata == nil {
		return
	}

	// check if client is nil to avoid panic
	if c.client() == nil {
		return
	}

	// unqualified table names
	// use lookup to avoid dupes from dynamic plugins
	// (this is needed as GetFirstSearchPathConnectionForPlugins will return ALL dynamic connections)
	var unqualifiedTablesToAdd = make(map[string]struct{})

	// add connection state and rate limit
	unqualifiedTablesToAdd[constants.ConnectionTable] = struct{}{}
	unqualifiedTablesToAdd[constants.PluginInstanceTable] = struct{}{}
	unqualifiedTablesToAdd[constants.RateLimiterDefinitionTable] = struct{}{}
	unqualifiedTablesToAdd[constants.PluginColumnTable] = struct{}{}
	unqualifiedTablesToAdd[constants.ServerSettingsTable] = struct{}{}

	// get the first search path connection for each plugin
	firstConnectionPerPlugin := connectionStateMap.GetFirstSearchPathConnectionForPlugins(c.client().GetRequiredSessionSearchPath())
	firstConnectionPerPluginLookup := utils.SliceToLookup(firstConnectionPerPlugin)
	// NOTE: add temporary schema into firstConnectionPerPluginLookup
	// as we want to add unqualified tables from there into autocomplete
	firstConnectionPerPluginLookup[c.schemaMetadata.TemporarySchemaName] = struct{}{}

	for schemaName, schemaDetails := range c.schemaMetadata.Schemas {
		if connectionState, found := connectionStateMap[schemaName]; found && connectionState.State != constants.ConnectionStateReady {
			log.Println("[TRACE] could not find schema in state map or connection is not Ready", schemaName)
			continue
		}

		// fully qualified table names
		var qualifiedTablesToAdd []prompt.Suggest

		isTemporarySchema := schemaName == c.schemaMetadata.TemporarySchemaName
		if !isTemporarySchema {
			// add the schema into the list of schema
			// we don't need to escape schema names, since schema names are derived from connection names
			// which are validated so that we don't end up with names which need it
			c.suggestions.schemas = append(c.suggestions.schemas, prompt.Suggest{Text: schemaName, Description: "Schema", Output: schemaName})
		}

		// add qualified names of all tables
		for tableName := range schemaDetails {
			// do not add temp tables to qualified tables
			if !isTemporarySchema {
				qualifiedTableName := fmt.Sprintf("%s.%s", schemaName, sanitiseTableName(tableName))
				qualifiedTablesToAdd = append(qualifiedTablesToAdd, prompt.Suggest{Text: qualifiedTableName, Description: "Table", Output: qualifiedTableName})
			}
			if _, addToUnqualified := firstConnectionPerPluginLookup[schemaName]; addToUnqualified {
				unqualifiedTablesToAdd[tableName] = struct{}{}
			}
		}

		// add qualified table to tablesBySchema with size limits
		if len(qualifiedTablesToAdd) > 0 {
			c.suggestions.setTablesForSchema(schemaName, qualifiedTablesToAdd)
		}
	}

	// add unqualified table suggestions
	for tableName := range unqualifiedTablesToAdd {
		c.suggestions.unqualifiedTables = append(c.suggestions.unqualifiedTables, prompt.Suggest{Text: tableName, Description: "Table", Output: sanitiseTableName(tableName)})
	}
}

func (c *InteractiveClient) initialiseQuerySuggestions() {
	//	 TODO add sql files???
}

func sanitiseTableName(strToEscape string) string {
	tokens := helpers.SplitByRune(strToEscape, '.')
	var escaped []string
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
