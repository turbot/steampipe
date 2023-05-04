package steampipeconfig

type WaitModeValue int

const (
	NoWait WaitModeValue = iota
	WaitForLoading
	WaitForReady
	WaitForSearchPath
)

type LoadConnectionStateConfiguration struct {
	WaitMode    WaitModeValue
	Connections []string
	SearchPath  []string
}

type LoadConnectionStateOption = func(config *LoadConnectionStateConfiguration)

// WithWaitUntilLoading waits until no connections are in pending state
var WithWaitUntilLoading = func() func(config *LoadConnectionStateConfiguration) {
	return func(config *LoadConnectionStateConfiguration) {
		config.WaitMode = WaitForLoading
	}
}

var WithWaitForSearchPath = func(searchPath []string) func(config *LoadConnectionStateConfiguration) {
	return func(config *LoadConnectionStateConfiguration) {
		config.WaitMode = WaitForSearchPath
		config.SearchPath = searchPath
	}
}

// WithWaitUntilReady waits until all are in ready state
var WithWaitUntilReady = func(connections ...string) func(config *LoadConnectionStateConfiguration) {
	return func(config *LoadConnectionStateConfiguration) {
		config.Connections = connections
		config.WaitMode = WaitForReady
	}
}
