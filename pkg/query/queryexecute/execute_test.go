package queryexecute

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	pqueryresult "github.com/turbot/pipe-fittings/v2/queryresult"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/export"
	"github.com/turbot/steampipe/v2/pkg/initialisation"
	"github.com/turbot/steampipe/v2/pkg/query"
	"github.com/turbot/steampipe/v2/pkg/query/queryresult"
)

// Test Helpers

// createMockInitData creates a mock InitData for testing
func createMockInitData(t *testing.T) *query.InitData {
	t.Helper()

	initData := &query.InitData{
		InitData: initialisation.InitData{
			Result:        &db_common.InitResult{},
			ExportManager: export.NewManager(),
			Client:        &mockClient{}, // Add mock client to prevent nil pointer panics
		},
		Loaded:    make(chan struct{}),
		StartTime: time.Now(),
		Queries:   []*modconfig.ResolvedQuery{},
	}

	return initData
}

// closeInitDataLoaded closes the Loaded channel to simulate initialization completion
func closeInitDataLoaded(initData *query.InitData) {
	select {
	case <-initData.Loaded:
		// already closed
	default:
		close(initData.Loaded)
	}
}

// Test Suite: RunBatchSession

func TestRunBatchSession_NilInitData(t *testing.T) {
	ctx := context.Background()

	// This should not panic - function should validate initData is non-nil
	failures, err := RunBatchSession(ctx, nil)

	if err == nil {
		t.Fatal("Expected error when initData is nil, got nil")
	}

	if failures != 0 {
		t.Errorf("Expected 0 failures when initData is nil, got %d", failures)
	}
}

func TestRunBatchSession_EmptyQueries(t *testing.T) {
	// ARRANGE: Create initData with no queries
	ctx := context.Background()
	initData := createMockInitData(t)
	initData.Queries = []*modconfig.ResolvedQuery{} // explicitly empty

	// Simulate successful initialization
	closeInitDataLoaded(initData)

	// ACT: Run batch session
	failures, err := RunBatchSession(ctx, initData)

	// ASSERT: Should return 0 failures and no error
	assert.NoError(t, err, "RunBatchSession should not error with empty queries")
	assert.Equal(t, 0, failures, "Should return 0 failures when no queries to execute")
}

func TestRunBatchSession_InitError(t *testing.T) {
	// ARRANGE: Create initData with an initialization error
	ctx := context.Background()
	initData := createMockInitData(t)

	// Simulate initialization error
	expectedErr := assert.AnError
	initData.Result.Error = expectedErr
	closeInitDataLoaded(initData)

	// ACT: Run batch session
	failures, err := RunBatchSession(ctx, initData)

	// ASSERT: Should return the init error immediately
	assert.Equal(t, expectedErr, err, "Should return initialization error")
	assert.Equal(t, 0, failures, "Should return 0 failures when init fails")
}

// TestRunBatchSession_NilClient tests that RunBatchSession handles nil Client gracefully
func TestRunBatchSession_NilClient(t *testing.T) {
	// Create initData with nil Client
	initData := &query.InitData{
		InitData: initialisation.InitData{
			Result: &db_common.InitResult{},
			Client: nil, // nil Client should be handled gracefully
		},
		Loaded: make(chan struct{}),
	}

	// Signal that init is complete
	close(initData.Loaded)

	// This should not panic - it should handle nil Client gracefully
	_, err := RunBatchSession(context.Background(), initData)

	// We expect an error indicating that Client is required, not a panic
	if err == nil {
		t.Error("Expected error when Client is nil, got nil")
	}
}

// TestRunBatchSession_LoadedTimeout demonstrates that RunBatchSession blocks forever
// if initData.Loaded never closes, even when the context is cancelled.
// References issue #4781
func TestRunBatchSession_LoadedTimeout(t *testing.T) {

	// Create a context with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Create InitData with a Loaded channel that will never close
	initData := &query.InitData{
		InitData: initialisation.InitData{
			Result: &db_common.InitResult{},
		},
		Loaded: make(chan struct{}), // This channel will never close
	}

	// This should return within the timeout, but currently blocks forever
	done := make(chan bool)
	var failures int
	var err error

	go func() {
		failures, err = RunBatchSession(ctx, initData)
		done <- true
	}()

	select {
	case <-done:
		// Function returned, check that it returned an error due to context cancellation
		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
		assert.Equal(t, 0, failures)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("RunBatchSession blocked forever despite context cancellation - bug #4781")
	}
}

// Test Suite: Helper Functions

func TestNeedSnapshot_DefaultValues(t *testing.T) {
	// This test verifies the needSnapshot function behavior with default config
	// Note: This is a simple test but ensures the function doesn't panic

	// ACT: Call needSnapshot with default viper config
	result := needSnapshot()

	// ASSERT: Should return false with default settings
	assert.False(t, result, "needSnapshot should return false with default settings")
}

func TestShowBlankLineBetweenResults_DefaultValues(t *testing.T) {
	// This test verifies showBlankLineBetweenResults function with default config

	// ACT: Call function with default viper config
	result := showBlankLineBetweenResults()

	// ASSERT: Should return true with default settings (not CSV without header)
	assert.True(t, result, "Should show blank lines with default settings")
}

func TestHandlePublishSnapshotError_PaymentRequired(t *testing.T) {
	// ARRANGE: Create a 402 Payment Required error
	err := assert.AnError
	err = &mockError{msg: "402 Payment Required"}

	// ACT: Handle the error
	result := handlePublishSnapshotError(err)

	// ASSERT: Should reword the error message
	assert.Error(t, result)
	assert.Contains(t, result.Error(), "maximum number of snapshots reached")
}

func TestHandlePublishSnapshotError_OtherError(t *testing.T) {
	// ARRANGE: Create a different error
	err := assert.AnError

	// ACT: Handle the error
	result := handlePublishSnapshotError(err)

	// ASSERT: Should return the error unchanged
	assert.Equal(t, err, result)
}

// Test Suite: Edge Cases and Resource Management

func TestExecuteQueries_EmptyQueriesList(t *testing.T) {
	// ARRANGE: InitData with empty queries list
	ctx := context.Background()
	initData := createMockInitData(t)
	initData.Queries = []*modconfig.ResolvedQuery{}

	// ACT: Execute queries directly
	failures := executeQueries(ctx, initData)

	// ASSERT: Should return 0 failures
	assert.Equal(t, 0, failures, "Should return 0 failures for empty queries list")
}

// TestExecuteQueries_NilClient tests that executeQueries handles nil Client gracefully
// Related to issue #4797
func TestExecuteQueries_NilClient(t *testing.T) {
	ctx := context.Background()

	// Create initData with nil Client but with queries
	// This simulates a scenario where initialization failed but queries were still provided
	initData := &query.InitData{
		InitData: *initialisation.NewInitData(),
		Queries: []*modconfig.ResolvedQuery{
			{
				Name:       "test_query",
				ExecuteSQL: "SELECT 1",
				RawSQL:     "SELECT 1",
			},
		},
	}
	// Explicitly set Client to nil to test the nil case
	initData.Client = nil

	// This should not panic - it should handle nil Client gracefully
	// Currently this will panic with nil pointer dereference
	failures := executeQueries(ctx, initData)

	// We expect 1 failure (the query should fail gracefully, not panic)
	if failures != 1 {
		t.Errorf("Expected 1 failure with nil client, got %d", failures)
	}
}

// Test Suite: Context and Cancellation

func TestRunBatchSession_CancelHandlerSetup(t *testing.T) {
	// This test verifies that the cancel handler doesn't cause panics
	// We can't easily test the actual cancellation behavior without integration tests

	// ARRANGE
	ctx := context.Background()
	initData := createMockInitData(t)
	closeInitDataLoaded(initData)

	// ACT: Run batch session
	// Note: This test just verifies no panic occurs when setting up cancel handler
	assert.NotPanics(t, func() {
		_, _ = RunBatchSession(ctx, initData)
	}, "Should not panic when setting up cancel handler")
}

// Test Suite: Result Wrapping

func TestWrapResult_NotNil(t *testing.T) {
	// This test ensures WrapResult doesn't panic and returns a valid wrapper

	// ARRANGE: Create a basic result from pipe-fittings
	// Note: We need to use the pipe-fittings queryresult package
	// This test verifies the wrapper functionality exists and doesn't panic
	wrapped := queryresult.NewResult(nil)

	// ASSERT: Should return a valid result
	assert.NotNil(t, wrapped, "NewResult should not return nil")
}

// Mock Types

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

// mockClient is a minimal mock implementation of db_common.Client for testing
type mockClient struct {
	customSearchPath   []string
	requiredSearchPath []string
}

func (m *mockClient) Close(ctx context.Context) error {
	return nil
}

func (m *mockClient) LoadUserSearchPath(ctx context.Context) error {
	return nil
}

func (m *mockClient) SetRequiredSessionSearchPath(ctx context.Context) error {
	return nil
}

func (m *mockClient) GetRequiredSessionSearchPath() []string {
	return m.requiredSearchPath
}

func (m *mockClient) GetCustomSearchPath() []string {
	return m.customSearchPath
}

func (m *mockClient) AcquireManagementConnection(ctx context.Context) (*pgxpool.Conn, error) {
	return nil, nil
}

func (m *mockClient) AcquireSession(ctx context.Context) *db_common.AcquireSessionResult {
	return nil
}

func (m *mockClient) ExecuteSync(ctx context.Context, query string, args ...any) (*pqueryresult.SyncQueryResult, error) {
	return nil, nil
}

func (m *mockClient) Execute(ctx context.Context, query string, args ...any) (*queryresult.Result, error) {
	return nil, nil
}

func (m *mockClient) ExecuteSyncInSession(ctx context.Context, session *db_common.DatabaseSession, query string, args ...any) (*pqueryresult.SyncQueryResult, error) {
	return nil, nil
}

func (m *mockClient) ExecuteInSession(ctx context.Context, session *db_common.DatabaseSession, onConnectionLost func(), query string, args ...any) (*queryresult.Result, error) {
	return nil, nil
}

func (m *mockClient) ResetPools(ctx context.Context) {
}

func (m *mockClient) GetSchemaFromDB(ctx context.Context) (*db_common.SchemaMetadata, error) {
	return nil, nil
}

func (m *mockClient) ServerSettings() *db_common.ServerSettings {
	return nil
}

func (m *mockClient) RegisterNotificationListener(f func(notification *pgconn.Notification)) {
}
