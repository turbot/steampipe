package cmdconfig

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
	constants.EnvInstallDir:  {ConfigVar: []string{constants.ArgInstallDir}, VarType: cmdconfig.EnvVarTypeString},
	constants.EnvModLocation: {ConfigVar: []string{constants.ArgModLocation}, VarType: cmdconfig.EnvVarTypeString},
}

// NOTE: EnvWorkspaceProfile has already been set as a viper default as we have already loaded workspace profiles
// (EnvInstallDir has already been set at same time but we set it again to make sure it has the correct precedence)

// a map of known environment variables to map to viper keys - these are set as part of LoadGlobalConfig
var envMappings = map[string]cmdconfig.EnvMapping{
	constants.EnvInstallDir:    {ConfigVar: []string{constants.ArgInstallDir}, VarType: cmdconfig.EnvVarTypeString},
	constants.EnvModLocation:   {ConfigVar: []string{constants.ArgModLocation}, VarType: cmdconfig.EnvVarTypeString},
	constants.EnvIntrospection: {ConfigVar: []string{constants.ArgIntrospection}, VarType: cmdconfig.EnvVarTypeString},
	constants.EnvTelemetry:     {ConfigVar: []string{constants.ArgTelemetry}, VarType: cmdconfig.EnvVarTypeString},
	constants.EnvUpdateCheck:   {ConfigVar: []string{constants.ArgUpdateCheck}, VarType: cmdconfig.EnvVarTypeBool},
	// EnvPipesHost needs to be defined before EnvCloudHost,
	// so that if EnvCloudHost is defined, it can override EnvPipesHost
	constants.EnvPipesHost: {ConfigVar: []string{constants.ArgCloudHost}, VarType: cmdconfig.EnvVarTypeString},
	constants.EnvCloudHost: {ConfigVar: []string{constants.ArgCloudHost}, VarType: cmdconfig.EnvVarTypeString},
	// EnvPipesToken needs to be defined before EnvCloudToken,
	// so that if EnvCloudToken is defined, it can override EnvPipesToken
	constants.EnvPipesToken: {ConfigVar: []string{constants.ArgCloudToken}, VarType: cmdconfig.EnvVarTypeString},
	constants.EnvCloudToken: {ConfigVar: []string{constants.ArgCloudToken}, VarType: cmdconfig.EnvVarTypeString},
	//
	constants.EnvSnapshotLocation:      {ConfigVar: []string{constants.ArgSnapshotLocation}, VarType: cmdconfig.EnvVarTypeString},
	constants.EnvWorkspaceDatabase:     {ConfigVar: []string{constants.ArgWorkspaceDatabase}, VarType: cmdconfig.EnvVarTypeString},
	constants.EnvServicePassword:       {ConfigVar: []string{constants.ArgServicePassword}, VarType: cmdconfig.EnvVarTypeString},
	constants.EnvDisplayWidth:          {ConfigVar: []string{constants.ArgDisplayWidth}, VarType: cmdconfig.EnvVarTypeInt},
	constants.EnvMaxParallel:           {ConfigVar: []string{constants.ArgMaxParallel}, VarType: cmdconfig.EnvVarTypeInt},
	constants.EnvQueryTimeout:          {ConfigVar: []string{constants.ArgDatabaseQueryTimeout}, VarType: cmdconfig.EnvVarTypeInt},
	constants.EnvDatabaseStartTimeout:  {ConfigVar: []string{constants.ArgDatabaseStartTimeout}, VarType: cmdconfig.EnvVarTypeInt},
	constants.EnvDashboardStartTimeout: {ConfigVar: []string{constants.ArgDashboardStartTimeout}, VarType: cmdconfig.EnvVarTypeInt},
	constants.EnvCacheTTL:              {ConfigVar: []string{constants.ArgCacheTtl}, VarType: cmdconfig.EnvVarTypeInt},
	constants.EnvCacheMaxTTL:           {ConfigVar: []string{constants.ArgCacheMaxTtl}, VarType: cmdconfig.EnvVarTypeInt},
	constants.EnvMemoryMaxMb:           {ConfigVar: []string{constants.ArgMemoryMaxMb}, VarType: cmdconfig.EnvVarTypeInt},
	constants.EnvMemoryMaxMbPlugin:     {ConfigVar: []string{constants.ArgMemoryMaxMbPlugin}, VarType: cmdconfig.EnvVarTypeInt},

	// we need this value to go into different locations
	constants.EnvCacheEnabled: {ConfigVar: []string{
		constants.ArgClientCacheEnabled,
		constants.ArgServiceCacheEnabled,
	}, VarType: cmdconfig.EnvVarTypeBool},
}
