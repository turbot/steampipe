package db_common

import (
	"context"
	"database/sql"
)

// get the unique postgres identifier for a database session
func GetBackendPid(ctx context.Context, session *sql.Conn) (int64, error) {
	var pid int64
	rows, err := session.QueryContext(ctx, "select pg_backend_pid()")
	if err != nil {
		return pid, err
	}
	rows.Next()
	rows.Scan(&pid)
	rows.Close()
	return pid, nil
}
