package db_common

import (
	"context"
	"database/sql"

	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/steampipeconfig"
)

type EnsureSessionStateCallback = func(context.Context, *DatabaseSession) (err error, warnings []string)

type Client interface {
	Close(ctx context.Context) error

	ForeignSchemas() []string
	ConnectionMap() *steampipeconfig.ConnectionDataMap

	GetCurrentSearchPath(context.Context) ([]string, error)
	GetCurrentSearchPathForDbConnection(context.Context, *sql.Conn) ([]string, error)
	SetRequiredSessionSearchPath(ctx context.Context) error
	ContructSearchPath(ctx context.Context, requiredSearchPath []string, searchPathPrefix []string) ([]string, error)

	AcquireSession(context.Context) *AcquireSessionResult

	ExecuteSync(ctx context.Context, query string) (*queryresult.SyncQueryResult, error)
	Execute(ctx context.Context, query string) (res *queryresult.Result, err error)

	ExecuteSyncInSession(ctx context.Context, session *DatabaseSession, query string) (*queryresult.SyncQueryResult, error)
	ExecuteInSession(ctx context.Context, session *DatabaseSession, query string, onComplete func()) (res *queryresult.Result, err error)

	CacheOn(context.Context) error
	CacheOff(context.Context) error
	CacheClear(context.Context) error

	SetEnsureSessionDataFunc(EnsureSessionStateCallback)
	RefreshSessions(ctx context.Context) *AcquireSessionResult
	GetSchemaFromDB(context.Context) (*schema.Metadata, error)
	// remote client will have empty implementation
	RefreshConnectionAndSearchPaths(context.Context) *steampipeconfig.RefreshConnectionResult
	LoadForeignSchemaNames(context.Context) error
}
