package db_common

import (
	"context"
	"database/sql"

	"github.com/turbot/steampipe/pkg/query/queryresult"
	"github.com/turbot/steampipe/pkg/schema"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
)

type Client interface {
	Close(ctx context.Context) error

	ForeignSchemaNames() []string
	LoadForeignSchemaNames(ctx context.Context) error
	ConnectionMap() *steampipeconfig.ConnectionDataMap

	GetCurrentSearchPath(context.Context) ([]string, error)
	GetCurrentSearchPathForDbConnection(context.Context, *sql.Conn) ([]string, error)
	SetRequiredSessionSearchPath(context.Context) error
	GetRequiredSessionSearchPath() []string
	ContructSearchPath(context.Context, []string, []string) ([]string, error)

	AcquireSession(context.Context) *AcquireSessionResult

	ExecuteSync(context.Context, string, ...any) (*queryresult.SyncQueryResult, error)
	Execute(context.Context, string, ...any) (*queryresult.Result, error)

	ExecuteSyncInSession(context.Context, *DatabaseSession, string, ...any) (*queryresult.SyncQueryResult, error)
	ExecuteInSession(context.Context, *DatabaseSession, func(), string, ...any) (*queryresult.Result, error)

	CacheOn(context.Context) error
	CacheOff(context.Context) error
	CacheClear(context.Context) error

	RefreshSessions(ctx context.Context) *AcquireSessionResult
	GetSchemaFromDB(context.Context) (*schema.Metadata, error)
	// remote client will have empty implementation
	RefreshConnectionAndSearchPaths(context.Context, ...string) *steampipeconfig.RefreshConnectionResult
}
