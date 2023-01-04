package constants

const (
	ExitCodeSuccessful                     = 0
	ExitCodeControlsAlarm                  = 1   // check - no runtime errors, 1 or more control alarms, no control errors
	ExitCodeControlsError                  = 2   // check - no runtime errors, 1 or more control errors
	ExitCodeLoadingError                   = 5   // plugin - loading error
	ExitCodePluginListFailure              = 6   // plugin - list failure
	ExitCodeNoModFile                      = 15  // dashboard - no mod file
	ExitCodeBindPortUnavailable            = 31  // service/dashboard - binding failure
	ExitCodeSnapshotCreationFailed         = 51  // snapshot
	ExitCodeSnapshotUploadFailed           = 52  // snapshot
	ExitCodeFileSystemAccessFailure        = 61  // service
	ExitCodePGServiceSetupFailure          = 62  // service
	ExitCodePGServiceStartupFailure        = 63  // service
	ExitCodeDashboardServiceStartupFailure = 64  // service
	ExitCodeSomeQueryExecutionFailed       = 81  // query
	ExitCodeCloudConnectionFailed          = 91  // login
	ExitCodeInsufficientOrWrongInputs      = 252 // check/plugin/mod - runtime error(insufficient or wrong input)
	ExitCodeUnknownErrorPanic              = 253 // all - runtime error(unknown)
)
