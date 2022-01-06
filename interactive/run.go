package interactive

import (
	"context"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_local"
	"github.com/turbot/steampipe/query"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/utils"
)

// RunInteractivePrompt starts the interactive query prompt
func RunInteractivePrompt(ctx context.Context, initData *query.InitData) (*queryresult.ResultStreamer, error) {
	resultsStreamer := queryresult.NewResultStreamer()

	interactiveClient, err := newInteractiveClient(ctx, initData, resultsStreamer)
	if err != nil {
		utils.ShowErrorWithMessage(ctx, err, "interactive client failed to initialize")
		// do not bind shutdown to any cancellable context
		db_local.ShutdownService(ctx, constants.InvokerQuery)
		return nil, err
	}

	// start the interactive prompt in a go routine
	go interactiveClient.InteractivePrompt(ctx)

	return resultsStreamer, nil
}
