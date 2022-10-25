package cloud

import (
	"fmt"
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/utils"
	"net/http"
	"net/url"
	"os"
	"path"
)

// WebLogin POSTs to ${envBaseUrl}/api/latest/login/token to retrieve a login is
// it then opens the login webpage and returns th eid
func WebLogin() (string, error) {
	//  POST `${envBaseUrl}/api/latest/login/token`
	baseURL := getBaseApiUrl()
	client := &http.Client{}
	urlPath, err := url.JoinPath(baseURL, loginIdAPI)
	if err != nil {
		return "", err
	}

	var resp = map[string]any{}
	err = postToAPI(urlPath, "", "", client, &resp)
	if err != nil {
		return "", err
	}

	// Open browser at `${envBaseUrl}/login/token?r=${id}`
	// build browser url
	id := resp["id"].(string)

	browserUrl, err := url.JoinPath(baseURL, webLoginTokenUrl)
	if err != nil {
		return "", err
	}
	// add in id query string
	browserUrl = fmt.Sprintf("%s?r=%s", browserUrl, id)

	fmt.Printf("\nOpening %s\n", browserUrl)
	err = utils.OpenBrowser(browserUrl)
	if err != nil {
		return "", err
	}
	return id, nil

}

// GetLoginToken uses the login id and code and retrieves an authentication token
func GetLoginToken(id, code string) (string, error) {
	// GET `/api/latest/login/token/${id}?code=${fourDigitCode}`
	baseURL := getBaseApiUrl()
	client := &http.Client{}

	getLoginTokenApiPath, err := url.JoinPath(baseURL, fmt.Sprintf(loginTokenAPIFormat, id))
	if err != nil {
		return "", err
	}
	// add in code
	urlPath := fmt.Sprintf("%s?code=%s", getLoginTokenApiPath, code)

	var resp = map[string]any{}
	err = getFromAPI(urlPath, "", client, &resp)
	if err != nil {
		return "", err
	}
	// ensure the result is successful
	if resp["state"].(string) != "confirmed" {
		return "", fmt.Errorf("invalid code")
	}

	token := resp["token"].(string)
	return token, nil
}

// SaveToken writes the token to  ~/.steampipe/internal/{cloud-host}.sptt
func SaveToken(token string) error {
	tokenPath := tokenFilePath(viper.GetString(constants.ArgCloudHost))
	return os.WriteFile(tokenPath, []byte(token), 0600)
}

func LoadToken() (string, error) {
	tokenPath := tokenFilePath(viper.GetString(constants.ArgCloudHost))
	if !filehelpers.FileExists(tokenPath) {
		return "", nil
	}
	tokenBytes, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", fmt.Errorf("failed to load token file '%s': %s", tokenPath, err.Error())
	}
	return string(tokenBytes), nil
}

func GetUserName(token string) (string, error) {
	baseURL := getBaseApiUrl()
	client := &http.Client{}
	bearer := getBearerToken(token)

	actor, err := getActor(baseURL, bearer, client)
	if err != nil {
		return "", err
	}

	return actor.DisplayName, nil
}

func tokenFilePath(cloudHost string) string {
	tokenPath := path.Join(filepaths.EnsureInternalDir(), fmt.Sprintf("%s%s", cloudHost, constants.TokenExtension))
	return tokenPath
}
