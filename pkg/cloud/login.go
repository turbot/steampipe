package cloud

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	steampipecloud "github.com/turbot/steampipe-cloud-sdk-go"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/utils"
	"net/url"
	"os"
	"path"
)

// WebLogin POSTs to ${envBaseUrl}/api/latest/login/token to retrieve a login is
// it then opens the login webpage and returns th eid
func WebLogin(ctx context.Context) (string, error) {
	client := newSteampipeCloudClient(viper.GetString(constants.ArgCloudToken))

	tempTokenReq, _, err := client.Auth.LoginTokenCreate(ctx).Execute()
	if err != nil {
		return "", err
	}
	id := tempTokenReq.Id
	browserUrl, err := url.JoinPath(getBaseApiUrl(), id)
	if err != nil {
		return "", err
	}
	// add in id query string
	browserUrl = fmt.Sprintf("%s?r=%s", browserUrl, tempTokenReq)

	fmt.Printf("Opening %s\n", browserUrl)
	err = utils.OpenBrowser(browserUrl)
	if err != nil {
		return "", err
	}
	return id, nil

}

// GetLoginToken uses the login id and code and retrieves an authentication token
func GetLoginToken(id, code string) (string, error) {
	//// GET `/api/latest/login/token/${id}?code=${fourDigitCode}`
	//baseURL := getBaseApiUrl()
	//client := &http.Client{}
	//
	//getLoginTokenApiPath, err := url.JoinPath(baseURL, fmt.Sprintf(loginTokenAPIFormat, id))
	//if err != nil {
	//	return "", err
	//}
	//// add in code
	//urlPath := fmt.Sprintf("%s?code=%s", getLoginTokenApiPath, code)
	//
	//var resp = map[string]any{}
	//err = getFromAPI(urlPath, "", client, &resp)
	//if err != nil {
	//	return "", err
	//}
	//// ensure the result is successful
	//if resp["state"].(string) != "confirmed" {
	//	return "", fmt.Errorf("invalid code")
	//}
	//
	//token := resp["token"].(string)
	var token string
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

func GetUserName(ctx context.Context, token string) (string, error) {
	client := newSteampipeCloudClient(token)
	actor, _, err := client.Actors.Get(ctx).Execute()
	if err != nil {
		return "", err
	}
	return getActorName(actor), nil
}

func getActorName(actor steampipecloud.User) string {
	if actor.DisplayName != nil {
		return *actor.DisplayName
	}
	return actor.Handle
}

func tokenFilePath(cloudHost string) string {
	tokenPath := path.Join(filepaths.EnsureInternalDir(), fmt.Sprintf("%s%s", cloudHost, constants.TokenExtension))
	return tokenPath
}
