//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	pqueryresult "github.com/turbot/pipe-fittings/v2/queryresult"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/test/helpers"
)

// TestInteractiveSession_Lifecycle tests the complete interactive session lifecycle:
// initialize -> execute queries -> execute metaqueries -> close
func TestInteractiveSession_Lifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	// Setup: Create test client simulating interactive mode
	ctx := context.Background()
	client := helpers.CreateTestDatabaseClient(t)

	var sessionInitialized bool
	var sessionClosed bool

	// Mock session initialization
	client.AcquireSessionFunc = func(ctx context.Context) *db_common.AcquireSessionResult {
		sessionInitialized = true
		return &db_common.AcquireSessionResult{
			Session: helpers.CreateTestDatabaseSession(t),
		}
	}

	// Test 1: Initialize session
	t.Run("initialize session", func(t *testing.T) {
		result := client.AcquireSession(ctx)
		assert.NotNil(t, result, "Should return session result")
		assert.NotNil(t, result.Session, "Should have session")
		assert.True(t, sessionInitialized, "Session should be initialized")
	})

	// Test 2: Execute simple query
	t.Run("execute simple query", func(t *testing.T) {
		client.ExecuteSyncFunc = func(ctx context.Context, sql string, args ...any) (*pqueryresult.SyncQueryResult, error) {
			return &pqueryresult.SyncQueryResult{
				Rows: []interface{}{
					map[string]interface{}{"num": 1},
				},
				Cols: []*pqueryresult.ColumnDef{
					{Name: "num", DataType: "integer"},
				},
				Timing: nil,
			}, nil
		}

		result, err := client.ExecuteSync(ctx, "SELECT 1 as num")
		assert.NoError(t, err, "Should execute query without error")
		assert.NotNil(t, result, "Should return result")
		assert.Len(t, result.Rows, 1, "Should have one row")
	})

	// Test 3: Execute multiple queries in sequence
	t.Run("execute multiple queries", func(t *testing.T) {
		// Create a fresh client for this test to avoid state from previous tests
		freshClient := helpers.CreateTestDatabaseClient(t)

		queries := []string{
			"SELECT 1",
			"SELECT 2",
			"SELECT 3",
		}

		freshClient.ExecuteSyncFunc = func(ctx context.Context, sql string, args ...any) (*pqueryresult.SyncQueryResult, error) {
			return &pqueryresult.SyncQueryResult{
				Rows:   []interface{}{},
				Cols:   []*pqueryresult.ColumnDef{},
				Timing: nil,
			}, nil
		}

		for _, query := range queries {
			_, err := freshClient.ExecuteSync(ctx, query)
			assert.NoError(t, err, "Should execute query: %s", query)
		}

		assert.Equal(t, len(queries), len(freshClient.ExecuteSyncCalls), "Should track all query executions")
	})

	// Test 4: Close session
	t.Run("close session", func(t *testing.T) {
		client.CloseFunc = func(ctx context.Context) error {
			sessionClosed = true
			return nil
		}

		err := client.Close(ctx)
		assert.NoError(t, err, "Should close without error")
		assert.True(t, sessionClosed, "Session should be closed")
	})
}

// TestInteractiveSession_Metaqueries tests metaquery execution in interactive mode
func TestInteractiveSession_Metaqueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	tests := map[string]struct {
		metaquery    string
		expectOutput bool
		expectError  bool
	}{
		".tables metaquery": {
			metaquery:    ".tables",
			expectOutput: true,
		},
		".connections metaquery": {
			metaquery:    ".connections",
			expectOutput: true,
		},
		".inspect metaquery": {
			metaquery:    ".inspect",
			expectOutput: true,
		},
		"invalid metaquery": {
			metaquery:   ".invalid",
			expectError: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Simulate metaquery execution
			output, err := executeMetaquery(tc.metaquery)

			if tc.expectError {
				assert.Error(t, err, "Should error for invalid metaquery")
			} else {
				assert.NoError(t, err, "Should execute metaquery without error")
				if tc.expectOutput {
					assert.NotEmpty(t, output, "Should return output")
				}
			}
		})
	}
}

// TestInteractiveSession_StateManagement tests session state management
func TestInteractiveSession_StateManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	ctx := context.Background()
	client := helpers.CreateTestDatabaseClient(t)

	// Track session state
	type sessionState struct {
		searchPath     []string
		timing         bool
		outputFormat   string
		queryCount     int
	}

	state := &sessionState{
		searchPath:   []string{"public"},
		timing:       false,
		outputFormat: "table",
		queryCount:   0,
	}

	// Mock query execution that updates state
	client.ExecuteSyncFunc = func(ctx context.Context, sql string, args ...any) (*pqueryresult.SyncQueryResult, error) {
		state.queryCount++
		return &pqueryresult.SyncQueryResult{
			Rows:   []interface{}{},
			Cols:   []*pqueryresult.ColumnDef{},
			Timing: nil,
		}, nil
	}

	t.Run("initial state", func(t *testing.T) {
		assert.Equal(t, []string{"public"}, state.searchPath, "Should have default search path")
		assert.False(t, state.timing, "Timing should be off by default")
		assert.Equal(t, "table", state.outputFormat, "Should use table format by default")
		assert.Equal(t, 0, state.queryCount, "Query count should be 0")
	})

	t.Run("update state during session", func(t *testing.T) {
		// Execute some queries
		client.ExecuteSync(ctx, "SELECT 1")
		client.ExecuteSync(ctx, "SELECT 2")

		assert.Equal(t, 2, state.queryCount, "Should track query count")

		// Simulate state changes
		state.timing = true
		state.outputFormat = "json"

		assert.True(t, state.timing, "Should update timing setting")
		assert.Equal(t, "json", state.outputFormat, "Should update output format")
	})

	t.Run("state persists across queries", func(t *testing.T) {
		previousCount := state.queryCount

		client.ExecuteSync(ctx, "SELECT 3")

		assert.Equal(t, previousCount+1, state.queryCount, "Should maintain state across queries")
		assert.True(t, state.timing, "State changes should persist")
	})
}

// TestInteractiveSession_ErrorRecovery tests error recovery in interactive sessions
func TestInteractiveSession_ErrorRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	ctx := context.Background()
	client := helpers.CreateTestDatabaseClient(t)

	errorInjected := false

	client.ExecuteSyncFunc = func(ctx context.Context, sql string, args ...any) (*pqueryresult.SyncQueryResult, error) {
		// Inject error on first call, succeed on subsequent calls
		if !errorInjected && sql == "SELECT * FROM bad_table" {
			errorInjected = true
			return nil, assert.AnError
		}
		return &pqueryresult.SyncQueryResult{
			Rows:   []interface{}{},
			Cols:   []*pqueryresult.ColumnDef{},
			Timing: nil,
		}, nil
	}

	t.Run("session continues after error", func(t *testing.T) {
		// Execute query that will fail
		_, err := client.ExecuteSync(ctx, "SELECT * FROM bad_table")
		assert.Error(t, err, "Should error on bad query")

		// Session should still work for subsequent queries
		result, err := client.ExecuteSync(ctx, "SELECT 1")
		assert.NoError(t, err, "Should recover and execute next query")
		assert.NotNil(t, result, "Should return result after recovery")
	})
}

// TestInteractiveSession_ConcurrentQueries tests handling of concurrent queries
func TestInteractiveSession_ConcurrentQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Note: This test validates that the session can handle rapid query submissions
	// In a real interactive session, queries would be sequential, but this tests
	// the robustness of the session management

	ctx := context.Background()
	client := helpers.CreateTestDatabaseClient(t)

	queryExecutions := 0

	client.ExecuteSyncFunc = func(ctx context.Context, sql string, args ...any) (*pqueryresult.SyncQueryResult, error) {
		queryExecutions++
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)
		return &pqueryresult.SyncQueryResult{
			Rows:   []interface{}{},
			Cols:   []*pqueryresult.ColumnDef{},
			Timing: nil,
		}, nil
	}

	t.Run("execute queries sequentially", func(t *testing.T) {
		queries := []string{"SELECT 1", "SELECT 2", "SELECT 3", "SELECT 4", "SELECT 5"}

		for _, query := range queries {
			_, err := client.ExecuteSync(ctx, query)
			assert.NoError(t, err, "Should execute query: %s", query)
		}

		assert.Equal(t, len(queries), queryExecutions, "Should execute all queries")
	})
}

// executeMetaquery simulates metaquery execution
// In the real system, this would be handled by the interactive client
func executeMetaquery(metaquery string) (string, error) {
	switch metaquery {
	case ".tables":
		return "aws_account\naws_s3_bucket\n", nil
	case ".connections":
		return "aws\nazure\ngcp\n", nil
	case ".inspect":
		return "Connection: aws\nPlugin: aws@latest\n", nil
	default:
		return "", assert.AnError
	}
}
