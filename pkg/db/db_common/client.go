package db_common

import (
	"context"
	"github.com/turbot/steampipe/pkg/query/queryresult"
	"github.com/turbot/steampipe/pkg/schema"
)

type Client interface {
	Close(ctx context.Context) error

	//ForeignSchemaNames() []string
	//AllSchemaNames() []string
	LoadSchemaNames(ctx context.Context) error

	GetCurrentSearchPath(context.Context) ([]string, error)
	SetRequiredSessionSearchPath(context.Context) error
	GetRequiredSessionSearchPath(context.Context) ([]string, error)

	AcquireSession(context.Context) *AcquireSessionResult

	ExecuteSync(context.Context, string, ...any) (*queryresult.SyncQueryResult, error)
	Execute(context.Context, string, ...any) (*queryresult.Result, error)

	ExecuteSyncInSession(context.Context, *DatabaseSession, string, ...any) (*queryresult.SyncQueryResult, error)
	ExecuteInSession(context.Context, *DatabaseSession, func(), string, ...any) (*queryresult.Result, error)

	RefreshSessions(context.Context) *AcquireSessionResult
	GetSchemaFromDB(context.Context, ...string) (*schema.Metadata, error)
}
