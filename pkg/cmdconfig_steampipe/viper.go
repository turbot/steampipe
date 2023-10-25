package cmdconfig_steampipe

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/steampipeconfig"
	"log"
)

// TODO kai this also exists in pipe-fittings - sort out
// bootstrapViper sets up viper with the essential path config (workspace-chdir and install-dir)
func bootstrapViper(loader *steampipeconfig.WorkspaceProfileLoader, cmd *cobra.Command) error {
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
	return cmdconfig.TildefyPaths()
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

		// dashboard
		constants.ArgDashboardStartTimeout: constants.DashboardServiceStartTimeout.Seconds(),

		// memory
		constants.ArgMemoryMaxMbPlugin: 1024,
		constants.ArgMemoryMaxMb:       1024,
	}

	for k, v := range defaults {
		viper.SetDefault(k, v)
	}
}

type envMapping struct {
	configVar []string
	varType   cmdconfig.EnvVarType
}

// set default values of INSTALL_DIR and ModLocation from env vars
func setDirectoryDefaultsFromEnv() {
	envMappings := map[string]envMapping{
		constants.EnvInstallDir:     {[]string{constants.ArgInstallDir}, cmdconfig.EnvVarTypeString},
		constants.EnvWorkspaceChDir: {[]string{constants.ArgModLocation}, cmdconfig.EnvVarTypeString},
		constants.EnvModLocation:    {[]string{constants.ArgModLocation}, cmdconfig.EnvVarTypeString},
	}

	for envVar, mapping := range envMappings {
		cmdconfig.SetConfigFromEnv(envVar, mapping.configVar, mapping.varType)
	}
}

// setDefaultsFromEnv sets default values from env vars
func setDefaultsFromEnv() {
	// NOTE: EnvWorkspaceProfile has already been set as a viper default as we have already loaded workspace profiles
	// (EnvInstallDir has already been set at same time but we set it again to make sure it has the correct precedence)

	// a map of known environment variables to map to viper keys
	envMappings := map[string]envMapping{
		constants.EnvInstallDir:     {[]string{constants.ArgInstallDir}, cmdconfig.EnvVarTypeString},
		constants.EnvWorkspaceChDir: {[]string{constants.ArgModLocation}, cmdconfig.EnvVarTypeString},
		constants.EnvModLocation:    {[]string{constants.ArgModLocation}, cmdconfig.EnvVarTypeString},
		constants.EnvIntrospection:  {[]string{constants.ArgIntrospection}, cmdconfig.EnvVarTypeString},
		constants.EnvTelemetry:      {[]string{constants.ArgTelemetry}, cmdconfig.EnvVarTypeString},
		constants.EnvUpdateCheck:    {[]string{constants.ArgUpdateCheck}, cmdconfig.EnvVarTypeBool},
		// PIPES_HOST needs to be defined before STEAMPIPE_CLOUD_HOST,
		// so that if STEAMPIPE_CLOUD_HOST is defined, it can override PIPES_HOST
		constants.EnvPipesHost: {[]string{constants.ArgCloudHost}, cmdconfig.EnvVarTypeString},
		constants.EnvCloudHost: {[]string{constants.ArgCloudHost}, cmdconfig.EnvVarTypeString},
		// PIPES_TOKEN needs to be defined before STEAMPIPE_CLOUD_TOKEN,
		// so that if STEAMPIPE_CLOUD_TOKEN is defined, it can override PIPES_TOKEN
		constants.EnvPipesToken: {[]string{constants.ArgCloudToken}, cmdconfig.EnvVarTypeString},
		constants.EnvCloudToken: {[]string{constants.ArgCloudToken}, cmdconfig.EnvVarTypeString},
		//
		constants.EnvSnapshotLocation:      {[]string{constants.ArgSnapshotLocation}, cmdconfig.EnvVarTypeString},
		constants.EnvWorkspaceDatabase:     {[]string{constants.ArgWorkspaceDatabase}, cmdconfig.EnvVarTypeString},
		constants.EnvServicePassword:       {[]string{constants.ArgServicePassword}, cmdconfig.EnvVarTypeString},
		constants.EnvDisplayWidth:          {[]string{constants.ArgDisplayWidth}, cmdconfig.EnvVarTypeInt},
		constants.EnvMaxParallel:           {[]string{constants.ArgMaxParallel}, cmdconfig.EnvVarTypeInt},
		constants.EnvQueryTimeout:          {[]string{constants.ArgDatabaseQueryTimeout}, cmdconfig.EnvVarTypeInt},
		constants.EnvDatabaseStartTimeout:  {[]string{constants.ArgDatabaseStartTimeout}, cmdconfig.EnvVarTypeInt},
		constants.EnvDashboardStartTimeout: {[]string{constants.ArgDashboardStartTimeout}, cmdconfig.EnvVarTypeInt},
		constants.EnvCacheTTL:              {[]string{constants.ArgCacheTtl}, cmdconfig.EnvVarTypeInt},
		constants.EnvCacheMaxTTL:           {[]string{constants.ArgCacheMaxTtl}, cmdconfig.EnvVarTypeInt},
		constants.EnvMemoryMaxMb:           {[]string{constants.ArgMemoryMaxMb}, cmdconfig.EnvVarTypeInt},
		constants.EnvMemoryMaxMbPlugin:     {[]string{constants.ArgMemoryMaxMbPlugin}, cmdconfig.EnvVarTypeInt},

		// we need this value to go into different locations
		constants.EnvCacheEnabled: {[]string{
			constants.ArgClientCacheEnabled,
			constants.ArgServiceCacheEnabled,
		}, cmdconfig.EnvVarTypeBool},
	}

	for envVar, v := range envMappings {
		cmdconfig.SetConfigFromEnv(envVar, v.configVar, v.varType)
	}
}
