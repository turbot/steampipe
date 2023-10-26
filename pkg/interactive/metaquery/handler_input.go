package metaquery

import (
	"github.com/c-bata/go-prompt"
	"github.com/turbot/pipe-fittings/db_common"
	"github.com/turbot/steampipe/pkg/db/steampipe_db_client"
	"github.com/turbot/steampipe/pkg/steampipe_config_local"
)

// HandlerInput defines input data for the metaquery handler
type HandlerInput struct {
	Client steampipe_db_client.SteampipeDbClient
	Schema *db_common.SchemaMetadata

	Prompt          *prompt.Prompt
	ClosePrompt     func()
	Query           string
	ConnectionState steampipe_config_local.ConnectionStateMap
	SearchPath      []string
}

func (h *HandlerInput) args() []string {
	return getArguments(h.Query)
}
