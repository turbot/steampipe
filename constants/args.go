package constants

// Argument name constants
const (
	ArgJSON              = "json"
	ArgCSV               = "csv"
	ArgTable             = "table"
	ArgLine              = "line"
	ArgListAllTableNames = "L"
	ArgSelectAll         = "A"
	ArgForce             = "force"
	ArgTimer             = "timing"
	ArgOn                = "on"
	ArgOff               = "off"
	ArgInstallDir        = "install-dir"
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
