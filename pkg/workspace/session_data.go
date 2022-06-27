package workspace

import (
	"context"
	"errors"

	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/utils"
)

// EnsureSessionData determines whether session scoped data (introspection tables and prepared statements)
// exists for this session, and if not, creates it
func EnsureSessionData(ctx context.Context, source *SessionDataSource, session *db_common.DatabaseSession) (err error, warnings []string) {
	utils.LogTime("workspace.EnsureSessionData start")
	defer utils.LogTime("workspace.EnsureSessionData end")

	if session == nil {
		return errors.New("nil session passed to EnsureSessionData"), nil
	}

	// check for introspection tables
	row := session.Connection.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema LIKE 'pg_temp%' AND table_name='steampipe_mod' ")

	var count int
	err = row.Scan(&count)
	if err != nil {
		return err, warnings
	}

	// if the steampipe_mod table is missing, assume we have no session data - go ahead and create it
	if count == 0 {
		session.LifeCycle.Add("prepared_statement_start")
		err, warnings = db_common.CreatePreparedStatements(ctx, source.PreparedStatementSource(), session)
		if err != nil {
			return err, warnings
		}
		session.LifeCycle.Add("prepared_statement_finish")

		session.LifeCycle.Add("introspection_table_start")
		err = db_common.CreateIntrospectionTables(ctx, source.IntrospectionTableSource(), session)
		session.LifeCycle.Add("introspection_table_finish")
		if err != nil {
			return err, warnings
		}

	}
	return nil, warnings
}
