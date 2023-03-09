package cloud

import (
	"fmt"
	"net/url"

	"github.com/spf13/viper"
	steampipecloud "github.com/turbot/steampipe-cloud-sdk-go"
	"github.com/turbot/steampipe/pkg/constants"
)

func newSteampipeCloudClient(token string) *steampipecloud.APIClient {
	// Create a default configuration
	configuration := steampipecloud.NewConfiguration()
	configuration.Host = viper.GetString(constants.ArgCloudHost)

	// Add your Steampipe Cloud user token as an auth header
	if token != "" {
		configuration.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	// Create a client
	return steampipecloud.NewAPIClient(configuration)
}

func getLoginTokenConfirmUIUrl() string {
	url := url.URL{
		Scheme: "https",
		Host:   viper.GetString(constants.ArgCloudHost),
		Path:   "/login/token",
	}
	return url.String()
}
