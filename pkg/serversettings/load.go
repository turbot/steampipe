package serversettings

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/sperr"
)

func Load(ctx context.Context, conn *pgx.Conn) (_ *db_common.ServerSettings, e error) {
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

	settings := new(db_common.ServerSettings)
	pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[db_common.ServerSettings])
	return settings, nil
}
