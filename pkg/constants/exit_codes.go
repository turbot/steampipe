package constants

const (
	ExitCodeSuccessful                   = 0
	ExitCodeControlsAlarm                = 1 // no runtime errors, 1 or more control alarms, no control errors
	ExitCodeControlsError                = 2 // no runtime errors, 1 or more control errors
	ExitCodeRuntimeError                 = 3 // runtime errors
	ExitCodeUnknownErrorPanic            = 1 // what to do?
	ExitCodeInsufficientOrWrongArguments = 2 // what to do?
	ExitCodeLoadingError                 = 3
	ExitCodePluginListFailure            = 4
	ExitCodeNoModFile                    = 15
	ExitCodeBindPortUnavailable          = 31
)
