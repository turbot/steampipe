package db_common

import (
	"context"

	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/steampipeconfig"

	"github.com/turbot/steampipe/query/queryresult"
)

type Client interface {
	Close() error
	GetCurrentSearchPath() ([]string, error)
	SetClientSearchPath() error
	ExecuteSync(ctx context.Context, query string, disableSpinner bool) (*queryresult.SyncQueryResult, error)
	Execute(ctx context.Context, query string, disableSpinner bool) (res *queryresult.Result, err error)
	SchemaMetadata() *schema.Metadata

	CacheOn() error
	CacheOff() error
	CacheClear() error

	// todo share this between locan and remote client?
	LoadSchema()

	// RemoteClient will have empty implementations

	ConnectionMap() *steampipeconfig.ConnectionMap
	RefreshConnectionAndSearchPaths() *RefreshConnectionResult
}
