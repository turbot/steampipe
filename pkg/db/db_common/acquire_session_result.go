package db_common

type AcquireSessionResult struct {
	Session  *DatabaseSession
	Error    error
	Warnings []string
}
