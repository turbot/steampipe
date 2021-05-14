package tabledisplay

import (
	"github.com/logrusorgru/aurora"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/control/execute"
)

// TODO handle light/dark scheme? maybe make this a map

// groups display

// id color
var colorId = constants.BrightWhite

// count
var colorCountZeroFail = constants.Gray1
var colorCountZeroFailDivider = constants.Gray1
var colorCountFail = constants.BoldBrightRed
var colorCountTotal = constants.BrightWhite
var colorCountTotalAllPassed = constants.BoldBrightGreen

func colorCountDivider(arg interface{}) aurora.Value {
	return constants.Bold(constants.Gray2(arg))
}

// count graph
var colorCountGraphFail = constants.BoldBrightRed
var colorCountGraphPass = constants.BrightGreen

// result colors
// state
var statusColors = map[string]func(arg interface{}) aurora.Value{
	execute.ControlAlarm: constants.BoldBrightRed,
	execute.ControlError: constants.BoldBrightRed,
	execute.ControlSkip:  constants.Gray3,
	execute.ControlInfo:  constants.BrightCyan,
	execute.ControlOk:    constants.BrightGreen,
}
var reasonColors = map[string]func(arg interface{}) aurora.Value{
	execute.ControlAlarm: constants.BoldBrightRed,
	execute.ControlError: constants.BoldBrightRed,
	execute.ControlSkip:  constants.Gray3,
	execute.ControlInfo:  constants.BrightCyan,
	execute.ControlOk:    constants.Gray4,
}

// spacer dots used by both group and result
var colorSpacer = constants.Gray1
