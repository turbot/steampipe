package cloud

import (
	"fmt"
	"net/http"
	"net/url"
)

// GetUserWorkspaceHandles returns all user workspace handles for the user with the given token
// (this is expected to be 0 or 1 workspace handle)
func GetUserWorkspaceHandles(token string) ([]string, string, error) {
	baseURL := getBaseApiUrl()
	bearer := getBearerToken(token)
	client := &http.Client{}
	// get actor
	userHandle, _, err := getActor(baseURL, bearer, client)
	if err != nil {
		return nil, "", err
	}
	workspaceHandles, err := getUserWorkspaceHandles(baseURL, bearer, userHandle, client)
	if err != nil {
		return nil, "", err
	}
	return workspaceHandles, userHandle, nil
}

func getUserWorkspaceHandles(baseURL, bearer, userHandle string, client *http.Client) ([]string, error) {
	workspaceApiPath, err := url.JoinPath(baseURL, fmt.Sprintf(userWorkspaceFormat, userHandle))
	if err != nil {
		return nil, err
	}

	workspaceApiPathWithLimit := fmt.Sprintf("%s?limit=2", workspaceApiPath)

	resp, err := getFromAPI(workspaceApiPathWithLimit, bearer, client)
	if err != nil {
		return nil, fmt.Errorf("failed to get user workspace from Steampipe Cloud API: %s ", err.Error())
	}
	items := resp["items"].([]any)

	res := make([]string, len(items))
	for i, item := range items {
		workspace := item.(map[string]any)

		res[i] = fmt.Sprintf("%s/%s", userHandle, workspace["handle"].(string))
	}
	return res, nil
}
