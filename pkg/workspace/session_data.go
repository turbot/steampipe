package workspace

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/utils"
)

// EnsureSessionData determines whether session scoped data (introspection tables and prepared statements)
// exists for this session, and if not, creates it
func EnsureSessionData(ctx context.Context, source *SessionDataSource, conn *pgx.Conn) (error) {
	utils.LogTime("workspace.EnsureSessionData start")
	defer utils.LogTime("workspace.EnsureSessionData end")

	if conn == nil {
		return errors.New("nil conn passed to EnsureSessionData")
	}

	// check for introspection tables
	// if the steampipe_mod table is missing, assume we have no session data - go ahead and create it
	row := conn.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema LIKE 'pg_temp%' AND table_name='steampipe_mod' ")

	var count int
	err := row.Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {

		err = db_common.CreateIntrospectionTables(ctx, source.IntrospectionTableSource(), conn)
		if err != nil {
			return err
		}
	}
	return nil
}
