package metaquery

import (
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/turbot/steampipe/autocomplete"
	"github.com/turbot/steampipe/connection_config"
	"github.com/turbot/steampipe/schema"
)

// CompleterInput :: input interface for the metaquery completer
type CompleterInput struct {
	Query       string
	Schema      *schema.Metadata
	Connections *connection_config.ConnectionMap
}

func (h *CompleterInput) args() []string {
	return getArguments(h.Query)
}

type completer func(input *CompleterInput) []prompt.Suggest

// Complete :: return completions for metaqueries.
func Complete(input *CompleterInput) []prompt.Suggest {
	input.Query = strings.TrimSuffix(input.Query, ";")
	var s = strings.Fields(input.Query)

	metaQueryObj, found := metaQueryDefinitions[s[0]]
	if !found {
		return []prompt.Suggest{}
	}
	if metaQueryObj.completer == nil {
		return []prompt.Suggest{}
	}
	return metaQueryObj.completer(input)
}

func completerFromArgsOf(cmd string) completer {
	return func(input *CompleterInput) []prompt.Suggest {
		metaQueryDefinition, _ := metaQueryDefinitions[cmd]
		suggestions := make([]prompt.Suggest, len(metaQueryDefinition.args))
		for idx, arg := range metaQueryDefinition.args {
			suggestions[idx] = prompt.Suggest{Text: arg.value, Description: arg.description}
		}
		return suggestions
	}
}

func inspectCompleter(input *CompleterInput) []prompt.Suggest {
	return autocomplete.GetTableAutoCompleteSuggestions(input.Schema, input.Connections)
}
