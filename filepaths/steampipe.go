package filepaths

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/turbot/steampipe/utils"
)

// Constants for Config
const (
	DefaultInstallDir           = "~/.steampipe"
	connectionsStateFileName    = "connection.json"
	versionFileName             = "versions.json"
	databaseRunningInfoFileName = "steampipe.json"
	pluginManagerStateFileName  = "plugin_manager.json"
)

var SteampipeDir string

func steampipeSubDir(dirName string) string {
	if SteampipeDir == "" {
		panic(fmt.Errorf("cannot call any Steampipe directory functions before SteampipeDir is set"))
	}
	subDir := filepath.Join(SteampipeDir, dirName)

	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		err = os.MkdirAll(subDir, 0755)
		utils.FailOnErrorWithMessage(err, fmt.Sprintf("could not create %s directory", dirName))
	}

	return subDir
}

// TemplateDir returns the path to the templates directory (creates if missing)
func TemplateDir() string {
	return steampipeSubDir(filepath.Join("check", "templates"))
}

// PluginDir returns the path to the plugins directory (creates if missing)
func PluginDir() string {
	return steampipeSubDir("plugins")
}

// ConnectionStatePath returns the path of the connections state file
func ConnectionStatePath() string {
	return filepath.Join(InternalDir(), connectionsStateFileName)
}

// ConfigDir returns the path to the config directory (creates if missing)
func ConfigDir() string {
	return steampipeSubDir("config")
}

// InternalDir returns the path to the internal directory (creates if missing)
func InternalDir() string {
	return steampipeSubDir("internal")
}

// DatabaseDir returns the path to the db directory (creates if missing)
func DatabaseDir() string {
	return steampipeSubDir("db")
}

// LogDir returns the path to the db log directory (creates if missing)
func LogDir() string {
	return steampipeSubDir("logs")
}

// LegacyVersionFilePath returns the legacy version file path
func LegacyVersionFilePath() string {
	return filepath.Join(InternalDir(), versionFileName)
}

// PluginVersionFilePath returns the plugin version file path
func PluginVersionFilePath() string {
	return filepath.Join(PluginDir(), versionFileName)
}

// DatabaseVersionFilePath returns the plugin version file path
func DatabaseVersionFilePath() string {
	return filepath.Join(DatabaseDir(), versionFileName)
}

func ReportAssetsPath() string {
	return steampipeSubDir(filepath.Join(filepath.Join("report", "assets")))
}

// ReportAssetsVersionFilePath returns the report assets version file path
func ReportAssetsVersionFilePath() string {
	return filepath.Join(ReportAssetsPath(), versionFileName)
}

func RunningInfoFilePath() string {
	return filepath.Join(InternalDir(), databaseRunningInfoFileName)
}

func PluginManagerStateFilePath() string {
	return filepath.Join(InternalDir(), pluginManagerStateFileName)
}
