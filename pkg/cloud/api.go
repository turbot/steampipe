package cloud

import (
	"fmt"
	"github.com/turbot/pipe-fittings/constants"
	"net/url"

	"github.com/spf13/viper"
	steampipecloud "github.com/turbot/steampipe-cloud-sdk-go"
)

func newSteampipeCloudClient(token string) *steampipecloud.APIClient {
	// Create a default configuration
	configuration := steampipecloud.NewConfiguration()
	configuration.Host = viper.GetString(constants.ArgPipesHost)

	// Add your Turbot Pipes user token as an auth header
	if token != "" {
		configuration.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	// Create a client
	return steampipecloud.NewAPIClient(configuration)
}

func getLoginTokenConfirmUIUrl() string {
	confirmUrl := url.URL{
		Scheme: "https",
		Host:   viper.GetString(constants.ArgPipesHost),
		Path:   "/login/token",
	}
	return confirmUrl.String()
}
