package constants

// Argument name constants
const (
	ArgJSON                    = "json"
	ArgCSV                     = "csv"
	ArgTable                   = "table"
	ArgLine                    = "line"
	ArgForce                   = "force"
	ArgTimer                   = "timing"
	ArgOn                      = "on"
	ArgOff                     = "off"
	ArgPortDeprecated          = "db-port"
	ArgPort                    = "database-port"
	ArgListenAddressDeprecated = "listen"
	ArgListenAddress           = "database-listen"
	ArgInvoker                 = "invoker"
	ArgRefresh                 = "refresh"
	ArgUpdateCheck             = "update-check"
	ArgInstallDir              = "install-dir"
	ArgWorkspace               = "workspace"
	ArgSearchPath              = "search-path"
	ArgSearchPathPrefix        = "search-path-prefix"
	ArgWatch                   = "watch"
	ArgTheme                   = "theme"
	ArgProgress                = "progress"
	ArgExport                  = "export"
	ArgDryRun                  = "dry-run"
	ArgWhere                   = "where"
)

/// metaquery mode arguments

var ArgOutput = ArgFromMetaquery(CmdOutput)
var ArgSeparator = ArgFromMetaquery(CmdSeparator)
var ArgHeader = ArgFromMetaquery(CmdHeaders)
var ArgMultiLine = ArgFromMetaquery(CmdMulti)

// BoolToOnOff :: convert a boolean value onto the string "on" or "off"
func BoolToOnOff(val bool) string {
	if val {
		return ArgOn
	}
	return ArgOff
}

// BoolToEnableDisable :: convert a boolean value onto the string "enable" or "disable"
func BoolToEnableDisable(val bool) string {
	if val {
		return "enable"
	}
	return "disable"

}
