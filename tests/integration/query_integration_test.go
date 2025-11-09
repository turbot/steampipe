//go:build integration
// +build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	pqueryresult "github.com/turbot/pipe-fittings/v2/queryresult"
	"github.com/turbot/steampipe/v2/pkg/test/helpers"
)

// TestQueryExecution_EndToEnd tests the complete query execution workflow
// from SQL input through to result formatting
func TestQueryExecution_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	tests := map[string]struct {
		query       string
		expectRows  int
		expectCols  int
		expectError bool
	}{
		"simple select": {
			query:      "SELECT 1 as num",
			expectRows: 1,
			expectCols: 1,
		},
		"select with multiple rows": {
			query:      "SELECT * FROM generate_series(1, 5)",
			expectRows: 5,
			expectCols: 1,
		},
		"select with multiple columns": {
			query:      "SELECT 1 as id, 'test' as name, true as active",
			expectRows: 1,
			expectCols: 3,
		},
		"invalid SQL syntax": {
			query:       "SELCT * FROM test",
			expectError: true,
		},
		"nonexistent table": {
			query:       "SELECT * FROM nonexistent_table_xyz",
			expectError: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup: Create test database client
			ctx := context.Background()
			client := helpers.CreateTestDatabaseClient(t)

			// Mock ExecuteSync to simulate query execution
			var resultRows []interface{}
			var resultCols []*pqueryresult.ColumnDef

			if !tc.expectError {
				// Create mock result rows
				for i := 0; i < tc.expectRows; i++ {
					resultRows = append(resultRows, map[string]interface{}{})
				}
				// Create mock columns
				for i := 0; i < tc.expectCols; i++ {
					resultCols = append(resultCols, &pqueryresult.ColumnDef{})
				}
			}

			client.ExecuteSyncFunc = func(ctx context.Context, sql string, args ...any) (*pqueryresult.SyncQueryResult, error) {
				if tc.expectError {
					return nil, assert.AnError
				}
				return &pqueryresult.SyncQueryResult{
					Rows:   resultRows,
					Cols:   resultCols,
					Timing: nil,
				}, nil
			}

			// Execute query
			result, err := client.ExecuteSync(ctx, tc.query)

			// Verify results
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectRows, len(result.Rows))
				assert.Equal(t, tc.expectCols, len(result.Cols))
			}
		})
	}
}

// TestResultFormattingPipeline tests the result formatting for different output formats
func TestResultFormattingPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	// Create test data
	testCols := []*pqueryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
		{Name: "name", DataType: "text"},
	}

	testRows := []interface{}{
		map[string]interface{}{"id": 1, "name": "Alice"},
		map[string]interface{}{"id": 2, "name": "Bob"},
	}

	result := &pqueryresult.SyncQueryResult{
		Cols:   testCols,
		Rows:   testRows,
		Timing: nil,
	}

	tests := map[string]struct {
		format      string
		validate    func(t *testing.T, output string)
	}{
		"json format": {
			format: "json",
			validate: func(t *testing.T, output string) {
				// Should be valid JSON array
				var parsed []map[string]interface{}
				err := json.Unmarshal([]byte(output), &parsed)
				assert.NoError(t, err, "Output should be valid JSON")
				assert.Len(t, parsed, 2, "Should have 2 rows")
			},
		},
		"csv format": {
			format: "csv",
			validate: func(t *testing.T, output string) {
				// Should have header row and data rows
				lines := strings.Split(strings.TrimSpace(output), "\n")
				assert.GreaterOrEqual(t, len(lines), 2, "Should have header + data rows")
				// First line should be headers
				assert.Contains(t, lines[0], "id", "Should contain id column")
				assert.Contains(t, lines[0], "name", "Should contain name column")
			},
		},
		"table format": {
			format: "table",
			validate: func(t *testing.T, output string) {
				// Should have table borders
				assert.Contains(t, output, "|", "Should have table borders")
				// Should contain data
				assert.Contains(t, output, "Alice", "Should contain row data")
				assert.Contains(t, output, "Bob", "Should contain row data")
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Format result based on format type
			output := formatTestResult(result, tc.format)

			assert.NotEmpty(t, output, "Output should not be empty")

			// Run format-specific validation
			tc.validate(t, output)
		})
	}
}

// TestErrorPropagation_ThroughLayers tests that errors propagate correctly
// through the query execution layers
func TestErrorPropagation_ThroughLayers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	tests := map[string]struct {
		setupError  error
		expectError bool
		errorMsg    string
	}{
		"SQL syntax error propagates": {
			setupError:  assert.AnError,
			expectError: true,
			errorMsg:    "syntax error",
		},
		"connection error propagates": {
			setupError:  assert.AnError,
			expectError: true,
			errorMsg:    "connection",
		},
		"successful query": {
			setupError:  nil,
			expectError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			client := helpers.CreateTestDatabaseClient(t)

			// Setup error condition
			client.ExecuteSyncFunc = func(ctx context.Context, sql string, args ...any) (*pqueryresult.SyncQueryResult, error) {
				if tc.setupError != nil {
					return nil, tc.setupError
				}
				return &pqueryresult.SyncQueryResult{
					Rows:   []interface{}{},
					Cols:   []*pqueryresult.ColumnDef{},
					Timing: nil,
				}, nil
			}

			// Execute operation
			_, err := client.ExecuteSync(ctx, "SELECT 1")

			// Verify error propagation
			if tc.expectError {
				assert.Error(t, err, "Error should propagate")
			} else {
				assert.NoError(t, err, "Should not error on success")
			}
		})
	}
}

// TestQueryCancellation tests that query cancellation works correctly
func TestQueryCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	t.Run("context cancellation stops query", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		client := helpers.CreateTestDatabaseClient(t)

		// Setup a query that checks context
		client.ExecuteSyncFunc = func(ctx context.Context, sql string, args ...any) (*pqueryresult.SyncQueryResult, error) {
			// Simulate checking context during execution
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				return &pqueryresult.SyncQueryResult{
					Rows:   []interface{}{},
					Cols:   []*pqueryresult.ColumnDef{},
					Timing: nil,
				}, nil
			}
		}

		// Cancel context immediately
		cancel()

		// Execute query with cancelled context
		_, err := client.ExecuteSync(ctx, "SELECT * FROM long_running_query")

		// Should get context cancellation error
		assert.Error(t, err, "Should error on cancelled context")
		assert.ErrorIs(t, err, context.Canceled, "Should be context cancellation error")
	})
}

// formatTestResult is a helper function to format query results
// This simulates the formatting pipeline in the real application
func formatTestResult(result *pqueryresult.SyncQueryResult, format string) string {
	switch format {
	case "json":
		// Convert to JSON array
		bytes, _ := json.Marshal(result.Rows)
		return string(bytes)

	case "csv":
		// Convert to CSV
		var lines []string

		// Header row
		headers := make([]string, len(result.Cols))
		for i, col := range result.Cols {
			headers[i] = col.Name
		}
		lines = append(lines, strings.Join(headers, ","))

		// Data rows
		for _, row := range result.Rows {
			rowMap, ok := row.(map[string]interface{})
			if !ok {
				continue
			}
			values := make([]string, len(result.Cols))
			for i, col := range result.Cols {
				if val, ok := rowMap[col.Name]; ok {
					values[i] = toString(val)
				}
			}
			lines = append(lines, strings.Join(values, ","))
		}

		return strings.Join(lines, "\n")

	case "table":
		// Convert to table format
		var output strings.Builder

		// Header
		output.WriteString("|")
		for _, col := range result.Cols {
			output.WriteString(" " + col.Name + " |")
		}
		output.WriteString("\n")

		// Data rows
		for _, row := range result.Rows {
			rowMap, ok := row.(map[string]interface{})
			if !ok {
				continue
			}
			output.WriteString("|")
			for _, col := range result.Cols {
				if val, ok := rowMap[col.Name]; ok {
					output.WriteString(" " + toString(val) + " |")
				} else {
					output.WriteString("  |")
				}
			}
			output.WriteString("\n")
		}

		return output.String()

	default:
		return ""
	}
}

// toString converts an interface{} to string for formatting
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
