package workspace

import (
	"context"

	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
)

// EnsureServiceState queries the database and makes sure that workspace temp tables
// and prepared statements are available in the database
func EnsureServiceState(ctx context.Context, preparedStatementProviders *modconfig.WorkspaceResourceMaps, client db_common.Client) error {
	utils.LogTime("workspace.EnsureServiceState start")
	defer utils.LogTime("workspace.EnsureServiceState end")

	// check for introspection tables
	result, err := client.ExecuteSync(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema LIKE 'pg_temp%' AND table_name='steampipe_mod' ", true)
	if err != nil {
		return err
	}

	// since we are quering with a 'select count...', we will always have exactly one cell with the value
	count := result.Rows[0].(*queryresult.RowResult).Data[0].(int64)
	if count == 0 {
		err = db_common.CreatePreparedStatements(context.Background(), preparedStatementProviders, client)
		if err != nil {
			return err
		}
		if err = db_common.CreateIntrospectionTables(ctx, preparedStatementProviders, client); err != nil {
			return err
		}
	}
	return nil
}
