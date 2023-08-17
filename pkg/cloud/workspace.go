package cloud

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/error_helpers"
)

// GetUserWorkspaceHandle returns the handle of the user workspace
//
//	in format actorHandle/workspaceHandle
//
// if there are 0 or > 1 workspaces this is an error
func GetUserWorkspaceHandle(ctx context.Context, token string) (string, error) {
	client := newSteampipeCloudClient(token)
	actor, _, err := client.Actors.Get(ctx).Execute()
	if err != nil {
		return "", error_helpers.InvalidCloudTokenError
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
