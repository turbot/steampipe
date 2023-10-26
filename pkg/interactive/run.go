package interactive

import (
	"context"
	"github.com/turbot/pipe-fittings/constants"

	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/queryresult"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/query"
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
