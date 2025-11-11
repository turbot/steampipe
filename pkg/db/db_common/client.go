package db_common

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	pqueryresult "github.com/turbot/pipe-fittings/v2/queryresult"
	"github.com/turbot/steampipe/v2/pkg/query/queryresult"
)

type Client interface {
	Close(context.Context) error
	LoadUserSearchPath(context.Context) error

	SetRequiredSessionSearchPath(context.Context) error
	GetRequiredSessionSearchPath() []string
	GetCustomSearchPath() []string

	// acquire a management database connection - must be closed
	AcquireManagementConnection(context.Context) (*pgxpool.Conn, error)
	// acquire a query execution session (which search pathand cache options  set) - must be closed
	AcquireSession(context.Context) *AcquireSessionResult

	ExecuteSync(context.Context, string, ...any) (*pqueryresult.SyncQueryResult, error)
	Execute(context.Context, string, ...any) (*queryresult.Result, error)

	ExecuteSyncInSession(context.Context, *DatabaseSession, string, ...any) (*pqueryresult.SyncQueryResult, error)
	ExecuteInSession(context.Context, *DatabaseSession, func(), string, ...any) (*queryresult.Result, error)

	ResetPools(context.Context)
	GetSchemaFromDB(context.Context) (*SchemaMetadata, error)

	ServerSettings() *ServerSettings
	RegisterNotificationListener(f func(notification *pgconn.Notification))
}
