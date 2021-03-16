package task

import (
	"os"
	"strings"
)

func shouldDoUpdateCheck() bool {
	// TODO USE VIPER / OPTIONS
	// if legacy env var SP_DISABLE_UPDATE_CHECK is true, do nothing
	if v, ok := os.LookupEnv(legacyDisableUpdatesCheckEnvVar); ok && strings.ToLower(v) == "true" {
		return false
	}
	// if STEAMPIPE_UPDATE_CHECK is false, do nothing
	if v, ok := os.LookupEnv(updatesCheckEnvVar); ok && strings.ToLower(v) == "false" {
		return false
	}
	return true
}
