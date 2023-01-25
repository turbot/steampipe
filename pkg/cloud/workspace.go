package cloud

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/sperr"
)

// GetUserWorkspaceHandle returns the handle of the user workspace
//
//	in format actorHandle/workspaceHandle
//
// if there are 0 or > 1 workspaces this is an error
func GetUserWorkspaceHandle(ctx context.Context, token string) (string, error) {
	client := newSteampipeCloudClient(token)
	actor, _, err := client.Actors.Get(ctx).Execute()
	if err.Error() == "400 Bad Request" {
		return "", fmt.Errorf("Invalid token.\nPlease run %s or setup a token.", constants.Bold(fmt.Sprintf("steampipe login")))
	}
	if err != nil {
		return "", sperr.Wrap(err)
	}
	userHandler := actor.Handle
	workspacesResponse, _, err := client.UserWorkspaces.List(ctx, userHandler).Execute()
	if err != nil {
		return "", sperr.Wrap(err)
	}
	workspaces := workspacesResponse.GetItems()

	if len(workspaces) == 0 {
		return "", sperr.New("snapshot-location is not specified and no workspaces exist for user %s", getActorName(actor))
	}
	if len(workspaces) > 1 {
		return "", sperr.New("more than one workspace found for user %s", getActorName(actor))
	}

	workspaceHandle := fmt.Sprintf("%s/%s", actor.GetHandle(), workspaces[0].GetHandle())
	return workspaceHandle, nil
}
