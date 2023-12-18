package steampipe_db_common

import (
	"context"
	"database/sql"
)

// ExecContext is an interface exposing an ExecContext method.
type ExecContext interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}
