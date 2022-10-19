package cloud

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"io"
	"net/http"
)

func getBearerToken(token string) string {
	// create a 'bearer' string by appending the access token
	var bearer = "Bearer " + token
	return bearer
}

func getBaseApiUrl() string {
	baseURL := fmt.Sprintf("https://%s", viper.GetString(constants.ArgCloudHost))
	return baseURL
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
