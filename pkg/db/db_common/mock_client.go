package db_common

import (
	"context"
	"database/sql"

	"github.com/turbot/steampipe/pkg/query/queryresult"
	"github.com/turbot/steampipe/pkg/schema"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
)

type MockClient struct{}

func (m *MockClient) Close(ctx context.Context) error                        { return nil }
func (m *MockClient) ForeignSchemaNames() []string                           { return []string{} }
func (m *MockClient) LoadForeignSchemaNames(ctx context.Context) error       { return nil }
func (m *MockClient) ConnectionMap() *steampipeconfig.ConnectionDataMap      { return nil }
func (m *MockClient) GetCurrentSearchPath(context.Context) ([]string, error) { return []string{}, nil }
func (m *MockClient) GetCurrentSearchPathForDbConnection(context.Context, *sql.Conn) ([]string, error) {
	return []string{}, nil
}
func (m *MockClient) SetRequiredSessionSearchPath(context.Context) error { return nil }
func (m *MockClient) GetRequiredSessionSearchPath() []string             { return []string{} }
func (m *MockClient) ContructSearchPath(context.Context, []string, []string) ([]string, error) {
	return []string{}, nil
}
func (m *MockClient) AcquireSession(context.Context) *AcquireSessionResult { return nil }
func (m *MockClient) ExecuteSync(context.Context, string) (*queryresult.SyncQueryResult, error) {
	return nil, nil
}
func (m *MockClient) Execute(context.Context, string) (*queryresult.Result, error) { return nil, nil }
func (m *MockClient) ExecuteSyncInSession(context.Context, *DatabaseSession, string) (*queryresult.SyncQueryResult, error) {
	return nil, nil
}
func (m *MockClient) ExecuteInSession(context.Context, *DatabaseSession, string, func()) (*queryresult.Result, error) {
	return nil, nil
}
func (m *MockClient) CacheOn(context.Context) error                             { return nil }
func (m *MockClient) CacheOff(context.Context) error                            { return nil }
func (m *MockClient) CacheClear(context.Context) error                          { return nil }
func (m *MockClient) RefreshSessions(ctx context.Context) *AcquireSessionResult { return nil }
func (m *MockClient) GetSchemaFromDB(context.Context) (*schema.Metadata, error) { return nil, nil }
func (m *MockClient) RefreshConnectionAndSearchPaths(context.Context) *steampipeconfig.RefreshConnectionResult {
	return nil
}
