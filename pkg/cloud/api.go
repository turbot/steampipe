package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"io"
	"net/http"
	"net/url"
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
	urlPath, err := url.JoinPath(baseURL, actorWorkspacesAPI)
	if err != nil {
		return nil, err
	}

	resp := map[string]any{}
	err = getFromAPI(urlPath, bearer, client, &resp)
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

func getActor(baseURL, bearer string, client *http.Client) (*Actor, error) {
	urlPath, err := url.JoinPath(baseURL, actorAPI)
	if err != nil {
		return nil, err
	}
	resp := &Actor{}
	err = getFromAPI(urlPath, bearer, client, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get actor from Steampipe Cloud API: %s ", err.Error())
	}

	return resp, nil
}

func getPassword(baseURL, userHandle, bearer string, client *http.Client) (string, error) {
	urlPath, err := url.JoinPath(baseURL, fmt.Sprintf(passwordAPIFormat, userHandle))
	if err != nil {
		return "", err
	}

	resp := map[string]any{}
	err = getFromAPI(urlPath, bearer, client, &resp)
	if err != nil {
		return "", fmt.Errorf("failed to get password from Steampipe Cloud API: %s ", err.Error())
	}

	password, ok := resp["$password"].(string)
	if !ok {
		return "", fmt.Errorf("failed to read password from Steampipe Cloud API")
	}
	return password, nil
}

func getFromAPI[T any](urlPath, bearer string, client *http.Client, dest *T) error {
	// build request
	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Add("Authorization", bearer)
	}

	return executeAPICall(req, client, dest)
}

func postToAPI[T any](urlPath, bearer, bodyStr string, client *http.Client, dest *T) error {
	// build request
	req, err := http.NewRequest("POST", urlPath, bytes.NewBuffer([]byte(bodyStr)))
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Add("Authorization", bearer)
	}

	return executeAPICall(req, client, dest)
}

func executeAPICall[T any](req *http.Request, client *http.Client, dest *T) error {
	// execute
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 206 {
		return fmt.Errorf("%s", resp.Status)
	}
	defer resp.Body.Close()

	// read response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// unmarshal response

	err = json.Unmarshal(bodyBytes, dest)
	if err != nil {
		return err
	}
	return nil
}
