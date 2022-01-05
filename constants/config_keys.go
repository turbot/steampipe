package constants

// viper config keys
const (
	// ConfigKeyDatabaseSearchPath is used to store the search path set in the database config in viper
	// the viper value will be set via via a call to getScopedKey in steampipeconfig/steampipeconfig.go
	ConfigKeyDatabaseSearchPath = "database.search-path"
	ConfigKeyInteractive        = "interactive"
	ConfigKeyActiveCommand      = "cmd"
	ConfigKeyActiveCommandArgs  = "cmd_args"
	ConfigInteractiveVariables  = "interactive_var"
	ConfigKeyIsTerminalTTY      = "is_terminal"
)
