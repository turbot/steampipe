package cloud

import (
	"context"
	"fmt"
	"strings"

	steampipecloud "github.com/turbot/steampipe-cloud-sdk-go"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/sperr"
)

func GetCloudMetadata(ctx context.Context, workspaceDatabaseString, token string) (*steampipeconfig.CloudMetadata, error) {
	client := newSteampipeCloudClient(token)

	parts := strings.Split(workspaceDatabaseString, "/")
	if len(parts) != 2 {
		return nil, sperr.New("invalid 'workspace-database' argument '%s' - must be either a connection string or in format <identity>/<workspace>", workspaceDatabaseString)
	}
	identityHandle := parts[0]
	workspaceHandle := parts[1]

	// get the identity
	identity, _, err := client.Identities.Get(ctx, identityHandle).Execute()
	if err != nil {
		return nil, sperr.New("Invalid 'workspace-database' argument '%s'.\nPlease check the identity and workspace names and try again.", workspaceDatabaseString)
	}

	// get the workspace
	var cloudWorkspace steampipecloud.Workspace
	if identity.Type == "user" {
		cloudWorkspace, _, err = client.UserWorkspaces.Get(ctx, identityHandle, workspaceHandle).Execute()
	} else {
		cloudWorkspace, _, err = client.OrgWorkspaces.Get(ctx, identityHandle, workspaceHandle).Execute()
	}

	if error_helpers.IsBadWorkspaceDatabaseArg(err) {
		return nil, sperr.New("Invalid 'workspace-database' argument '%s'.\nPlease check the workspace name and try again.", workspaceDatabaseString)
	} else if error_helpers.IsInvalidCloudToken(err) {
		return nil, error_helpers.InvalidCloudTokenError
	}

	workspaceHost := cloudWorkspace.GetHost()
	databaseName := cloudWorkspace.GetDatabaseName()

	actor, _, err := client.Actors.Get(ctx).Execute()
	if err != nil {
		return nil, error_helpers.InvalidCloudTokenError
	}

	password, _, err := client.Users.GetDBPassword(ctx, actor.GetHandle()).Execute()
	if err != nil {
		return nil, sperr.Wrap(err)
	}

	connectionString := fmt.Sprintf("postgresql://%s:%s@%s-%s.%s:9193/%s", actor.Handle, password.Password, identityHandle, workspaceHandle, workspaceHost, databaseName)

	cloudMetadata := &steampipeconfig.CloudMetadata{
		Actor: &steampipeconfig.ActorMetadata{
			Id:     actor.Id,
			Handle: actor.Handle,
		},
		Identity: &steampipeconfig.IdentityMetadata{
			Id:     cloudWorkspace.IdentityId,
			Type:   identity.Type,
			Handle: identityHandle,
		},
		WorkspaceDatabase: &steampipeconfig.WorkspaceMetadata{
			Id:     cloudWorkspace.Id,
			Handle: cloudWorkspace.Handle,
		},

		ConnectionString: connectionString,
	}

	return cloudMetadata, nil
}
