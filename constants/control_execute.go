package constants

// ParallelControlMultiplier is used to determine the nbumber of goroutines to start for the control run
// this is a multiplier for the max db connections which are configred
const ParallelControlMultiplier = 1

// The maximum number of seconds to wait for control queries to finish cancelling
const QueryCancellationTimeout = 30
