package cloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	steampipecloud "github.com/turbot/steampipe-cloud-sdk-go"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/utils"
	"os"
	"path"
)

var UnconfirmedError = "Not confirmed"

// WebLogin POSTs to ${envBaseUrl}/api/latest/login/token to retrieve a login is
// it then opens the login webpage and returns th eid
func WebLogin(ctx context.Context) (string, error) {
	client := newSteampipeCloudClient(viper.GetString(constants.ArgCloudToken))

	tempTokenReq, _, err := client.Auth.LoginTokenCreate(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create login token: %s", err)
	}
	id := tempTokenReq.Id
	// add in id query string
	browserUrl := fmt.Sprintf("%s?r=%s", getLoginTokenConfirmUIUrl(), id)

	fmt.Println()
	fmt.Printf("Verify login at %s\n", browserUrl)
	err = utils.OpenBrowser(browserUrl)
	if err != nil {
		return "", fmt.Errorf("failed to open login webpage: %s", err)
	}
	return id, nil

}

// GetLoginToken uses the login id and code and retrieves an authentication token
func GetLoginToken(ctx context.Context, id, code string) (string, error) {
	client := newSteampipeCloudClient("")
	tokenResp, _, err := client.Auth.LoginTokenGet(ctx, id).Code(code).Execute()
	if err != nil {
		if apiErr, ok := err.(steampipecloud.GenericOpenAPIError); ok {
			var body = map[string]any{}
			if err := json.Unmarshal(apiErr.Body(), &body); err == nil {
				return "", errors.New(body["detail"].(string))
			}
		}
		return "", err
	}
	if tokenResp.GetToken() == "" && tokenResp.GetState() == "pending" {
		return "", fmt.Errorf("login request has not been confirmed - select 'Verify' and enter the verification code")
	}
	return tokenResp.GetToken(), nil
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
	if name, ok := actor.GetDisplayNameOk(); ok {
		return *name
	}
	return actor.Handle
}

func tokenFilePath(cloudHost string) string {
	tokenPath := path.Join(filepaths.EnsureInternalDir(), fmt.Sprintf("%s%s", cloudHost, constants.TokenExtension))
	return tokenPath
}
