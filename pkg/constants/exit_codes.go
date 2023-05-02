package constants

const (
	ExitCodeSuccessful                  = 0
	ExitCodeControlsAlarm               = 1   // check - no runtime errors, 1 or more control alarms, no control errors
	ExitCodeControlsError               = 2   // check - no runtime errors, 1 or more control errors
	ExitCodePluginLoadingError          = 11  // plugin - loading error
	ExitCodePluginListFailure           = 12  // plugin - listing failed
	ExitCodePluginNotFound              = 13  // plugin - not found
	ExitCodePluginInstallFailure        = 14  // plugin - install failed
	ExitCodeSnapshotCreationFailed      = 21  // snapshot - creation failed
	ExitCodeSnapshotUploadFailed        = 22  // snapshot - upload failed
	ExitCodeServiceSetupFailure         = 31  // service - setup failed
	ExitCodeServiceStartupFailure       = 32  // service - start failed
	ExitCodeServiceStopFailure          = 33  // service - stop failed
	ExitCodeQueryExecutionFailed        = 41  // query - 1 or more queries failed - change in behavior(previously the exitCode used to be the number of queries that failed)
	ExitCodeLoginCloudConnectionFailed  = 51  // login - connecting to cloud failed
	ExitCodeModInitFailed               = 61  // mod - init failed
	ExitCodeModInstallFailed            = 62  // mod - install failed
	ExitCodeInvalidExecutionEnvironment = 249 // common - when steampipe is run in an unsupported environment
	ExitCodeInitializationFailed        = 250 // common - initialization failed
	ExitCodeBindPortUnavailable         = 251 // common(service/dashboard) - port binding failed
	ExitCodeNoModFile                   = 252 // common - no mod file
	ExitCodeFileSystemAccessFailure     = 253 // common - file system access failed
	ExitCodeInsufficientOrWrongInputs   = 254 // common - runtime error(insufficient or wrong input)
	ExitCodeUnknownErrorPanic           = 255 // common - runtime error(unknown panic)
)
