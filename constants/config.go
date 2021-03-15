package constants

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/turbot/steampipe/utils"
)

// Constants for Config
const (
	DefaultInstallDir        = "~/.steampipe"
	PluginExtension          = ".plugin"
	ConfigExtension          = ".spc"
	ConnectionsStateFileName = "connection.json"
)

var SteampipeDir string

func steampipeSubDir(dirName string) string {
	subDir := filepath.Join(SteampipeDir, dirName)

	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		err = os.MkdirAll(subDir, 0755)
		utils.FailOnErrorWithMessage(err, fmt.Sprintf("could not create %s directory", dirName))
	}

	return subDir
}

// PluginDir :: returns the path to the plugins directory (creates if missing)
func PluginDir() string {
	return steampipeSubDir("plugins")
}

// ConnectionStatePath :: returns the path of the connections state file
func ConnectionStatePath() string {
	return filepath.Join(InternalDir(), ConnectionsStateFileName)
}

// ModsDir :: returns the path to the mods directory (creates if missing)
func ModsDir() string {
	return steampipeSubDir("mods")
}

// ConfigDir :: returns the path to the config directory (creates if missing)
func ConfigDir() string {
	return steampipeSubDir("config")
}

// InternalDir :: returns the path to the internal directory (creates if missing)
func InternalDir() string {
	return steampipeSubDir("internal")
}

// DatabaseDir :: returns the path to the db directory (creates if missing)
func DatabaseDir() string {
	return steampipeSubDir("db")
}

// LogDir :: returns the path to the db log directory (creates if missing)
func LogDir() string {
	return steampipeSubDir("logs")
}

// TempDir :: returns the path to the steampipe tmp directory (creates if missing)
func TempDir() string {
	return steampipeSubDir("tmp")
}
