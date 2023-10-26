package db_common

import (
	"context"
	"database/sql"
)

type ExecContext interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}
