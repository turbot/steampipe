package constants

const (
	ExitCodeSuccessful                 = 0
	ExitCodeControlsAlarm              = 1   // check - no runtime errors, 1 or more control alarms, no control errors
	ExitCodeControlsError              = 2   // check - no runtime errors, 1 or more control errors
	ExitCodePluginLoadingError         = 11  // plugin - loading error
	ExitCodePluginListFailure          = 12  // plugin - list failure
	ExitCodeSnapshotCreationFailed     = 21  // snapshot
	ExitCodeSnapshotUploadFailed       = 22  // snapshot
	ExitCodeServiceSetupFailure        = 31  // service
	ExitCodeServiceStartupFailure      = 32  // service
	ExitCodeQueryExecutionFailed       = 41  // query - 1 or more queries failed - change in behavior(previously the exitCode used to be the number of queries that failed)
	ExitCodeLoginCloudConnectionFailed = 51  // login
	ExitCodeBindPortUnavailable        = 251 // common(service/dashboard) - port binding failure
	ExitCodeNoModFile                  = 252 // common - no mod file
	ExitCodeFileSystemAccessFailure    = 253 // common - file system access failure
	ExitCodeInsufficientOrWrongInputs  = 254 // common - runtime error(insufficient or wrong input)
	ExitCodeUnknownErrorPanic          = 255 // common - runtime error(unknown panic)
)
