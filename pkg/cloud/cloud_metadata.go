package cloud

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/turbot/steampipe/pkg/steampipeconfig"
)

const actorAPI = "/api/v1/actor"
const loginTokenAPI = "/api/latest/login/token"
const actorWorkspacesAPI = "/api/v1/actor/workspace"
const passwordAPIFormat = "/api/v1/user/%s/password"
const userWorkspaceFormat = "/api/v1/user/%s/workspace"

func GetCloudMetadata(workspaceDatabaseString, token string) (*steampipeconfig.CloudMetadata, error) {
	bearer := getBearerToken(token)
	client := &http.Client{}
	baseURL := getBaseApiUrl()

	parts := strings.Split(workspaceDatabaseString, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid 'workspace-database' argument '%s' - must be either a connection string or in format <identity>/<workspace>", workspaceDatabaseString)
	}
	identityHandle := parts[0]
	workspaceHandle := parts[1]

	// org or user?
	workspaces, err := getWorkspaces(baseURL, bearer, client)
	if err != nil {
		return nil, err
	}
	workspaceData := getWorkspaceData(workspaces, identityHandle, workspaceHandle)
	if workspaceData == nil {
		return nil, fmt.Errorf("failed to resolve workspace with identity handle '%s', workspace handle '%s'", identityHandle, workspaceHandle)
	}

	workspace := workspaceData["workspace"].(map[string]any)
	workspaceHost := workspace["host"].(string)
	databaseName := workspace["database_name"].(string)

	userHandle, userId, err := getActor(baseURL, bearer, client)
	if err != nil {
		return nil, err
	}
	password, err := getPassword(baseURL, userHandle, bearer, client)
	if err != nil {
		return nil, err
	}

	connectionString := fmt.Sprintf("postgresql://%s:%s@%s-%s.%s:9193/%s", userHandle, password, identityHandle, workspaceHandle, workspaceHost, databaseName)

	identity := workspaceData["identity"].(map[string]any)

	cloudMetadata := &steampipeconfig.CloudMetadata{
		Actor: &steampipeconfig.ActorMetadata{
			Id:     userId,
			Handle: userHandle,
		},
		Identity: &steampipeconfig.IdentityMetadata{
			Id:     identity["id"].(string),
			Type:   identity["type"].(string),
			Handle: identityHandle,
		},
		WorkspaceDatabase: &steampipeconfig.WorkspaceMetadata{
			Id:     workspace["id"].(string),
			Handle: workspace["handle"].(string),
		},

		ConnectionString: connectionString,
	}

	return cloudMetadata, nil
}
