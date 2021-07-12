package utils

// InteractiveExitStatus :: exist status from the interative prompt
//
// We exit go-prompt after every query (in order to manage the prompt history to only include complete queries)
// We therefore need to distinguish between:
//
// a user requested exit (ctrl+D or .exit) - indicated by a non-nil exit code and restart=false and a value ,
// go-prompt being terminated after a query completion and requiring restarting - indicated by restart=true
type InteractiveExitStatus struct {
	// TODO remove altogether
	ExitCode int
}
