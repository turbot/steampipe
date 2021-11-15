package workspace

import (
	"context"
	"errors"
	"time"

	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/utils"
)

// EnsureSessionData determines whether session scoped data (introspection tables and prepared statements)
// exists for this session, and if not, creates it
func EnsureSessionData(ctx context.Context, source *SessionDataSource, session *db_common.DBSession) error {
	utils.LogTime("workspace.EnsureSessionData start")
	defer utils.LogTime("workspace.EnsureSessionData end")

	if session == nil {
		return errors.New("session cannot be nil")
	}

	// check for introspection tables
	row := session.GetRaw().QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema LIKE 'pg_temp%' AND table_name='steampipe_mod' ")

	var count int
	err := row.Scan(&count)
	if err != nil {
		return err
	}

	// if the steampipe_mod table is missing, assume we have no session data - go ahead and create it
	if count == 0 {
		session.Timeline.PreparedStatementStart = time.Now()
		err = db_common.CreatePreparedStatements(ctx, source.PreparedStatementSource, session)
		session.Timeline.PreparedStatementFinish = time.Now()
		if err != nil {
			return err
		}
		session.Timeline.IntrospectionTableStart = time.Now()
		err = db_common.CreateIntrospectionTables(ctx, source.IntrospectionTableSource, session)
		session.Timeline.IntrospectionTableFinish = time.Now()
		if err != nil {
			return err
		}
	}
	return nil
}
