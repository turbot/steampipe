package db_common

import (
	"github.com/turbot/pipe-fittings/error_helpers"
)

type AcquireSessionResult struct {
	Session *DatabaseSession
	error_helpers.ErrorAndWarnings
}
