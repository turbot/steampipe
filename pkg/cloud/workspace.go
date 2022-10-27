package cloud

import (
	"context"
	"fmt"
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
		return "", err
	}
	userHandler := actor.Handle
	workspacesResponse, _, err := client.UserWorkspaces.List(ctx, userHandler).Execute()
	if err != nil {
		return "", err
	}
	workspaces := workspacesResponse.GetItems()

	if len(workspaces) == 0 {
		return "", fmt.Errorf("snapshot-location is not specified and no workspaces exist for user %s", getActorName(actor))
	}
	if len(workspaces) > 1 {
		return "", fmt.Errorf("more than one workspace found for user %s", getActorName(actor))
	}

	workspaceHandle := fmt.Sprintf("%s/%s", actor.GetHandle(), workspaces[0].GetHandle())
	return workspaceHandle, nil
}
