package cloud

import (
	"fmt"
	"net/http"
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
