package db_common

import (
	"context"
	"database/sql"
	"github.com/turbot/steampipe/pkg/query/queryresult"
	"github.com/turbot/steampipe/pkg/schema"
)

type Client interface {
	Close(ctx context.Context) error

	ForeignSchemaNames() []string
	AllSchemaNames() []string
	LoadSchemaNames(ctx context.Context) error

	GetCurrentSearchPath(context.Context) ([]string, error)
	GetCurrentSearchPathForDbConnection(context.Context, *sql.Conn) ([]string, error)
	SetRequiredSessionSearchPath(context.Context) error
	GetRequiredSessionSearchPath() []string
	ConstructSearchPath(context.Context, []string, []string) ([]string, error)

	AcquireSession(context.Context) *AcquireSessionResult

	ExecuteSync(context.Context, string, ...any) (*queryresult.SyncQueryResult, error)
	Execute(context.Context, string, ...any) (*queryresult.Result, error)

	ExecuteSyncInSession(context.Context, *DatabaseSession, string, ...any) (*queryresult.SyncQueryResult, error)
	ExecuteInSession(context.Context, *DatabaseSession, func(), string, ...any) (*queryresult.Result, error)

	CacheOn(context.Context) error
	CacheOff(context.Context) error
	CacheClear(context.Context) error

	RefreshSessions(ctx context.Context) *AcquireSessionResult
	GetSchemaFromDB(context.Context, ...string) (*schema.Metadata, error)
}
