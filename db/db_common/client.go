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
	LoadSchema()

	SchemaMetadata() *schema.Metadata
	ConnectionMap() *steampipeconfig.ConnectionDataMap

	GetCurrentSearchPath() ([]string, error)
	SetSessionSearchPath(...string) error
	ContructSearchPath(requiredSearchPath []string, searchPathPrefix []string, currentSearchPath []string) ([]string, error)

	AcquireSession(ctx context.Context) (*DatabaseSession, error, []string)

	ExecuteSync(ctx context.Context, query string, disableSpinner bool) (*queryresult.SyncQueryResult, error)
	Execute(ctx context.Context, query string, disableSpinner bool) (res *queryresult.Result, err error)

	ExecuteSyncInSession(ctx context.Context, session *DatabaseSession, query string, disableSpinner bool) (*queryresult.SyncQueryResult, error)
	ExecuteInSession(ctx context.Context, session *DatabaseSession, query string, onComplete func(), disableSpinner bool) (res *queryresult.Result, err error)

	CacheOn() error
	CacheOff() error
	CacheClear() error

	SetEnsureSessionDataFunc(EnsureSessionStateCallback)
	RefreshSession(ctx context.Context) (error, []string)

	// remote client will have empty implementation
	RefreshConnectionAndSearchPaths() *steampipeconfig.RefreshConnectionResult
}
