package workspace

import (
	"context"
	"database/sql"

	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/utils"
)

// EnsureSessionData determines whether session scoped data (introspection tables and prepared statements)
// exists for this session, and if not, creates it
func EnsureSessionData(ctx context.Context, source *SessionDataSource, session *sql.Conn) error {
	utils.LogTime("workspace.EnsureSessionData start")
	defer utils.LogTime("workspace.EnsureSessionData end")

	// check for introspection tables
	row := session.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema LIKE 'pg_temp%' AND table_name='steampipe_mod' ")

	var count int
	err := row.Scan(&count)
	if err != nil {
		return err
	}

	// if the steampipe_mod table is missing, assume we have no session data - go ahead and create it
	if count == 0 {
		if err = db_common.CreatePreparedStatements(ctx, source.PreparedStatementSource, session); err != nil {
			return err
		}
		if err = db_common.CreateIntrospectionTables(ctx, source.IntrospectionTableSource, session); err != nil {
			return err
		}
	}
	return nil
}
