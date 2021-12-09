package db_common

import (
	"context"

	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/steampipeconfig"
)

type EnsureSessionStateCallback = func(context.Context, *DatabaseSession) (err error, warnings []string)

type Client interface {
	Close() error

	ForeignSchemas() []string
	ConnectionMap() *steampipeconfig.ConnectionDataMap

	GetCurrentSearchPath(context.Context) ([]string, error)
	SetSessionSearchPath(ctx context.Context, newSearchPath ...string) error
	ContructSearchPath(ctx context.Context, requiredSearchPath []string, searchPathPrefix []string, currentSearchPath []string) ([]string, error)

	AcquireSession(context.Context) *AcquireSessionResult

	ExecuteSync(ctx context.Context, query string, disableSpinner bool) (*queryresult.SyncQueryResult, error)
	Execute(ctx context.Context, query string, disableSpinner bool) (res *queryresult.Result, err error)

	ExecuteSyncInSession(ctx context.Context, session *DatabaseSession, query string, disableSpinner bool) (*queryresult.SyncQueryResult, error)
	ExecuteInSession(ctx context.Context, session *DatabaseSession, query string, onComplete func(), disableSpinner bool) (res *queryresult.Result, err error)

	CacheOn(context.Context) error
	CacheOff(context.Context) error
	CacheClear(context.Context) error

	SetEnsureSessionDataFunc(EnsureSessionStateCallback)
	RefreshSessions(ctx context.Context) *AcquireSessionResult
	GetSchemaFromDB(context.Context, []string) (*schema.Metadata, error)
	// remote client will have empty implementation
	RefreshConnectionAndSearchPaths(context.Context) *steampipeconfig.RefreshConnectionResult
	LoadForeignSchemaNames(context.Context) error
}
