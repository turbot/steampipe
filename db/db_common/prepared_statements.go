package db_common

import (
	"context"
	"fmt"
	"strings"

	typehelpers "github.com/turbot/go-kit/types"

	"github.com/turbot/steampipe/utils"
)

func CreatePreparedStatements(ctx context.Context, workspace WorkspaceResourceProvider, client Client) error {
	utils.LogTime("db.CreatePreparedStatements start")
	defer utils.LogTime("db.CreatePreparedStatements end")

	var sql []string
	for name, query := range workspace.GetQueryMap() {
		// query map contains long and short names for queries - avoid dupes
		if !strings.HasPrefix(name, "query.") {
			continue
		}
		sql = append(sql, fmt.Sprintf("PREPARE %s AS (\n%s\n)", query.ShortName, typehelpers.SafeString(query.SQL)))
	}
	// execute the query, passing 'true' to disable the spinner
	_, err := client.ExecuteSync(context.Background(), strings.Join(sql, "\n"), true)
	if err != nil {
		return fmt.Errorf("failed to create prepared statements tables: %v", err)
	}

	// return context error - this enables calling code to respond to cancellation
	return ctx.Err()
}
