package cmdconfig

import (
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/viper"
	pfilepaths "github.com/turbot/pipe-fittings/v2/filepaths"

	"github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

// SetAppSpecificConstants sets app specific constants defined in pipe-fittings
func SetAppSpecificConstants() {
	app_specific.AppName = "steampipe"

	// set an initial value for the version
	initialVersion := "0.0.0"

	versionString := viper.GetString("main.version")

	// check if the version is set in viper, otherwise use the initial value
	// this is required since when the FDW is initialized SetAppSpecificConstants is called, at that time
	// the viper config will have not been initialized yet and the version will not be set, which will cause
	// semver.MustParse to panic
	if versionString == "" {
		versionString = initialVersion
	} else {
		app_specific.AppVersion = semver.MustParse(versionString)
	}

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
	defaultPipesInstallDir, err := files.Tildefy("~/.pipes")
	pfilepaths.DefaultPipesInstallDir = defaultPipesInstallDir
	error_helpers.FailOnError(err)

	// check whether install-dir env has been set - if so, respect it
	if envInstallDir, ok := os.LookupEnv(app_specific.EnvInstallDir); ok {
		app_specific.InstallDir = envInstallDir
	} else {
		// NOTE: install dir will be set to configured value at the end of InitGlobalConfig
		app_specific.InstallDir = defaultInstallDir
	}

	// ociinstaller
	app_specific.DefaultImageRepoActualURL = "ghcr.io/turbot/steampipe"
	app_specific.DefaultImageRepoDisplayURL = "hub.steampipe.io"

}
