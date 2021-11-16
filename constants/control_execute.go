package constants

// the number of goroutines to start.
// this is a multiplier to the max parallel input
// essentially, twice as many go routines wait for DBSession
const ParallelControlMultiplier = 3
