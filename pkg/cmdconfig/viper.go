package cmdconfig

import (
	"fmt"
	"log"
	"os"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/steampipeconfig"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/constants"
)

// Viper fetches the global viper instance
func Viper() *viper.Viper {
	return viper.GetViper()
}

// BootstrapViper sets up viper with the essential path config (workspace-chdir and install-dir)
func BootstrapViper(loader *steampipeconfig.WorkspaceProfileLoader, cmd *cobra.Command) error {
	// set defaults  for keys which do not have a corresponding command flag
	setBaseDefaults()

	// set defaults from defaultWorkspaceProfile
	SetDefaultsFromConfig(loader.DefaultProfile.ConfigMap(cmd))

	// set defaults for install dir and mod location from env vars
	// this needs to be done since the workspace profile definitions exist in the
	// default install dir
	setDirectoryDefaultsFromEnv()

	// NOTE: if an explicit workspace profile was set, default the mod location and install dir _now_
	// All other workspace profile values are defaults _after defaulting to the connection config options
	// to give them higher precedence, but these must be done now as subsequent operations depend on them
	// (and they cannot be set from hcl options)
	if loader.ConfiguredProfile != nil {
		if loader.ConfiguredProfile.ModLocation != nil {
			log.Printf("[TRACE] setting mod location from configured profile '%s' to '%s'", loader.ConfiguredProfile.Name(), *loader.ConfiguredProfile.ModLocation)
			viper.SetDefault(constants.ArgModLocation, *loader.ConfiguredProfile.ModLocation)
		}
		if loader.ConfiguredProfile.InstallDir != nil {
			log.Printf("[TRACE] setting install dir from configured profile '%s' to '%s'", loader.ConfiguredProfile.Name(), *loader.ConfiguredProfile.InstallDir)
			viper.SetDefault(constants.ArgInstallDir, *loader.ConfiguredProfile.InstallDir)
		}
	}

	// tildefy all paths in viper
	return tildefyPaths()
}

// tildefyPaths cleans all path config values and replaces '~' with the home directory
func tildefyPaths() error {
	pathArgs := []string{
		constants.ArgModLocation,
		constants.ArgInstallDir,
	}
	var err error
	for _, argName := range pathArgs {
		if argVal := viper.GetString(argName); argVal != "" {
			if argVal, err = filehelpers.Tildefy(argVal); err != nil {
				return err
			}
			if viper.IsSet(argName) {
				// if the value was already set re-set
				viper.Set(argName, argVal)
			} else {
				// otherwise just update the default
				viper.SetDefault(argName, argVal)
			}
		}
	}
	return nil
}

// SetDefaultsFromConfig overrides viper default values from hcl config values
func SetDefaultsFromConfig(configMap map[string]interface{}) {
	for k, v := range configMap {
		viper.SetDefault(k, v)
	}
}

// for keys which do not have a corresponding command flag, we need a separate defaulting mechanism
// any option setting, workspace profile property or env var which does not have a command line
// MUST have a default (unless we want the zero value to take effect)
//
// Do not add keys here which have command line defaults - the way this is setup, this value takes
// precedence over command line default
func setBaseDefaults() {
	defaults := map[string]interface{}{
		// global general options
		constants.ArgTelemetry:   constants.TelemetryInfo,
		constants.ArgUpdateCheck: true,

		// workspace profile
		constants.ArgAutoComplete:  true,
		constants.ArgIntrospection: constants.IntrospectionNone,

		// from global database options
		constants.ArgDatabasePort:         constants.DatabaseDefaultPort,
		constants.ArgDatabaseStartTimeout: constants.DBStartTimeout.Seconds(),
		constants.ArgServiceCacheEnabled:  true,
		constants.ArgCacheMaxTtl:          300,
	}

	for k, v := range defaults {
		viper.SetDefault(k, v)
	}
}

type envMapping struct {
	configVar []string
	varType   EnvVarType
}

// set default values of INSTALL_DIR and ModLocation from env vars
func setDirectoryDefaultsFromEnv() {
	envMappings := map[string]envMapping{
		constants.EnvInstallDir:     {[]string{constants.ArgInstallDir}, String},
		constants.EnvWorkspaceChDir: {[]string{constants.ArgModLocation}, String},
		constants.EnvModLocation:    {[]string{constants.ArgModLocation}, String},
	}

	for envVar, mapping := range envMappings {
		setConfigFromEnv(envVar, mapping.configVar, mapping.varType)
	}
}

// set default values from env vars
func SetDefaultsFromEnv() {
	// NOTE: EnvWorkspaceProfile has already been set as a viper default as we have already loaded workspace profiles
	// (EnvInstallDir has already been set at same time but we set it again to make sure it has the correct precedence)

	// a map of known environment variables to map to viper keys
	envMappings := map[string]envMapping{
		constants.EnvInstallDir:           {[]string{constants.ArgInstallDir}, String},
		constants.EnvWorkspaceChDir:       {[]string{constants.ArgModLocation}, String},
		constants.EnvModLocation:          {[]string{constants.ArgModLocation}, String},
		constants.EnvIntrospection:        {[]string{constants.ArgIntrospection}, String},
		constants.EnvTelemetry:            {[]string{constants.ArgTelemetry}, String},
		constants.EnvUpdateCheck:          {[]string{constants.ArgUpdateCheck}, Bool},
		constants.EnvCloudHost:            {[]string{constants.ArgCloudHost}, String},
		constants.EnvCloudToken:           {[]string{constants.ArgCloudToken}, String},
		constants.EnvSnapshotLocation:     {[]string{constants.ArgSnapshotLocation}, String},
		constants.EnvWorkspaceDatabase:    {[]string{constants.ArgWorkspaceDatabase}, String},
		constants.EnvServicePassword:      {[]string{constants.ArgServicePassword}, String},
		constants.EnvDisplayWidth:         {[]string{constants.ArgDisplayWidth}, Int},
		constants.EnvMaxParallel:          {[]string{constants.ArgMaxParallel}, Int},
		constants.EnvQueryTimeout:         {[]string{constants.ArgDatabaseQueryTimeout}, Int},
		constants.EnvDatabaseStartTimeout: {[]string{constants.ArgDatabaseStartTimeout}, Int},
		constants.EnvCacheTTL:             {[]string{constants.ArgCacheTtl}, Int},
		constants.EnvCacheMaxTTL:          {[]string{constants.ArgCacheMaxTtl}, Int},

		// we need this value to go into different locations
		constants.EnvCacheEnabled: {[]string{
			constants.ArgClientCacheEnabled,
			constants.ArgServiceCacheEnabled,
		}, Bool},
	}

	for envVar, v := range envMappings {
		setConfigFromEnv(envVar, v.configVar, v.varType)
	}
}

func setConfigFromEnv(envVar string, configs []string, varType EnvVarType) {
	for _, configVar := range configs {
		SetDefaultFromEnv(envVar, configVar, varType)
	}
}

func SetDefaultFromEnv(k string, configVar string, varType EnvVarType) {
	if val, ok := os.LookupEnv(k); ok {
		switch varType {
		case String:
			viper.SetDefault(configVar, val)
		case Bool:
			if boolVal, err := types.ToBool(val); err == nil {
				viper.SetDefault(configVar, boolVal)
			}
		case Int:
			if intVal, err := types.ToInt64(val); err == nil {
				viper.SetDefault(configVar, intVal)
			}
		default:
			// must be an invalid value in the map above
			panic(fmt.Sprintf("invalid env var mapping type: %s", varType))
		}
	}
}
