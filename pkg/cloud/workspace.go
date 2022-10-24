package cloud

import (
	"fmt"
	"net/http"
	"net/url"
)

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

func getUserWorkspaceHandle(baseURL, bearer, userHandle string, client *http.Client) (string, error) {
	urlPath, err := url.JoinPath(baseURL, fmt.Sprintf(userWorkspaceFormat, userHandle)+"?limit=2")
	if err != nil {
		return "", err
	}

	resp, err := fetchAPIData(urlPath, bearer, client)
	if err != nil {
		return "", fmt.Errorf("failed to get user workspace from Steampipe Cloud API: %s ", err.Error())
	}
	items := resp["items"].([]any)

	if len(items) == 0 {
		return "", fmt.Errorf("snapshot-location is not specified and no workspaces exist for user %s", userHandle)
	}
	if len(items) > 1 {
		return "", fmt.Errorf("snapshot-location is not specified and more than one workspace found for user %s", userHandle)
	}
	workspace := items[0].(map[string]any)

	return workspace["handle"].(string), nil
}
