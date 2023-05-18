package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	steampipecloud "github.com/turbot/steampipe-cloud-sdk-go"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/sperr"
)

var UnconfirmedError = "Not confirmed"

// WebLogin POSTs to ${envBaseUrl}/api/latest/login/token to retrieve a login is
// it then opens the login webpage and returns th eid
func WebLogin(ctx context.Context) (string, error) {
	client := newSteampipeCloudClient(viper.GetString(constants.ArgCloudToken))

	tempTokenReq, _, err := client.Auth.LoginTokenCreate(ctx).Execute()
	if err != nil {
		return "", sperr.WrapWithMessage(err, "failed to create login token")
	}
	id := tempTokenReq.Id
	// add in id query string
	browserUrl := fmt.Sprintf("%s?r=%s", getLoginTokenConfirmUIUrl(), id)

	fmt.Println()
	fmt.Printf("Verify login at %s\n", browserUrl)

	if err = utils.OpenBrowser(browserUrl); err != nil {
		error_helpers.ShowWarning(fmt.Sprintf("Failed to start browser. Please navigate to %s", constants.Bold(browserUrl)))
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
				return "", sperr.New("%s", body["detail"])
			}
		}
		return "", sperr.Wrap(err)
	}
	if tokenResp.GetToken() == "" && tokenResp.GetState() == "pending" {
		return "", sperr.New("login request has not been confirmed - select 'Verify' and enter the verification code")
	}
	return tokenResp.GetToken(), nil
}

// SaveToken writes the token to  ~/.steampipe/internal/{cloud-host}.sptt
func SaveToken(token string) error {
	tokenPath := tokenFilePath(viper.GetString(constants.ArgCloudHost))
	return sperr.Wrap(os.WriteFile(tokenPath, []byte(token), 0600))
}

func LoadToken() (string, error) {
	tokenPath := tokenFilePath(viper.GetString(constants.ArgCloudHost))
	if !filehelpers.FileExists(tokenPath) {
		return "", nil
	}
	tokenBytes, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", sperr.WrapWithMessage(err, "failed to load token file '%s'", tokenPath)
	}
	return string(tokenBytes), nil
}

func GetUserName(ctx context.Context, token string) (string, error) {
	client := newSteampipeCloudClient(token)
	actor, _, err := client.Actors.Get(ctx).Execute()
	if err != nil {
		return "", sperr.Wrap(err)
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
