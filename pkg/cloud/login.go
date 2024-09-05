package cloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	constants2 "github.com/turbot/pipe-fittings/constants"
	"log"
	"os"
	"path"

	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/utils"
	steampipecloud "github.com/turbot/steampipe-cloud-sdk-go"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
)

var UnconfirmedError = "Not confirmed"

// WebLogin POSTs to ${envBaseUrl}/api/latest/login/token to retrieve a login is
// it then opens the login webpage and returns th eid
func WebLogin(ctx context.Context) (string, error) {
	client := newSteampipeCloudClient(viper.GetString(constants2.ArgPipesToken))

	tempTokenReq, _, err := client.Auth.LoginTokenCreate(ctx).Execute()
	if err != nil {
		return "", sperr.WrapWithMessage(err, "failed to create login token")
	}
	id := tempTokenReq.Id
	// add in id query string
	browserUrl := fmt.Sprintf("%s?r=%s", getLoginTokenConfirmUIUrl(), id)

	fmt.Printf("\nVerify login at %s\n", browserUrl)

	if err = utils.OpenBrowser(browserUrl); err != nil {
		log.Println("[INFO] failed to open login web page")
	}

	return id, nil
}

// GetLoginToken uses the login id and code and retrieves an authentication token
func GetLoginToken(ctx context.Context, id, code string) (string, error) {
	client := newSteampipeCloudClient("")
	tokenResp, _, err := client.Auth.LoginTokenGet(ctx, id).Code(code).Execute()
	if err != nil {
		var apiErr steampipecloud.GenericOpenAPIError
		if errors.As(err, &apiErr) {
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

// SaveToken writes the token to  ~/.steampipe/internal/{cloud-host}.tptt
func SaveToken(token string) error {
	tokenPath := tokenFilePath(viper.GetString(constants2.ArgPipesHost))
	return sperr.Wrap(os.WriteFile(tokenPath, []byte(token), 0600))
}

func LoadToken() (string, error) {
	if err := migrateDefaultTokenFile(); err != nil {
		log.Println("[TRACE] ERROR during migrating token file", err)
	}
	tokenPath := tokenFilePath(viper.GetString(constants2.ArgPipesHost))
	if !filehelpers.FileExists(tokenPath) {
		return "", nil
	}
	tokenBytes, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", sperr.WrapWithMessage(err, "failed to load token file '%s'", tokenPath)
	}
	return string(tokenBytes), nil
}

// migrateDefaultTokenFile migrates the cloud.steampipe.io.sptt token file
// to the pipes.turbot.com.tptt token file
// it also migrates the token file from the	~/.steampipe/internal directory to the ~/.pipes/internal directory
func migrateDefaultTokenFile() error {
	defaultTokenPath := tokenFilePath(constants.DefaultPipesHost)
	defaultLegacyTokenPaths := legacyTokenFilePaths()

	tokenExists := filehelpers.FileExists(defaultTokenPath)

	for _, legacyPath := range defaultLegacyTokenPaths {
		if filehelpers.FileExists(legacyPath) {
			if tokenExists {
				// try removing the old legacy file - no worries if os.Remove fails
				_ = os.Remove(legacyPath)
			} else {
				if err := utils.MoveFile(legacyPath, defaultTokenPath); err != nil {
					return err
				}
				// set token exists flag so any other legacy files are removed (we do not expect any more)
				tokenExists = true
			}
		}
	}

	return nil
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

func tokenFilePath(pipesHost string) string {
	tokenPath := path.Join(filepaths.EnsurePipesInternalDir(), fmt.Sprintf("%s%s", pipesHost, constants.TokenExtension))
	return tokenPath
}

func legacyTokenFilePaths() []string {
	return []string{path.Join(filepaths.EnsureInternalDir(), fmt.Sprintf("%s%s", constants.LegacyDefaultPipesHost, constants.LegacyTokenExtension)),
		path.Join(filepaths.EnsureInternalDir(), fmt.Sprintf("%s%s", constants.DefaultPipesHost, constants.TokenExtension))}
}
