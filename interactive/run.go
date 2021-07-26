package interactive

import (
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/db/local_db"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/utils"
)

// RunInteractivePrompt :: start the interactive query prompt
func RunInteractivePrompt(initChan *chan *db_common.QueryInitData) (*queryresult.ResultStreamer, error) {
	resultsStreamer := queryresult.NewResultStreamer()

	interactiveClient, err := newInteractiveClient(initChan, resultsStreamer)
	if err != nil {
		utils.ShowErrorWithMessage(err, "interactive client failed to initialize")
		local_db.ShutdownService(constants.InvokerQuery)
		return nil, err
	}

	// start the interactive prompt in a go routine
	go interactiveClient.InteractiveQuery()

	return resultsStreamer, nil
}
