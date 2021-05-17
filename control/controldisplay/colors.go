package controldisplay

import (
	"github.com/logrusorgru/aurora"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/control/execute"
)

// TODO handle light/dark scheme? maybe make this a map

// groups display

// group title color
var colorGroupTitle = constants.BoldBrightWhite

// count
var colorCountZeroFail = constants.Gray1
var colorCountZeroFailDivider = constants.Gray1
var colorCountDivider = constants.Gray2
var colorCountFail = constants.BoldBrightRed
var colorCountTotal = constants.BrightWhite
var colorCountTotalAllPassed = constants.BoldBrightGreen

// count graph
var colorCountGraphFail = constants.BoldBrightRed
var colorCountGraphPass = constants.BrightGreen
var colorCountGraphBracket = constants.Gray2

// result colors
// status
var statusColors = map[string]func(arg interface{}) aurora.Value{
	execute.ControlAlarm: constants.BoldBrightRed,
	execute.ControlError: constants.BoldBrightRed,
	execute.ControlSkip:  constants.Gray3,
	execute.ControlInfo:  constants.BrightCyan,
	execute.ControlOk:    constants.BrightGreen,
}
var colorStatusColon = constants.Gray1

// reason
var reasonColors = map[string]func(arg interface{}) aurora.Value{
	execute.ControlAlarm: constants.BoldBrightRed,
	execute.ControlError: constants.BoldBrightRed,
	execute.ControlSkip:  constants.Gray3,
	execute.ControlInfo:  constants.BrightCyan,
	execute.ControlOk:    constants.Gray4,
}

// spacer dots used by both group and result
var colorSpacer = constants.Gray1
