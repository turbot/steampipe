package metaquery

import (
	"github.com/c-bata/go-prompt"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
)

// HandlerInput defines input data for the metaquery handler
type HandlerInput struct {
	Client db_common.Client
	Schema *db_common.SchemaMetadata

	Prompt          *prompt.Prompt
	ClosePrompt     func()
	Query           string
	ConnectionState steampipeconfig.ConnectionStateMap
	SearchPath      []string
}

func (h *HandlerInput) args() []string {
	return getArguments(h.Query)
}
