package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/turbot/steampipe/pkg/utils"
	"io"
	"net/http"
	"net/url"
)

func WebLogin(context.Context) error {
	// 1) POST `${envBaseUrl}/api/latest/login/token`
	client := &http.Client{}
	baseURL := getBaseApiUrl()

	result, err := getLoginTokenResponse(baseURL, client)
	if err != nil {
		return err
	}

	// Open browser at `${envBaseUrl}/login/token?r=${id}`
	// build browser url
	id := result["id"].(string)
	browserUrl, err := url.JoinPath(baseURL, fmt.Sprintf("login/token?r=", id))
	if err != nil {
		return err
	}
	return utils.OpenBrowser(browserUrl)
}

func getLoginTokenResponse(baseURL string, client *http.Client) (map[string]interface{}, error) {
	urlPath, err := url.JoinPath(baseURL, loginTokenAPI)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", urlPath, bytes.NewBuffer(nil))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 206 {
		return nil, fmt.Errorf("%s", resp.Status)
	}

	var result map[string]interface{}
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetLoginToken(ctx context.Context, code string) (string, error) {
	// GET `/api/latest/login/token/${id}?code=${fourDigitCode}`

	//client := &http.Client{}
	//baseURL := getBaseApiUrl()
	//
	////
	//
	//6) Get back object same shape as 2) with addition of the token
	//
	//	{
	//	  "id": "str_cd8r92r4lvk7115ofr30_21003pb8o07wf13cixfgh9j9i",
	//	  "client_ip": "127.0.0.1",
	//	  "state": "confirmed",
	//	  "created_at": "2022-10-24T12:50:19Z",
	//	  "updated_at": "2022-10-24T12:51:19Z",
	//	  "token": "sptt.eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJsb2NhbGhvc3QiLCJzdWIiOiJ1X2M5dWVzbDkxZXFsa3ZjMHE0NXYwIiwiZXhwIjoxNjY5MjEyNDkxLCJpYXQiOjE2NjY2MjA0OTEsImp0aSI6Ijg5M2JkMzU5LTM3NmYtNDU5OS05NWZmLTdlZTgyMTc2YmJkNSIsInNjb3BlcyI6WyJhZG1pbiJdfQ.n0Txe-ROU1-YkggsyHpLQM8_c4DVRCO-1L9p-3W4t40"
	//	}
	//
	//*/

	return "", nil
}

func SaveToken(ctx context.Context, token string) error {
	return nil
}

func EnsureWorkspace(ctx context.Context, token string) error {
	return nil
}
