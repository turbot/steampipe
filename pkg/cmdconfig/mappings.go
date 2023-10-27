package localcmdconfig

import (
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
)

var configDefaults = map[string]any{
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
	constants.ArgMaxCacheSizeMb:       constants.DefaultMaxCacheSizeMb,
}

// environment variable mappings for directory paths which must be set as part of the viper bootstrap process
var dirEnvMappings = map[string]cmdconfig.EnvMapping{
	constants.EnvInstallDir:     {[]string{constants.ArgInstallDir}, cmdconfig.EnvVarTypeString},
	constants.EnvWorkspaceChDir: {[]string{constants.ArgModLocation}, cmdconfig.EnvVarTypeString},
	constants.EnvModLocation:    {[]string{constants.ArgModLocation}, cmdconfig.EnvVarTypeString},
}

// NOTE: EnvWorkspaceProfile has already been set as a viper default as we have already loaded workspace profiles
// (EnvInstallDir has already been set at same time but we set it again to make sure it has the correct precedence)

// a map of known environment variables to map to viper keys - these are set as part of LoadGlobalConfig
var envMappings = map[string]cmdconfig.EnvMapping{
	constants.EnvInstallDir:     {[]string{constants.ArgInstallDir}, cmdconfig.EnvVarTypeString},
	constants.EnvWorkspaceChDir: {[]string{constants.ArgModLocation}, cmdconfig.EnvVarTypeString},
	constants.EnvModLocation:    {[]string{constants.ArgModLocation}, cmdconfig.EnvVarTypeString},
	constants.EnvIntrospection:  {[]string{constants.ArgIntrospection}, cmdconfig.EnvVarTypeString},
	constants.EnvTelemetry:      {[]string{constants.ArgTelemetry}, cmdconfig.EnvVarTypeString},
	constants.EnvUpdateCheck:    {[]string{constants.ArgUpdateCheck}, cmdconfig.EnvVarTypeBool},
	// EnvPipesHost needs to be defined before EnvCloudHost,
	// so that if EnvCloudHost is defined, it can override EnvPipesHost
	constants.EnvPipesHost: {[]string{constants.ArgCloudHost}, cmdconfig.EnvVarTypeString},
	constants.EnvCloudHost: {[]string{constants.ArgCloudHost}, cmdconfig.EnvVarTypeString},
	// EnvPipesToken needs to be defined before EnvCloudToken,
	// so that if EnvCloudToken is defined, it can override EnvPipesToken
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
