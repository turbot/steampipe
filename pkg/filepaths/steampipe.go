package filepaths

import (
	"fmt"
	"os"
	"path/filepath"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
)

// Constants for Config
const (
	connectionsStateFileName     = "connection.json"
	versionFileName              = "versions.json"
	databaseRunningInfoFileName  = "steampipe.json"
	pluginManagerStateFileName   = "plugin_manager.json"
	dashboardServerStateFileName = "dashboard_service.json"
	stateFileName                = "update_check.json"
	legacyStateFileName          = "update-check.json"
	availableVersionsFileName    = "available_versions.json"
	legacyNotificationsFileName  = "notifications.json"
	localPluginFolder            = "local"
)

func ensureSteampipeSubDir(dirName string) string {
	subDir := steampipeSubDir(dirName)

	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		err = os.MkdirAll(subDir, 0755)
		error_helpers.FailOnErrorWithMessage(err, fmt.Sprintf("could not create %s directory", dirName))
	}

	return subDir
}

func steampipeSubDir(dirName string) string {
	if app_specific.InstallDir == "" {
		panic(fmt.Errorf("cannot call any Steampipe directory functions before app_specific.InstallDir is set"))
	}
	return filepath.Join(app_specific.InstallDir, dirName)
}

// EnsureTemplateDir returns the path to the templates directory (creates if missing)
func EnsureTemplateDir() string {
	return ensureSteampipeSubDir(filepath.Join("check", "templates"))
}

// EnsureInternalDir returns the path to the internal directory (creates if missing)
func EnsureInternalDir() string {
	return ensureSteampipeSubDir("internal")
}

// EnsureBackupsDir returns the path to the backups directory (creates if missing)
func EnsureBackupsDir() string {
	return ensureSteampipeSubDir("backups")
}

// BackupsDir returns the path to the backups directory
func BackupsDir() string {
	return steampipeSubDir("backups")
}

// WorkspaceProfileDir returns the path to the workspace profiles directory
// if  STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set use that
// otherwise look in the config folder
// NOTE: unlike other path functions this accepts the install-dir as arg
// this is because of the slightly complex bootstrapping process required because the
// install-dir may be set in the workspace profile
func WorkspaceProfileDir(installDir string) (string, error) {
	if workspaceProfileLocation, ok := os.LookupEnv(constants.EnvWorkspaceProfileLocation); ok {
		return filehelpers.Tildefy(workspaceProfileLocation)
	}
	return filepath.Join(installDir, "config"), nil

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
	return ensureSteampipeSubDir(filepath.Join("dashboard", "assets"))
}

// LegacyDashboardAssetsDir returns the path to the legacy report assets folder
func LegacyDashboardAssetsDir() string {
	return steampipeSubDir("report")
}

// LegacyStateFilePath returns the path of the legacy update-check.json state file
func LegacyStateFilePath() string {
	return filepath.Join(EnsureInternalDir(), legacyStateFileName)
}

// StateFilePath returns the path of the update_check.json state file
func StateFilePath() string {
	return filepath.Join(EnsureInternalDir(), stateFileName)
}

// AvailableVersionsFilePath returns the path of the json file used to store cache available versions of installed plugins and the CLI
func AvailableVersionsFilePath() string {
	return filepath.Join(EnsureInternalDir(), availableVersionsFileName)
}

// LegacyNotificationsFilePath returns the path of the (legacy) notifications.json file used to store update notifications
func LegacyNotificationsFilePath() string {
	return filepath.Join(EnsureInternalDir(), legacyNotificationsFileName)
}

// ConnectionStatePath returns the path of the connections state file
func ConnectionStatePath() string {
	return filepath.Join(EnsureInternalDir(), connectionsStateFileName)
}

// LegacyVersionFilePath returns the legacy version file path
func LegacyVersionFilePath() string {
	return filepath.Join(EnsureInternalDir(), versionFileName)
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

func StateFileName() string {
	return stateFileName
}
