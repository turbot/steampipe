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
	actor, err := getActor(baseURL, bearer, client)
	if err != nil {
		return nil, "", err
	}
	workspaceHandles, err := getUserWorkspaceHandles(baseURL, bearer, actor.Handle, client)
	if err != nil {
		return nil, "", err
	}
	return workspaceHandles, actor.Handle, nil
}

func getUserWorkspaceHandles(baseURL, bearer, userHandle string, client *http.Client) ([]string, error) {
	urlPath, err := url.JoinPath(baseURL, fmt.Sprintf(userWorkspaceFormat, userHandle))
	if err != nil {
		return nil, err
	}
	// add in limit
	urlPath += "?limit=2"

	var resp map[string]any
	err = getFromAPI(urlPath, bearer, client, &resp)
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
