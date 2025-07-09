package metaquery

import (
	"context"

	"github.com/c-bata/go-prompt"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

type ConnectionStateGetter func(context.Context) (steampipeconfig.ConnectionStateMap, error)

// HandlerInput defines input data for the metaquery handler
type HandlerInput struct {
	Client db_common.Client
	Schema *db_common.SchemaMetadata

	Prompt                *prompt.Prompt
	ClosePrompt           func()
	Query                 string
	GetConnectionStateMap ConnectionStateGetter
	SearchPath            []string
}

func (h *HandlerInput) args() []string {
	return getArguments(h.Query)
}
