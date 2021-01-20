package metaquery

import (
	"fmt"
	"sort"
	"strings"

	"github.com/c-bata/go-prompt"
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
	return metaQueryObj.completer(input)
}
