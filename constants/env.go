package constants

// Known Environment Variables
const (
	EnvUpdateCheck       = "STEAMPIPE_UPDATE_CHECK"
	EnvInstallDir        = "STEAMPIPE_INSTALL_DIR"
	EnvInstallDatabase   = "STEAMPIPE_INITDB_DATABASE_NAME"
	EnvDatabaseBackend   = "STEAMPIPE_DATABASE_BACKEND"
	EnvCloudToken        = "STEAMPIPE_CLOUD_TOKEN"
	EnvServicePassword   = "STEAMPIPE_DATABASE_PASSWORD"
	EnvCheckDisplayWidth = "STEAMPIPE_CHECK_DISPLAY_WIDTH"
	// EnvInputVarPrefix is the prefix for environment variables that represent values for input variables.
	EnvInputVarPrefix = "SP_VAR_"
)
