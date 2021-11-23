package db_common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
)

const actorAPI = "/api/v1/actor"
const actorWorkspacesAPI = "/api/v1/actor/workspace"
const passwordAPIFormat = "/api/v1/user/%s/password"

func GetConnectionString(workspaceDatabaseString, token string) (string, error) {
	baseURL := fmt.Sprintf("https://%s", viper.GetString(constants.ArgCloudHost))
	parts := strings.Split(workspaceDatabaseString, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("workspace-database must be either a connection string or '<user handle>/<workspace handle>")
	}
	identityHandle := parts[0]
	workspaceHandle := parts[1]

	// create a 'bearer' string by appending the access token
	var bearer = "Bearer " + token

	client := &http.Client{}

	// org or user?
	workspace, err := GetWorkspaceData(baseURL, identityHandle, workspaceHandle, bearer, client)
	if err != nil {
		return "", err
	}
	workspaceHost := workspace["host"].(string)
	databaseName := workspace["database_name"].(string)

	userHandle, err := getActor(baseURL, bearer, client)
	if err != nil {
		return "", err
	}
	password, err := getPassword(baseURL, userHandle, bearer, client)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("postgresql://%s:%s@%s-%s.%s:9193/%s", userHandle, password, identityHandle, workspaceHandle, workspaceHost, databaseName), nil
}

func GetWorkspaceData(baseURL, identityHandle, workspaceHandle, bearer string, client *http.Client) (map[string]interface{}, error) {
	resp, err := fetchAPIData(baseURL+actorWorkspacesAPI, bearer, client)
	if err != nil {
		return nil, err
	}

	// TODO HANDLE PAGING
	items := resp["items"].([]interface{})
	for _, i := range items {
		item := i.(map[string]interface{})
		identity := item["identity"].(map[string]interface{})
		if identity["handle"] == identityHandle && item["handle"] == workspaceHandle {
			return item, nil
		}
	}
	return nil, nil
}

func getActor(baseURL, bearer string, client *http.Client) (string, error) {
	resp, err := fetchAPIData(baseURL+actorAPI, bearer, client)
	if err != nil {
		return "", err
	}

	handle, ok := resp["handle"].(string)
	if !ok {
		return "", fmt.Errorf("failed to read handle from Steampipe Cloud API")
	}
	return handle, nil
}

func getPassword(baseURL, userHandle, bearer string, client *http.Client) (string, error) {
	url := baseURL + fmt.Sprintf(passwordAPIFormat, userHandle)
	resp, err := fetchAPIData(url, bearer, client)
	if err != nil {
		return "", err
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
