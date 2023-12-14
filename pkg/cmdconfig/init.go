package cmdconfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/cloud"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/steampipeconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/steampipe_config_local"
)

// initConfig reads in config file and ENV variables if set.
func initGlobalConfig() *error_helpers.ErrorAndWarnings {
	utils.LogTime("cmdconfig.initGlobalConfig start")
	defer utils.LogTime("cmdconfig.initGlobalConfig end")

	// load workspace profile from the configured install dir
	loader, err := cmdconfig.GetWorkspaceProfileLoader[*modconfig.SteampipeWorkspaceProfile]()
	if err != nil {
		return error_helpers.NewErrorsAndWarning(err)
	}

	// get command out of viper - this will have been set by the root command pre-run
	var cmd = viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command)

	// set-up viper with defaults from the env and default workspace profile
	cmdconfig.BootstrapViper(loader, cmd,
		cmdconfig.WithConfigDefaults(configDefaults),
		cmdconfig.WithDirectoryEnvMappings(dirEnvMappings))

	// set global containing the configured install dir (create directory if needed)
	ensureInstallDir(viper.GetString(constants.ArgInstallDir))

	// load the connection config and HCL options
	config, loadConfigErrorsAndWarnings := steampipe_config_local.LoadSteampipeConfig(viper.GetString(constants.ArgModLocation), cmd.Name())
	if loadConfigErrorsAndWarnings.Error != nil {
		return loadConfigErrorsAndWarnings
	}

	// store global config
	steampipe_config_local.GlobalConfig = config

	// set viper defaults from this config
	cmdconfig.SetDefaultsFromConfig(steampipe_config_local.GlobalConfig.ConfigMap())

	// set the rest of the defaults from ENV
	// ENV takes precedence over any default configuration
	cmdconfig.SetDefaultsFromEnv(envMappings)

	// if an explicit workspace profile was set, add to viper as highest precedence default
	// NOTE: if install_dir/mod_location are set these will already have been passed to viper by BootstrapViper
	// since the "ConfiguredProfile" is passed in through a cmdline flag, it will always take precedence
	if loader.ConfiguredProfile != nil {
		cmdconfig.SetDefaultsFromConfig(loader.ConfiguredProfile.ConfigMap(cmd))
	}

	// NOTE: we need to resolve the token separately
	// - that is because we need the resolved value of ArgCloudHost in order to load any saved token
	// and we cannot get this until the other config has been resolved
	err = setCloudTokenDefault(loader)
	if err != nil {
		loadConfigErrorsAndWarnings.Error = err
		return loadConfigErrorsAndWarnings
	}

	// now validate all config values have appropriate values
	ew := validateConfig()
	error_helpers.FailOnErrorWithMessage(ew.Error, "failed to validate config")

	loadConfigErrorsAndWarnings.Merge(ew)

	return loadConfigErrorsAndWarnings
}

func setCloudTokenDefault(loader *steampipeconfig.WorkspaceProfileLoader[*modconfig.SteampipeWorkspaceProfile]) error {
	/*
	   saved cloud token
	   cloud_token in default workspace
	   explicit env var (STEAMIPE_CLOUD_TOKEN ) wins over
	   cloud_token in specific workspace
	*/
	// set viper defaults in order of increasing precedence
	// 1) saved cloud token
	savedToken, err := cloud.LoadToken()
	if err != nil {
		return err
	}
	if savedToken != "" {
		viper.SetDefault(constants.ArgCloudToken, savedToken)
	}
	// 2) default profile cloud token
	if loader.DefaultProfile.CloudToken != nil {
		viper.SetDefault(constants.ArgCloudToken, *loader.DefaultProfile.CloudToken)
	}
	// 3) env var (STEAMIPE_CLOUD_TOKEN )
	cmdconfig.SetDefaultFromEnv(app_specific.EnvCloudToken, constants.ArgCloudToken, cmdconfig.EnvVarTypeString)

	// 4) explicit workspace profile
	if p := loader.ConfiguredProfile; p != nil && p.CloudToken != nil {
		viper.SetDefault(constants.ArgCloudToken, *p.CloudToken)
	}
	return nil
}

// now validate  config values have appropriate values
// (currently validates telemetry)
func validateConfig() *error_helpers.ErrorAndWarnings {
	var res = &error_helpers.ErrorAndWarnings{}
	telemetry := viper.GetString(constants.ArgTelemetry)
	if !helpers.StringSliceContains(constants.TelemetryLevels, telemetry) {
		res.Error = sperr.New(`invalid value of 'telemetry' (%s), must be one of: %s`, telemetry, strings.Join(constants.TelemetryLevels, ", "))
		return res
	}
	if _, legacyDiagnosticsSet := os.LookupEnv(plugin.EnvLegacyDiagnosticsLevel); legacyDiagnosticsSet {
		res.AddWarning(fmt.Sprintf("Environment variable %s is deprecated - use %s", plugin.EnvLegacyDiagnosticsLevel, plugin.EnvDiagnosticsLevel))
	}
	res.Error = plugin.ValidateDiagnosticsEnvVar()

	return res
}
