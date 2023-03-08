package workspace

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
)

// EnsureSessionData determines whether introspection tables
// exists for this session, and if not, creates them if needed
func EnsureSessionData(ctx context.Context, source *modconfig.ResourceMaps, conn *pgx.Conn) error {
	utils.LogTime("workspace.EnsureSessionData start")
	defer utils.LogTime("workspace.EnsureSessionData end")

	if conn == nil {
		return errors.New("nil conn passed to EnsureSessionData")
	}

	// check for introspection tables
	// if the steampipe_mod table is missing, assume we have no session data - go ahead and create it
	row := conn.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema LIKE 'pg_temp%' AND table_name='steampipe_mod' ")
	var count int
	if err := row.Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if err := db_common.CreateIntrospectionTables(ctx, source, conn); err != nil {
			return err
		}
	}

	return nil
}
