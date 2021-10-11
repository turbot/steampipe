package constants

// Known Environment Variables
const (
	EnvUpdateCheck      = "STEAMPIPE_UPDATE_CHECK"
	EnvInstallDir       = "STEAMPIPE_INSTALL_DIR"
	EnvConnectionString = "STEAMPIPE_CONNECTION_STRING"
	EnvInstallDatabase  = "STEAMPIPE_INITDB_DATABASE_NAME"
	EnvDatabase         = "STEAMPIPE_DATABASE"
	EnvAPIKey           = "STEAMPIPE_API_KEY"
	EnvServicePassword  = "STEAMPIPE_DATABASE_PASSWORD"
	// EnvInputVarPrefix is the prefix for environment variables that represent values for input variables.
	EnvInputVarPrefix = "SP_VAR_"
)
