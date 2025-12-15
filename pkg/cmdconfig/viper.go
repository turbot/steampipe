package cmdconfig

import (
	"fmt"
	"log"
	"os"
	"sync"

	pfilepaths "github.com/turbot/pipe-fittings/v2/filepaths"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/types"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/parse"
	"github.com/turbot/pipe-fittings/v2/workspace_profile"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

// viperMutex protects concurrent access to Viper's global state
var viperMutex sync.RWMutex

// Viper fetches the global viper instance
func Viper() *viper.Viper {
	return viper.GetViper()
}

// bootstrapViper sets up viper with the essential path config (workspace-chdir and install-dir)
func bootstrapViper(loader *parse.WorkspaceProfileLoader[*workspace_profile.SteampipeWorkspaceProfile], cmd *cobra.Command) error {
	// set defaults  for keys which do not have a corresponding command flag
	if err := setBaseDefaults(); err != nil {
		return err
	}

	// set defaults from defaultWorkspaceProfile
	SetDefaultsFromConfig(loader.DefaultProfile.ConfigMap(cmd))

	// set defaults for install dir and mod location from env vars
	// this needs to be done since the workspace profile definitions exist in the
	// default install dir
	setDirectoryDefaultsFromEnv()

	// NOTE: if an explicit workspace profile was set, default the install dir _now_
	// All other workspace profile values are defaults _after defaulting to the connection config options
	// to give them higher precedence, but these must be done now as subsequent operations depend on them
	// (and they cannot be set from hcl options)
	if loader.ConfiguredProfile != nil {
		if loader.ConfiguredProfile.InstallDir != nil {
			log.Printf("[TRACE] setting install dir from configured profile '%s' to '%s'", loader.ConfiguredProfile.Name(), *loader.ConfiguredProfile.InstallDir)
			viperMutex.Lock()
			viper.SetDefault(pconstants.ArgInstallDir, *loader.ConfiguredProfile.InstallDir)
			viperMutex.Unlock()
		}
	}

	// tildefy all paths in viper
	return tildefyPaths()
}

// tildefyPaths cleans all path config values and replaces '~' with the home directory
func tildefyPaths() error {
	pathArgs := []string{
		pconstants.ArgModLocation,
		pconstants.ArgInstallDir,
	}
	var err error
	for _, argName := range pathArgs {
		viperMutex.RLock()
		argVal := viper.GetString(argName)
		isSet := viper.IsSet(argName)
		viperMutex.RUnlock()

		if argVal != "" {
			if argVal, err = filehelpers.Tildefy(argVal); err != nil {
				return err
			}
			viperMutex.Lock()
			if isSet {
				// if the value was already set re-set
				viper.Set(argName, argVal)
			} else {
				// otherwise just update the default
				viper.SetDefault(argName, argVal)
			}
			viperMutex.Unlock()
		}
	}
	return nil
}

// SetDefaultsFromConfig overrides viper default values from hcl config values
func SetDefaultsFromConfig(configMap map[string]interface{}) {
	viperMutex.Lock()
	defer viperMutex.Unlock()
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
func setBaseDefaults() error {
	defaults := map[string]interface{}{
		// global general options
		pconstants.ArgTelemetry:       constants.TelemetryInfo,
		pconstants.ArgUpdateCheck:     true,
		pconstants.ArgPipesInstallDir: pfilepaths.DefaultPipesInstallDir,

		// workspace profile
		pconstants.ArgAutoComplete: true,

		// from global database options
		pconstants.ArgDatabasePort:         constants.DatabaseDefaultPort,
		pconstants.ArgDatabaseStartTimeout: constants.DBStartTimeout.Seconds(),
		pconstants.ArgServiceCacheEnabled:  true,
		pconstants.ArgCacheMaxTtl:          300,

		// dashboard
		pconstants.ArgDashboardStartTimeout: constants.DashboardStartTimeout.Seconds(),

		// memory
		pconstants.ArgMemoryMaxMbPlugin: 1024,
		pconstants.ArgMemoryMaxMb:       1024,

		// plugin start timeout
		pconstants.ArgPluginStartTimeout: constants.PluginStartTimeout.Seconds(),
	}

	viperMutex.Lock()
	defer viperMutex.Unlock()
	for k, v := range defaults {
		viper.SetDefault(k, v)
	}
	return nil
}

type envMapping struct {
	configVar []string
	varType   EnvVarType
}

// set default values of INSTALL_DIR and ModLocation from env vars
func setDirectoryDefaultsFromEnv() {
	envMappings := map[string]envMapping{
		constants.EnvInstallDir:     {[]string{pconstants.ArgInstallDir}, String},
		constants.EnvWorkspaceChDir: {[]string{pconstants.ArgModLocation}, String},
	}

	for envVar, mapping := range envMappings {
		setConfigFromEnv(envVar, mapping.configVar, mapping.varType)
	}
}

// setDefaultsFromEnv sets default values from env vars
func setDefaultsFromEnv() {
	// NOTE: EnvWorkspaceProfile has already been set as a viper default as we have already loaded workspace profiles
	// (EnvInstallDir has already been set at same time but we set it again to make sure it has the correct precedence)

	// a map of known environment variables to map to viper keys
	envMappings := map[string]envMapping{
		constants.EnvInstallDir:            {[]string{pconstants.ArgInstallDir}, String},
		constants.EnvWorkspaceChDir:        {[]string{pconstants.ArgModLocation}, String},
		constants.EnvTelemetry:             {[]string{pconstants.ArgTelemetry}, String},
		constants.EnvUpdateCheck:           {[]string{pconstants.ArgUpdateCheck}, Bool},
		constants.EnvPipesHost:             {[]string{pconstants.ArgPipesHost}, String},
		constants.EnvPipesToken:            {[]string{pconstants.ArgPipesToken}, String},
		constants.EnvPipesInstallDir:       {[]string{pconstants.ArgPipesInstallDir}, String},
		constants.EnvSnapshotLocation:      {[]string{pconstants.ArgSnapshotLocation}, String},
		constants.EnvWorkspaceDatabase:     {[]string{pconstants.ArgWorkspaceDatabase}, String},
		constants.EnvServicePassword:       {[]string{pconstants.ArgServicePassword}, String},
		constants.EnvDisplayWidth:          {[]string{pconstants.ArgDisplayWidth}, Int},
		constants.EnvMaxParallel:           {[]string{pconstants.ArgMaxParallel}, Int},
		constants.EnvQueryTimeout:          {[]string{pconstants.ArgDatabaseQueryTimeout}, Int},
		constants.EnvDatabaseStartTimeout:  {[]string{pconstants.ArgDatabaseStartTimeout}, Int},
		constants.EnvDatabaseSSLPassword:   {[]string{pconstants.ArgDatabaseSSLPassword}, String},
		constants.EnvDashboardStartTimeout: {[]string{pconstants.ArgDashboardStartTimeout}, Int},
		constants.EnvCacheTTL:              {[]string{pconstants.ArgCacheTtl}, Int},
		constants.EnvCacheMaxTTL:           {[]string{pconstants.ArgCacheMaxTtl}, Int},
		constants.EnvMemoryMaxMb:           {[]string{pconstants.ArgMemoryMaxMb}, Int},
		constants.EnvMemoryMaxMbPlugin:     {[]string{pconstants.ArgMemoryMaxMbPlugin}, Int},
		constants.EnvPluginStartTimeout:    {[]string{pconstants.ArgPluginStartTimeout}, Int},

		// we need this value to go into different locations
		constants.EnvCacheEnabled: {[]string{
			pconstants.ArgClientCacheEnabled,
			pconstants.ArgServiceCacheEnabled,
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
		viperMutex.Lock()
		defer viperMutex.Unlock()
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
