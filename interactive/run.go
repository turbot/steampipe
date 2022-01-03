package interactive

import (
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_local"
	"github.com/turbot/steampipe/query"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/statusspinner"
	"github.com/turbot/steampipe/utils"
)

// RunInteractivePrompt starts the interactive query prompt
func RunInteractivePrompt(initData *query.InitData) (*queryresult.ResultStreamer, error) {
	resultsStreamer := queryresult.NewResultStreamer()

	interactiveClient, err := newInteractiveClient(initData, resultsStreamer)
	if err != nil {
		utils.ShowErrorWithMessage(err, "interactive client failed to initialize")
		// do not bind shutdown to any cancellable context
		// TODO CLEAR DELAY on status hook???
		db_local.ShutdownService(constants.InvokerQuery, statusHook)
		return nil, err
	}

	// start the interactive prompt in a go routine
	go interactiveClient.InteractivePrompt()

	return resultsStreamer, nil
}
