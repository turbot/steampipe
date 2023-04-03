package db_common

import "github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"

type AcquireSessionResult struct {
	Session *DatabaseSession
	modconfig.ErrorAndWarnings
}
