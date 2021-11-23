package db_common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const baseURL = "https://latestpipe.turbot.io"
const actorAPI = "/api/v1/actor"
const actorWorkspacesAPI = "/api/v1/actor/workspace"
const passwordAPIFormat = "/api/v1/user/%s/password"
const userWorkspaceAPIFormat = "/api/v1/user/%s/workspace/%s"
const orgWorkspaceAPIFormat = "/api/v1/org/%s/workspace/%s"

func GetConnectionString(workspaceHandle, token string) (string, error) {
	//	parts := strings.Split(workspaceHandle, "/")
	//	if len(parts) != )
	//	user := parts[0]
	//	workspaceHandle := parts[1]
	//	// create a 'bearer' string by appending the access token
	//	var bearer = "Bearer " + token
	//
	//	client := &http.Client{}
	//
	//	// org or user?
	//	//GetActorWorkspaces
	//
	//	// "kai/dev"
	//	//userHandle, err := GetActor(bearer, client)
	//	//if err != nil {
	//	//	return "", err
	//	//}
	//	//{
	//	//identity: {
	//	//	type: "user" / "org",
	//	//	handle: "kai" / "acme" // identity handle
	//	//}
	//	//handle: "dev" // workspace handle
	//	//}
	//
	//	//workspaceHost, workspaceRandString, err := GetWorkspaceHost(userHandle, workspaceHandle, bearer, client)
	//
	//	if err != nil {
	//		return "", err
	//	}
	//	password, err := GetPassword(userHandle, bearer, client)
	//	if err != nil {
	//		return "", err
	//	}
	//
	////	if kai was trying to connect to acme org “dev” workspace, it would end up being:
	////postgresql://kai:kai-password@acme-dev.workspace-host:9193/workspace-db-name
	//
	//
	//	return fmt.Sprintf("postgresql://%s:%s@%s:9193/%s-%s", userHandle, password, workspaceHost, workspaceRandString, workspaceHandle), nil
	//							postgresql://${user.handle}:${res.data.$password}@${identity.handle}-${workspace.handle}.${workspace.host}:9193/${workspace.database_name}
	return "", nil
}

func GetActorWorkspaces(bearer string, client *http.Client) (string, error) {
	resp, err := fetchAPIData(baseURL+actorWorkspacesAPI, bearer, client)
	if err != nil {
		return "", err
	}

	// identity.type=org/user
	// HANDLE PAGING
	handle, ok := resp["handle"].(string)
	if !ok {
		return "", fmt.Errorf("failed to read handle from Steampipe Cloud API")
	}
	return handle, nil
}

func GetActor(bearer string, client *http.Client) (string, error) {
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

func GetPassword(userHandle, bearer string, client *http.Client) (string, error) {
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

// idtype, idhandle, wshandle
//func GetWorkspaceHost(userHandle, workspaceHandle, bearer string, client *http.Client) (string, string, error) {
//
//	// deduce correct api
//
//	url := baseURL + fmt.Sprintf(userWorkspaceAPIFormat, idhandle, workspaceHandle)
//	resp, err := fetchAPIData(url, bearer, client)
//	if err != nil {
//		return "", "", err
//	}
//
//	host, ok := resp["host"].(string)
//	if !ok {
//		return "", "", fmt.Errorf("failed to read workspace data from Steampipe Cloud API")
//	}
//	randString, ok := resp["rand_string"].(string)
//	if !ok {
//		return "", "", fmt.Errorf("failed to read workspace data from Steampipe Cloud API")
//	}
//	return host, randString, nil
//}

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
