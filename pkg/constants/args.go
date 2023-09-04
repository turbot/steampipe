package constants

// Argument name constants
const (
	ArgHelp                    = "help"
	ArgVersion                 = "version"
	ArgForce                   = "force"
	ArgAll                     = "all"
	ArgTiming                  = "timing"
	ArgOn                      = "on"
	ArgOff                     = "off"
	ArgClear                   = "clear"
	ArgDatabaseListenAddresses = "database-listen"
	ArgDatabasePort            = "database-port"
	ArgDatabaseQueryTimeout    = "query-timeout"
	ArgServicePassword         = "database-password"
	ArgServiceShowPassword     = "show-password"
	ArgDashboard               = "dashboard"
	ArgDashboardListen         = "dashboard-listen"
	ArgDashboardPort           = "dashboard-port"
	ArgDashboardStartTimeout   = "dashboard-start-timeout"
	ArgSkipConfig              = "skip-config"
	ArgForeground              = "foreground"
	ArgInvoker                 = "invoker"
	ArgUpdateCheck             = "update-check"
	ArgTelemetry               = "telemetry"
	ArgInstallDir              = "install-dir"
	ArgWorkspaceDatabase       = "workspace-database"
	ArgSchemaComments          = "schema-comments"
	ArgCloudHost               = "cloud-host"
	ArgCloudToken              = "cloud-token"
	ArgSearchPath              = "search-path"
	ArgSearchPathPrefix        = "search-path-prefix"
	ArgWatch                   = "watch"
	ArgTheme                   = "theme"
	ArgProgress                = "progress"
	ArgExport                  = "export"
	ArgMaxParallel             = "max-parallel"
	ArgLogLevel                = "log-level"
	ArgDryRun                  = "dry-run"
	ArgWhere                   = "where"
	ArgTag                     = "tag"
	ArgVariable                = "var"
	ArgVarFile                 = "var-file"
	ArgConnectionString        = "connection-string"
	ArgDisplayWidth            = "display-width"
	ArgPrune                   = "prune"
	ArgModInstall              = "mod-install"
	ArgServiceMode             = "service-mode"
	ArgBrowser                 = "browser"
	ArgInput                   = "input"
	ArgDashboardInput          = "dashboard-input"
	ArgMaxCacheSizeMb          = "max-cache-size-mb"
	ArgCacheTtl                = "cache-ttl"
	ArgClientCacheEnabled      = "client-cache-enabled"
	ArgServiceCacheEnabled     = "service-cache-enabled"
	ArgCacheMaxTtl             = "cache-max-ttl"
	ArgIntrospection           = "introspection"
	ArgShare                   = "share"
	ArgSnapshot                = "snapshot"
	ArgSnapshotTag             = "snapshot-tag"
	ArgWorkspaceProfile        = "workspace"
	ArgModLocation             = "mod-location"
	ArgSnapshotLocation        = "snapshot-location"
	ArgSnapshotTitle           = "snapshot-title"
	ArgDatabaseStartTimeout    = "database-start-timeout"
	ArgMemoryMaxMb             = "memory-max-mb"
	ArgMemoryMaxMbPlugin       = "memory-max-mb-plugin"
)

// metaquery mode arguments

var ArgOutput = ArgFromMetaquery(CmdOutput)
var ArgSeparator = ArgFromMetaquery(CmdSeparator)
var ArgHeader = ArgFromMetaquery(CmdHeaders)
var ArgMultiLine = ArgFromMetaquery(CmdMulti)
var ArgAutoComplete = ArgFromMetaquery(CmdAutoComplete)

// BoolToOnOff converts a boolean value onto the string "on" or "off"
func BoolToOnOff(val bool) string {
	if val {
		return ArgOn
	}
	return ArgOff
}

// BoolToEnableDisable converts a boolean value onto the string "enable" or "disable"
func BoolToEnableDisable(val bool) string {
	if val {
		return "enable"
	}
	return "disable"

}
