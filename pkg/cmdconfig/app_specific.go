package cmdconfig

import (
	"os"

	"github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/version"
)

// SetAppSpecificConstants sets app specific constants defined in pipe-fittings
func SetAppSpecificConstants() {
	app_specific.AppName = "steampipe"

	app_specific.AppVersion = version.SteampipeVersion

	app_specific.SetAppSpecificEnvVarKeys("STEAMPIPE_")
	app_specific.ConfigExtension = ".spc"
	app_specific.PluginHub = constants.SteampipeHubOCIBase

	// Version check
	app_specific.VersionCheckHost = "hub.steampipe.io"
	app_specific.VersionCheckPath = "api/cli/version/latest"

	// set the default install dir
	defaultInstallDir, err := files.Tildefy("~/.steampipe")
	error_helpers.FailOnError(err)
	app_specific.DefaultInstallDir = defaultInstallDir

	// check whether install-dir env has been set - if so, respect it
	if envInstallDir, ok := os.LookupEnv(app_specific.EnvInstallDir); ok {
		app_specific.InstallDir = envInstallDir
	} else {
		// NOTE: install dir will be set to configured value at the end of InitGlobalConfig
		app_specific.InstallDir = defaultInstallDir
	}

}
