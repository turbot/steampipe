package constants

const (
	// ControlQueryCancellationTimeoutSecs is maximum number of seconds to wait for control queries to finish cancelling
	ControlQueryCancellationTimeoutSecs = 30
	// MaxControlRunAttempts determines how many time should a cotnrol run should be retried
	// in the case of a GRPC connectivity error
	MaxControlRunAttempts = 2
)
