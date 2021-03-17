package constants

// EnvironmentKeyToViperKey :: a map of environment variables to Viper Config Key
var EnvironmentKeyToViperKey = map[string]string{
	"STEAMPIPE_UPDATE_CHECK": ArgUpdateCheck,
}
