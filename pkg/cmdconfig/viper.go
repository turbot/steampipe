package cmdconfig

import (
	"fmt"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"log"
	"os"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/constants"
)

// Viper fetches the global viper instance
func Viper() *viper.Viper {
	return viper.GetViper()
}

// BootstrapViper sets up viper with the essential path config (workspace-chdir and install-dir)
func BootstrapViper(loader *steampipeconfig.WorkspaceProfileLoader) error {
	// set defaults  for keys which do not have a corresponding command flag
	setBaseDefaults()

	// set defaults from defaultWorkspaceProfile
	SetDefaultsFromConfig(loader.DefaultProfile.ConfigMap())

	// set defaults from env vars
	setDefaultsFromEnv()

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
func setBaseDefaults() {
	defaults := map[string]interface{}{
		constants.ArgUpdateCheck:    true,
		constants.ArgTelemetry:      constants.TelemetryInfo,
		constants.ArgDatabasePort:   constants.DatabaseDefaultPort,
		constants.ArgMaxCacheSizeMb: constants.DefaultMaxCacheSizeMb,
		constants.ArgAutoComplete:   true,
	}

	for k, v := range defaults {
		viper.SetDefault(k, v)
	}
}

type envMapping struct {
	configVar string
	// "string", "int", "bool"
	varType string
}

// set default values from env vars
func setDefaultsFromEnv() {
	// NOTE: EnvWorkspaceProfile has already been set as a viper default as we have already loaded workspace profiles
	// (EnvInstallDir has already been set at same time but we set it again to make sure it has the correct precedence)

	// a map of known environment variables to map to viper keys
	envMappings := map[string]envMapping{
		constants.EnvInstallDir:        {constants.ArgInstallDir, "string"},
		constants.EnvWorkspaceChDir:    {constants.ArgModLocation, "string"},
		constants.EnvModLocation:       {constants.ArgModLocation, "string"},
		constants.EnvIntrospection:     {constants.ArgIntrospection, "string"},
		constants.EnvTelemetry:         {constants.ArgTelemetry, "string"},
		constants.EnvUpdateCheck:       {constants.ArgUpdateCheck, "bool"},
		constants.EnvCloudHost:         {constants.ArgCloudHost, "string"},
		constants.EnvCloudToken:        {constants.ArgCloudToken, "string"},
		constants.EnvSnapshotLocation:  {constants.ArgSnapshotLocation, "string"},
		constants.EnvWorkspaceDatabase: {constants.ArgWorkspaceDatabase, "string"},
		constants.EnvServicePassword:   {constants.ArgServicePassword, "string"},
		constants.EnvCheckDisplayWidth: {constants.ArgCheckDisplayWidth, "int"},
		constants.EnvMaxParallel:       {constants.ArgMaxParallel, "int"},
		constants.EnvQueryTimeout:      {constants.ArgDatabaseQueryTimeout, "int"},
	}

	for k, v := range envMappings {
		SetDefaultFromEnv(k, v.configVar, v.varType)
	}
}

func SetDefaultFromEnv(k string, configVar string, varType string) {
	if val, ok := os.LookupEnv(k); ok {
		switch varType {
		case "string":
			viper.SetDefault(configVar, val)
		case "bool":
			if boolVal, err := types.ToBool(val); err == nil {
				viper.SetDefault(configVar, boolVal)
			}
		case "int":
			if intVal, err := types.ToInt64(val); err == nil {
				viper.SetDefault(configVar, intVal)
			}
		default:
			// must be an invalid value in the map above
			panic(fmt.Sprintf("invalid env var mapping type: %s", varType))
		}
	}
}
