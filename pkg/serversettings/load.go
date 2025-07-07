package serversettings

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
)

func Load(ctx context.Context, pool *pgxpool.Pool) (serverSettings *db_common.ServerSettings, e error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	defer func() {
		// this function uses reflection to extract and convert values
		// we need to be able to recover from panics while using reflection
		if r := recover(); r != nil {
			e = sperr.ToError(r, sperr.WithMessage("error loading server settings"))
		}
	}()
	rows, err := conn.Query(ctx, fmt.Sprintf("SELECT * FROM %s.%s", constants.InternalSchema, constants.ServerSettingsTable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	serverSettings, e = pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[db_common.ServerSettings])
	return
}
