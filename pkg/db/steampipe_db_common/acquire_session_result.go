package steampipe_db_common

import (
	"github.com/turbot/steampipe/pkg/error_helpers"
)

type AcquireSessionResult struct {
	Session *DatabaseSession
	error_helpers.ErrorAndWarnings
}
