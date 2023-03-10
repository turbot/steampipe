package db_common

// TODO KAI use ErrorsAndWarnings
type AcquireSessionResult struct {
	Session  *DatabaseSession
	Error    error
	Warnings []string
}
