package db_common

import (
	"github.com/turbot/pipe-fittings/v2/error_helpers"
)

type AcquireSessionResult struct {
	Session *DatabaseSession
	error_helpers.ErrorAndWarnings
}
