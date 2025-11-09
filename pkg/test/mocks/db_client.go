package mocks

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	pqueryresult "github.com/turbot/pipe-fittings/v2/queryresult"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/query/queryresult"
)

// MockClient is a mock implementation of db_common.Client for testing
type MockClient struct {
	// Function fields for configurable behavior
	CloseFunc                       func(context.Context) error
	LoadUserSearchPathFunc          func(context.Context) error
	SetRequiredSessionSearchPathFunc func(context.Context) error
	GetRequiredSessionSearchPathFunc func() []string
	GetCustomSearchPathFunc         func() []string
	AcquireManagementConnectionFunc func(context.Context) (*pgxpool.Conn, error)
	AcquireSessionFunc              func(context.Context) *db_common.AcquireSessionResult
	ExecuteSyncFunc                 func(context.Context, string, ...any) (*pqueryresult.SyncQueryResult, error)
	ExecuteFunc                     func(context.Context, string, ...any) (*queryresult.Result, error)
	ExecuteSyncInSessionFunc        func(context.Context, *db_common.DatabaseSession, string, ...any) (*pqueryresult.SyncQueryResult, error)
	ExecuteInSessionFunc            func(context.Context, *db_common.DatabaseSession, func(), string, ...any) (*queryresult.Result, error)
	ResetPoolsFunc                  func(context.Context)
	GetSchemaFromDBFunc             func(context.Context) (*db_common.SchemaMetadata, error)
	ServerSettingsFunc              func() *db_common.ServerSettings
	RegisterNotificationListenerFunc func(func(notification *pgconn.Notification))

	// Track calls
	CloseCalls                       int
	LoadUserSearchPathCalls          int
	SetRequiredSessionSearchPathCalls int
	GetRequiredSessionSearchPathCalls int
	GetCustomSearchPathCalls         int
	AcquireManagementConnectionCalls int
	AcquireSessionCalls              int
	ExecuteSyncCalls                 []ExecuteCall
	ExecuteCalls                     []ExecuteCall
	ExecuteSyncInSessionCalls        []ExecuteCall
	ExecuteInSessionCalls            []ExecuteCall
	ResetPoolsCalls                  int
	GetSchemaFromDBCalls             int
	ServerSettingsCalls              int
	RegisterNotificationListenerCalls int
}

// ExecuteCall tracks a single execute call
type ExecuteCall struct {
	SQL  string
	Args []any
}

// Ensure MockClient implements db_common.Client
var _ db_common.Client = (*MockClient)(nil)

// Close implements db_common.Client
func (m *MockClient) Close(ctx context.Context) error {
	m.CloseCalls++
	if m.CloseFunc != nil {
		return m.CloseFunc(ctx)
	}
	return nil
}

// LoadUserSearchPath implements db_common.Client
func (m *MockClient) LoadUserSearchPath(ctx context.Context) error {
	m.LoadUserSearchPathCalls++
	if m.LoadUserSearchPathFunc != nil {
		return m.LoadUserSearchPathFunc(ctx)
	}
	return nil
}

// SetRequiredSessionSearchPath implements db_common.Client
func (m *MockClient) SetRequiredSessionSearchPath(ctx context.Context) error {
	m.SetRequiredSessionSearchPathCalls++
	if m.SetRequiredSessionSearchPathFunc != nil {
		return m.SetRequiredSessionSearchPathFunc(ctx)
	}
	return nil
}

// GetRequiredSessionSearchPath implements db_common.Client
func (m *MockClient) GetRequiredSessionSearchPath() []string {
	m.GetRequiredSessionSearchPathCalls++
	if m.GetRequiredSessionSearchPathFunc != nil {
		return m.GetRequiredSessionSearchPathFunc()
	}
	return []string{"public"}
}

// GetCustomSearchPath implements db_common.Client
func (m *MockClient) GetCustomSearchPath() []string {
	m.GetCustomSearchPathCalls++
	if m.GetCustomSearchPathFunc != nil {
		return m.GetCustomSearchPathFunc()
	}
	return nil
}

// AcquireManagementConnection implements db_common.Client
func (m *MockClient) AcquireManagementConnection(ctx context.Context) (*pgxpool.Conn, error) {
	m.AcquireManagementConnectionCalls++
	if m.AcquireManagementConnectionFunc != nil {
		return m.AcquireManagementConnectionFunc(ctx)
	}
	return nil, nil
}

// AcquireSession implements db_common.Client
func (m *MockClient) AcquireSession(ctx context.Context) *db_common.AcquireSessionResult {
	m.AcquireSessionCalls++
	if m.AcquireSessionFunc != nil {
		return m.AcquireSessionFunc(ctx)
	}
	return &db_common.AcquireSessionResult{
		Session: &db_common.DatabaseSession{
			BackendPid: 12345,
			SearchPath: []string{"public"},
		},
	}
}

// ExecuteSync implements db_common.Client
func (m *MockClient) ExecuteSync(ctx context.Context, sql string, args ...any) (*pqueryresult.SyncQueryResult, error) {
	m.ExecuteSyncCalls = append(m.ExecuteSyncCalls, ExecuteCall{SQL: sql, Args: args})
	if m.ExecuteSyncFunc != nil {
		return m.ExecuteSyncFunc(ctx, sql, args...)
	}
	return &pqueryresult.SyncQueryResult{}, nil
}

// Execute implements db_common.Client
func (m *MockClient) Execute(ctx context.Context, sql string, args ...any) (*queryresult.Result, error) {
	m.ExecuteCalls = append(m.ExecuteCalls, ExecuteCall{SQL: sql, Args: args})
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, sql, args...)
	}
	// Return a properly initialized Result
	return queryresult.NewResult([]*pqueryresult.ColumnDef{}), nil
}

// ExecuteSyncInSession implements db_common.Client
func (m *MockClient) ExecuteSyncInSession(ctx context.Context, session *db_common.DatabaseSession, sql string, args ...any) (*pqueryresult.SyncQueryResult, error) {
	m.ExecuteSyncInSessionCalls = append(m.ExecuteSyncInSessionCalls, ExecuteCall{SQL: sql, Args: args})
	if m.ExecuteSyncInSessionFunc != nil {
		return m.ExecuteSyncInSessionFunc(ctx, session, sql, args...)
	}
	return &pqueryresult.SyncQueryResult{}, nil
}

// ExecuteInSession implements db_common.Client
func (m *MockClient) ExecuteInSession(ctx context.Context, session *db_common.DatabaseSession, onComplete func(), sql string, args ...any) (*queryresult.Result, error) {
	m.ExecuteInSessionCalls = append(m.ExecuteInSessionCalls, ExecuteCall{SQL: sql, Args: args})
	if m.ExecuteInSessionFunc != nil {
		return m.ExecuteInSessionFunc(ctx, session, onComplete, sql, args...)
	}
	return &queryresult.Result{}, nil
}

// ResetPools implements db_common.Client
func (m *MockClient) ResetPools(ctx context.Context) {
	m.ResetPoolsCalls++
	if m.ResetPoolsFunc != nil {
		m.ResetPoolsFunc(ctx)
	}
}

// GetSchemaFromDB implements db_common.Client
func (m *MockClient) GetSchemaFromDB(ctx context.Context) (*db_common.SchemaMetadata, error) {
	m.GetSchemaFromDBCalls++
	if m.GetSchemaFromDBFunc != nil {
		return m.GetSchemaFromDBFunc(ctx)
	}
	return &db_common.SchemaMetadata{}, nil
}

// ServerSettings implements db_common.Client
func (m *MockClient) ServerSettings() *db_common.ServerSettings {
	m.ServerSettingsCalls++
	if m.ServerSettingsFunc != nil {
		return m.ServerSettingsFunc()
	}
	return &db_common.ServerSettings{
		CacheEnabled: false,
	}
}

// RegisterNotificationListener implements db_common.Client
func (m *MockClient) RegisterNotificationListener(f func(notification *pgconn.Notification)) {
	m.RegisterNotificationListenerCalls++
	if m.RegisterNotificationListenerFunc != nil {
		m.RegisterNotificationListenerFunc(f)
	}
}
