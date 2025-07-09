package interactive

import (
	"context"

	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_local"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
	"github.com/turbot/steampipe/v2/pkg/query"
	"github.com/turbot/steampipe/v2/pkg/query/queryresult"
)

type RunInteractivePromptResult struct {
	Streamer  *queryresult.ResultStreamer
	PromptErr error
}

// RunInteractivePrompt starts the interactive query prompt
func RunInteractivePrompt(ctx context.Context, initData *query.InitData) *RunInteractivePromptResult {
	res := &RunInteractivePromptResult{
		Streamer: queryresult.NewResultStreamer(),
	}

	interactiveClient, err := newInteractiveClient(ctx, initData, res)
	if err != nil {
		error_helpers.ShowErrorWithMessage(ctx, err, "interactive client failed to initialize")
		// do not bind shutdown to any cancellable context
		db_local.ShutdownService(ctx, constants.InvokerQuery)
		res.PromptErr = err
		return res
	}

	// start the interactive prompt in a go routine
	go interactiveClient.InteractivePrompt(ctx)

	return res
}
