package cloud

import (
	"context"
	steampipecloud "github.com/turbot/steampipe-cloud-sdk-go"
)

// GetUserWorkspaceHandles returns all user workspace handles for the user with the given token
// (this is expected to be 0 or 1 workspace handle)
func GetUserWorkspaceHandles(ctx context.Context, token string) ([]steampipecloud.Workspace, string, error) {
	client := newSteampipeCloudClient(token)
	actor, _, err := client.Actors.Get(ctx).Execute()
	if err != nil {
		return nil, "", err
	}
	userHandler := actor.Handle
	workspaces, _, err := client.UserWorkspaces.List(ctx, userHandler).Execute()
	if err != nil {
		return nil, "", err
	}
	return workspaces.GetItems(), getActorName(actor), nil
}
