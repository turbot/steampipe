package db_common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const baseURL = "https://latestpipe.turbot.io"
const actorAPI = "/api/v1/actor"
const passwordAPIFormat = "/api/v1/user/%s/password"
const workspaceAPIFormat = "/api/v1/user/%s/workspace/%s"

func GetConnectionString(workspaceHandle, token string) (string, error) {
	// create a 'bearer' string by appending the access token
	var bearer = "Bearer " + token

	client := &http.Client{}

	userHandle, err := GetActor(bearer, client)
	if err != nil {
		return "", err
	}
	workspaceHost, workspaceRandString, err := GetWorkspaceHost(userHandle, workspaceHandle, bearer, client)
	if err != nil {
		return "", err
	}
	password, err := GetPassword(userHandle, bearer, client)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("postgresql://%s:%s@%s:9193/%s-%s", userHandle, password, workspaceHost, workspaceRandString, workspaceHandle), nil

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

func GetWorkspaceHost(userHandle, workspaceHandle, bearer string, client *http.Client) (string, string, error) {
	url := baseURL + fmt.Sprintf(workspaceAPIFormat, userHandle, workspaceHandle)
	resp, err := fetchAPIData(url, bearer, client)
	if err != nil {
		return "", "", err
	}

	host, ok := resp["host"].(string)
	if !ok {
		return "", "", fmt.Errorf("failed to read workspace data from Steampipe Cloud API")
	}
	randString, ok := resp["rand_string"].(string)
	if !ok {
		return "", "", fmt.Errorf("failed to read workspace data from Steampipe Cloud API")
	}
	return host, randString, nil
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
	bodyBytes, err := ioutil.ReadAll(resp.Body)
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
