package workspace

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/utils"
)

// EnsureSessionData determines whether session scoped data (introspection tables and prepared statements)
// exists for this session, and if not, creates it
func EnsureSessionData(ctx context.Context, source *SessionDataSource, conn *pgx.Conn) (err error, warnings []string) {
	utils.LogTime("workspace.EnsureSessionData start")
	defer utils.LogTime("workspace.EnsureSessionData end")

	if conn == nil {
		return errors.New("nil conn passed to EnsureSessionData"), nil
	}

	// check for introspection tables
	// if the steampipe_mod table is missing, assume we have no session data - go ahead and create it
	row := conn.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema LIKE 'pg_temp%' AND table_name='steampipe_mod' ")

	var count int
	err = row.Scan(&count)
	if err != nil {
		return err, warnings
	}
	if count == 0 {
		err, warnings = db_common.CreatePreparedStatements(ctx, source.PreparedStatementSource(), conn)
		if err != nil {
			return err, warnings
		}

		err = db_common.CreateIntrospectionTables(ctx, source.IntrospectionTableSource(), conn)
		if err != nil {
			return err, warnings
		}
	}
	return nil, warnings
}
