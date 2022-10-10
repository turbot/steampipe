package cloud

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/turbot/steampipe/pkg/steampipeconfig"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
)

const actorAPI = "/api/v1/actor"
const actorWorkspacesAPI = "/api/v1/actor/workspace"
const passwordAPIFormat = "/api/v1/user/%s/password"
const userWorkspaceFormat = "/api/v1/user/%s/workspace"

func GetUserWorkspace(token string) (string, error) {
	baseURL := getBaseApiUrl()
	bearer := getBearerToken(token)
	client := &http.Client{}
	// get actor
	userHandle, _, err := getActor(baseURL, bearer, client)
	if err != nil {
		return "", err
	}
	userWorkspace, err := getUserWorkspaceHandle(baseURL, bearer, userHandle, client)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", userHandle, userWorkspace), nil
}

func getBearerToken(token string) string {
	// create a 'bearer' string by appending the access token
	var bearer = "Bearer " + token
	return bearer
}

func getBaseApiUrl() string {
	baseURL := fmt.Sprintf("https://%s", viper.GetString(constants.ArgCloudHost))
	return baseURL
}

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

func getWorkspaces(baseURL, bearer string, client *http.Client) ([]any, error) {
	resp, err := fetchAPIData(baseURL+actorWorkspacesAPI, bearer, client)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace data from Steampipe Cloud API: %s ", err.Error())
	}

	// TODO HANDLE PAGING
	items := resp["items"]
	if items != nil {
		return items.([]any), nil
	}
	return nil, nil

}

func getWorkspaceData(itemsArray []any, identityHandle, workspaceHandle string) map[string]any {
	for _, i := range itemsArray {
		item := i.(map[string]any)
		workspace := item["workspace"].(map[string]any)
		identity := item["identity"].(map[string]any)
		if identity["handle"] == identityHandle && workspace["handle"] == workspaceHandle {
			return item
		}
	}

	return nil
}

func getActor(baseURL, bearer string, client *http.Client) (string, string, error) {
	resp, err := fetchAPIData(baseURL+actorAPI, bearer, client)
	if err != nil {
		return "", "", fmt.Errorf("failed to get actor from Steampipe Cloud API: %s ", err.Error())
	}

	handle, ok := resp["handle"].(string)
	if !ok {
		return "", "", fmt.Errorf("failed to read handle from Steampipe Cloud API")
	}

	id, ok := resp["id"].(string)
	if !ok {
		return "", "", fmt.Errorf("failed to read id from Steampipe Cloud API")
	}

	return handle, id, nil
}

func getUserWorkspaceHandle(baseURL, bearer, userHandle string, client *http.Client) (string, error) {
	url := baseURL + fmt.Sprintf(userWorkspaceFormat, userHandle) + "?limit=2"
	resp, err := fetchAPIData(url, bearer, client)
	if err != nil {
		return "", fmt.Errorf("failed to get user workspace from Steampipe Cloud API: %s ", err.Error())
	}
	items := resp["items"].([]any)

	if len(items) == 0 {
		// CREATE??
		return "", fmt.Errorf("no workspace found for user %s", userHandle)
	}
	if len(items) > 1 {
		return "", fmt.Errorf("more than one workspace found for user - specify which one to use with '--workspace'")
	}
	workspace := items[0].(map[string]any)

	return workspace["handle"].(string), nil
}

func getPassword(baseURL, userHandle, bearer string, client *http.Client) (string, error) {
	url := baseURL + fmt.Sprintf(passwordAPIFormat, userHandle)
	resp, err := fetchAPIData(url, bearer, client)
	if err != nil {
		return "", fmt.Errorf("failed to get password from Steampipe Cloud API: %s ", err.Error())
	}

	password, ok := resp["$password"].(string)
	if !ok {
		return "", fmt.Errorf("failed to read password from Steampipe Cloud API")
	}
	return password, nil
}

func fetchAPIData(url, bearer string, client *http.Client) (map[string]any, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", bearer)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 206 {
		return nil, fmt.Errorf("%s", resp.Status)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
