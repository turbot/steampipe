package filepaths

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/turbot/steampipe/utils"
)

// Constants for Config
const (
	DefaultInstallDir            = "~/.steampipe"
	connectionsStateFileName     = "connection.json"
	versionFileName              = "versions.json"
	databaseRunningInfoFileName  = "steampipe.json"
	pluginManagerStateFileName   = "plugin_manager.json"
	dashboardServerStateFileName = "dashboard_service.json"
)

var SteampipeDir string

func ensureSteampipeSubDir(dirName string) string {
	subDir := steampipeSubDir(dirName)

	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		err = os.MkdirAll(subDir, 0755)
		utils.FailOnErrorWithMessage(err, fmt.Sprintf("could not create %s directory", dirName))
	}
	log.Printf("[TRACE] subDir %s\n", subDir)

	return subDir
}

func steampipeSubDir(dirName string) string {
	if SteampipeDir == "" {
		panic(fmt.Errorf("cannot call any Steampipe directory functions before SteampipeDir is set"))
	}
	log.Printf("[TRACE] SteampipeDir: %s dirName: %s in steampipeSubDir\n", SteampipeDir, dirName)
	log.Printf("[TRACE] %s", debug.Stack())
	return filepath.Join(SteampipeDir, dirName)
}

// EnsureTemplateDir returns the path to the templates directory (creates if missing)
func EnsureTemplateDir() string {
	return ensureSteampipeSubDir(filepath.Join("check", "templates"))
}

// EnsurePluginDir returns the path to the plugins directory (creates if missing)
func EnsurePluginDir() string {
	return ensureSteampipeSubDir("plugins")
}

// EnsureConfigDir returns the path to the config directory (creates if missing)
func EnsureConfigDir() string {
	return ensureSteampipeSubDir("config")
}

// EnsureInternalDir returns the path to the internal directory (creates if missing)
func EnsureInternalDir() string {
	return ensureSteampipeSubDir("internal")
}

// EnsureDatabaseDir returns the path to the db directory (creates if missing)
func EnsureDatabaseDir() string {
	return ensureSteampipeSubDir("db")
}

// EnsureLogDir returns the path to the db log directory (creates if missing)
func EnsureLogDir() string {
	return ensureSteampipeSubDir("logs")
}

func EnsureDashboardAssetsDir() string {
	return ensureSteampipeSubDir(filepath.Join("report", "assets"))
}

// ConnectionStatePath returns the path of the connections state file
func ConnectionStatePath() string {
	return filepath.Join(EnsureInternalDir(), connectionsStateFileName)
}

// LegacyVersionFilePath returns the legacy version file path
func LegacyVersionFilePath() string {
	return filepath.Join(EnsureInternalDir(), versionFileName)
}

// PluginVersionFilePath returns the plugin version file path
func PluginVersionFilePath() string {
	return filepath.Join(EnsurePluginDir(), versionFileName)
}

// DatabaseVersionFilePath returns the plugin version file path
func DatabaseVersionFilePath() string {
	return filepath.Join(EnsureDatabaseDir(), versionFileName)
}

// ReportAssetsVersionFilePath returns the report assets version file path
func ReportAssetsVersionFilePath() string {
	return filepath.Join(EnsureDashboardAssetsDir(), versionFileName)
}

func RunningInfoFilePath() string {
	return filepath.Join(EnsureInternalDir(), databaseRunningInfoFileName)
}

func PluginManagerStateFilePath() string {
	return filepath.Join(EnsureInternalDir(), pluginManagerStateFileName)
}

func DashboardServiceStateFilePath() string {
	return filepath.Join(EnsureInternalDir(), dashboardServerStateFileName)
}
