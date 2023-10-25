package serversettings

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/db/steampipe_db_common"
)

func Load(ctx context.Context, pool *sql.DB) (serverSettings *steampipe_db_common.ServerSettings, e error) {
	conn, err := pool.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	defer func() {
		// this function uses reflection to extract and convert values
		// we need to be able to recover from panics while using reflection
		if r := recover(); r != nil {
			e = sperr.ToError(r, sperr.WithMessage("error loading server settings"))
		}
	}()
	rows, err := conn.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %s.%s", constants.InternalSchema, constants.ServerSettingsTable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	serverSettings, e = steampipe_db_common.CollectOneToStructByName[steampipe_db_common.ServerSettings](rows)
	return
}
