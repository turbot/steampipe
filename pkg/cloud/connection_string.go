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

func GetCloudMetadata(workspaceDatabaseString, token string) (*steampipeconfig.CloudMetadata, error) {
	baseURL := fmt.Sprintf("https://%s", viper.GetString(constants.ArgCloudHost))
	parts := strings.Split(workspaceDatabaseString, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid 'workspace-database' argument '%s' - must be either a connection string or in format <identity>/<workspace>", workspaceDatabaseString)
	}
	identityHandle := parts[0]
	workspaceHandle := parts[1]

	// create a 'bearer' string by appending the access token
	var bearer = "Bearer " + token

	client := &http.Client{}

	// org or user?
	workspaceData, err := getWorkspaceData(baseURL, identityHandle, workspaceHandle, bearer, client)
	if err != nil {
		return nil, err
	}
	if workspaceData == nil {
		return nil, fmt.Errorf("failed to resolve workspace with identity handle '%s', workspace handle '%s'", identityHandle, workspaceHandle)
	}

	workspace := workspaceData["workspace"].(map[string]interface{})
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

	identity := workspaceData["identity"].(map[string]interface{})

	cloudMetadata := steampipeconfig.NewCloudMetadata()
	cloudMetadata.Actor.Id = userId
	cloudMetadata.Actor.Handle = userHandle
	cloudMetadata.Identity.Id = identity["id"].(string)
	cloudMetadata.Identity.Type = identity["type"].(string)
	cloudMetadata.Identity.Handle = identityHandle
	cloudMetadata.Workspace.Id = workspace["id"].(string)
	cloudMetadata.Workspace.Handle = workspace["handle"].(string)
	cloudMetadata.ConnectionString = connectionString

	return cloudMetadata, nil
}

func getWorkspaceData(baseURL, identityHandle, workspaceHandle, bearer string, client *http.Client) (map[string]interface{}, error) {
	resp, err := fetchAPIData(baseURL+actorWorkspacesAPI, bearer, client)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace data from Steampipe Cloud API: %s ", err.Error())
	}

	// TODO HANDLE PAGING
	items := resp["items"]
	if items != nil {
		itemsArray := items.([]interface{})
		for _, i := range itemsArray {
			item := i.(map[string]interface{})
			workspace := item["workspace"].(map[string]interface{})
			identity := item["identity"].(map[string]interface{})
			if identity["handle"] == identityHandle && workspace["handle"] == workspaceHandle {
				return item, nil
			}
		}
	}
	return nil, nil
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

func fetchAPIData(url, bearer string, client *http.Client) (map[string]interface{}, error) {
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
	var result map[string]interface{}
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
