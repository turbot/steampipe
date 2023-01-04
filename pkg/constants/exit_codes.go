package constants

const (
	ExitCodeSuccessful                   = 0
	ExitCodeControlsAlarm                = 1 // check - no runtime errors, 1 or more control alarms, no control errors
	ExitCodeControlsError                = 2 // check - no runtime errors, 1 or more control errors
	ExitCodeUnknownErrorPanic            = 3 // check - runtime error
	ExitCodeInsufficientOrWrongArguments = 4 // check/plugin - runtime error
	ExitCodeLoadingError                 = 5 // plugin - loading error
	ExitCodePluginListFailure            = 6 // plugin - list failure
	ExitCodeNoModFile                    = 15
	ExitCodeBindPortUnavailable          = 31
)
