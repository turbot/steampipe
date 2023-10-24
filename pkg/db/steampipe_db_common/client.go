package steampipe_db_common

import (
	"context"
	"database/sql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/turbot/pipe-fittings/db_common"
	"github.com/turbot/pipe-fittings/queryresult"
)

type Client interface {
	//db_common.Client

	//acquire a management database connection - must be closed
	AcquireManagementConnection(context.Context) (*sql.Conn, error)
	//acquire a query execution session (which search pathand cache options  set) - must be closed
	AcquireSession(context.Context) *AcquireSessionResult

	ExecuteSyncInSession(context.Context, *DatabaseSession, string, ...any) (*queryresult.SyncQueryResult, error)
	ExecuteInSession(context.Context, *DatabaseSession, func(), string, ...any) (*queryresult.Result, error)

	ResetPools(context.Context)
	GetSchemaFromDB(context.Context) (*db_common.SchemaMetadata, error)

	ServerSettings() *ServerSettings
	RegisterNotificationListener(f func(notification *pgconn.Notification))
}
