package workspace

import (
	"context"

	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/utils"
)

// EnsureSessionData determines whether session scoped data (introspection tables and prepared statements)
// exists for this session, and if not, creates it
func EnsureSessionData(ctx context.Context, source *SessionDataSource, client db_common.Client) error {
	utils.LogTime("workspace.EnsureSessionData start")
	defer utils.LogTime("workspace.EnsureSessionData end")

	// check for introspection tables
	result, err := client.ExecuteSync(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema LIKE 'pg_temp%' AND table_name='steampipe_mod' ", true)
	if err != nil {
		return err
	}
	// since we are quering with a 'select count...', we will always have exactly one cell with the value
	count := result.Rows[0].(*queryresult.RowResult).Data[0].(int64)

	// if the steampipe_mod table is missing, assume we have no session data - go ahead and create it
	if count == 0 {
		err = db_common.CreatePreparedStatements(context.Background(), source.preparedStatementSource, client)
		if err != nil {
			return err
		}
		if err = db_common.CreateMetadataTables(ctx, source.introspectionTableSource, client); err != nil {
			return err
		}
	}
	return nil
}
