package db_common

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turbot/steampipe/pkg/query/queryresult"
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

	ExecuteSync(context.Context, string, ...any) (*queryresult.SyncQueryResult, error)
	Execute(context.Context, string, ...any) (*queryresult.Result, error)

	ExecuteSyncInSession(context.Context, *DatabaseSession, string, ...any) (*queryresult.SyncQueryResult, error)
	ExecuteInSession(context.Context, *DatabaseSession, func(), string, ...any) (*queryresult.Result, error)

	RefreshSessions(context.Context) *AcquireSessionResult
	GetSchemaFromDB(context.Context) (*SchemaMetadata, error)

	ServerSettings() *ServerSettings

	// DisablePool will disable the user pool and use a single connection for all query executions
	// This allows us to retain the state of the client when we are in the interactive prompt
	//
	// Note: this does not disable the management pool.
	DisablePool(context.Context) error
}
