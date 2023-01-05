package constants

const (
	ExitCodeSuccessful                     = 0
	ExitCodeControlsAlarm                  = 1   // check - no runtime errors, 1 or more control alarms, no control errors
	ExitCodeControlsError                  = 2   // check - no runtime errors, 1 or more control errors
	ExitCodeLoadingError                   = 11  // plugin - loading error
	ExitCodePluginListFailure              = 12  // plugin - list failure
	ExitCodeNoModFile                      = 21  // dashboard - no mod file
	ExitCodeSnapshotCreationFailed         = 41  // snapshot
	ExitCodeSnapshotUploadFailed           = 42  // snapshot
	ExitCodeFileSystemAccessFailure        = 61  // service
	ExitCodePGServiceSetupFailure          = 62  // service
	ExitCodePGServiceStartupFailure        = 63  // service
	ExitCodeDashboardServiceStartupFailure = 64  // service
	ExitCodeBindPortUnavailable            = 65  // service/dashboard - binding failure
	ExitCodeSomeQueryExecutionFailed       = 81  // query
	ExitCodeCloudConnectionFailed          = 91  // login
	ExitCodeInsufficientOrWrongInputs      = 252 // check/plugin/mod - runtime error(insufficient or wrong input)
	ExitCodeUnknownErrorPanic              = 253 // all - runtime error(unknown)
)
