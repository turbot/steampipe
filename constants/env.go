package constants

// Environment Variables
const (
	EnvUpdateCheck     = "STEAMPIPE_UPDATE_CHECK"
	EnvInstallDir      = "STEAMPIPE_INSTALL_DIR"
	EnvInstallDatabase = "STEAMPIPE_INITDB_DATABASE_NAME"
	EnvServicePassword = "STEAMPIPE_DATABASE_PASSWORD"
	EnvMaxParallel     = "STEAMPIPE_MAX_PARALLEL"

	EnvWorkspaceDatabase = "STEAMPIPE_WORKSPACE_DATABASE"
	EnvCloudHost         = "STEAMPIPE_CLOUD_HOST"
	EnvCloudToken        = "STEAMPIPE_CLOUD_TOKEN"

	EnvCheckDisplayWidth = "STEAMPIPE_CHECK_DISPLAY_WIDTH"
	EnvCacheEnabled      = "STEAMPIPE_CACHE"
	EnvCacheTTL          = "STEAMPIPE_CACHE_TTL"
	EnvConnectionWatcher = "STEAMPIPE_CONNECTION_WATCHER"
	EnvWorkspaceChDir    = "STEAMPIPE_WORKSPACE_CHDIR"

	// EnvInputVarPrefix is the prefix for environment variables that represent values for input variables.
	EnvInputVarPrefix = "SP_VAR_"
)
