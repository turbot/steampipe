package db_common

import (
	"context"

	"github.com/turbot/steampipe/steampipeconfig"

	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/schema"
)

type Client interface {
	Close() error
	LoadSchema()

	SchemaMetadata() *schema.Metadata
	ConnectionMap() *steampipeconfig.ConnectionMap

	GetCurrentSearchPath() ([]string, error)
	SetClientSearchPath(...string) error

	ExecuteSync(ctx context.Context, query string, disableSpinner bool) (*queryresult.SyncQueryResult, error)
	Execute(ctx context.Context, query string, disableSpinner bool) (res *queryresult.Result, err error)

	CacheOn() error
	CacheOff() error
	CacheClear() error

	// remote client will have empty implementation

	RefreshConnectionAndSearchPaths() *RefreshConnectionResult
}
