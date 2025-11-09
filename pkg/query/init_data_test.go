package query

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/steampipe/v2/pkg/test/helpers"
)

// TestResolveQueryAndArgsFromSQLString_DirectSQL tests resolving a direct SQL string
func TestResolveQueryAndArgsFromSQLString_DirectSQL(t *testing.T) {
	tests := []struct {
		name        string
		sqlString   string
		wantSQL     string
		wantError   bool
	}{
		{
			name:        "simple select",
			sqlString:   "SELECT 1",
			wantSQL:     "SELECT 1",
			wantError:   false,
		},
		{
			name:        "select with where",
			sqlString:   "SELECT * FROM aws_s3_bucket WHERE name = 'test'",
			wantSQL:     "SELECT * FROM aws_s3_bucket WHERE name = 'test'",
			wantError:   false,
		},
		{
			name:        "multi-line query",
			sqlString:   "SELECT id,\n       name,\n       region\nFROM aws_s3_bucket",
			wantSQL:     "SELECT id,\n       name,\n       region\nFROM aws_s3_bucket",
			wantError:   false,
		},
		{
			name:        "query with comments",
			sqlString:   "-- This is a comment\nSELECT 1",
			wantSQL:     "-- This is a comment\nSELECT 1",
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveQueryAndArgsFromSQLString(tt.sqlString)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantSQL, result.ExecuteSQL)
				assert.Equal(t, tt.wantSQL, result.RawSQL)
			}
		})
	}
}

// TestResolveQueryAndArgsFromSQLString_FromFile tests resolving a query from a file
func TestResolveQueryAndArgsFromSQLString_FromFile(t *testing.T) {
	// Create a temporary directory
	tempDir := helpers.CreateTempDir(t)

	// Create test SQL files
	simpleFile := helpers.WriteTestFile(t, tempDir, "simple.sql", "SELECT 1")
	complexFile := helpers.WriteTestFile(t, tempDir, "complex.sql", "SELECT * FROM aws_s3_bucket WHERE region = 'us-east-1'")
	emptyFile := helpers.WriteTestFile(t, tempDir, "empty.sql", "")

	tests := []struct {
		name      string
		filePath  string
		wantSQL   string
		wantError bool
	}{
		{
			name:      "simple query file",
			filePath:  simpleFile,
			wantSQL:   "SELECT 1",
			wantError: false,
		},
		{
			name:      "complex query file",
			filePath:  complexFile,
			wantSQL:   "SELECT * FROM aws_s3_bucket WHERE region = 'us-east-1'",
			wantError: false,
		},
		{
			name:      "empty query file",
			filePath:  emptyFile,
			wantSQL:   "",
			wantError: false,
		},
		{
			name:      "non-existent file",
			filePath:  filepath.Join(tempDir, "nonexistent.sql"),
			wantSQL:   "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveQueryAndArgsFromSQLString(tt.filePath)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantSQL, result.ExecuteSQL)
				assert.Equal(t, tt.wantSQL, result.RawSQL)
			}
		})
	}
}

// TestResolveQueryAndArgsFromSQLString_FileNotFound tests error handling for non-existent .sql files
func TestResolveQueryAndArgsFromSQLString_FileNotFound(t *testing.T) {
	// Test with a .sql extension that doesn't exist
	result, err := ResolveQueryAndArgsFromSQLString("/nonexistent/path/query.sql")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
	assert.Nil(t, result)
}

// TestResolveQueryAndArgsFromSQLString_NotAFile tests handling when string is not a file
func TestResolveQueryAndArgsFromSQLString_NotAFile(t *testing.T) {
	// Test with a string that's not a file and doesn't have .sql extension
	sqlString := "SELECT * FROM table WHERE id = 1"
	result, err := ResolveQueryAndArgsFromSQLString(sqlString)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, sqlString, result.ExecuteSQL)
}

// TestGetQueryFromFile tests reading queries from files
func TestGetQueryFromFile(t *testing.T) {
	tempDir := helpers.CreateTempDir(t)
	testFile := helpers.WriteTestFile(t, tempDir, "test.sql", "SELECT * FROM test")

	tests := []struct {
		name       string
		input      string
		wantSQL    string
		wantExists bool
		wantError  bool
	}{
		{
			name:       "existing file",
			input:      testFile,
			wantSQL:    "SELECT * FROM test",
			wantExists: true,
			wantError:  false,
		},
		{
			name:       "non-existing file",
			input:      filepath.Join(tempDir, "nonexistent.sql"),
			wantSQL:    "",
			wantExists: false,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, exists, err := getQueryFromFile(tt.input)

			assert.Equal(t, tt.wantExists, exists)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if exists {
					assert.Equal(t, tt.wantSQL, result.ExecuteSQL)
					assert.Equal(t, tt.wantSQL, result.RawSQL)
				}
			}
		})
	}
}

// TestGetQueriesFromArgs tests converting args to resolved queries
func TestGetQueriesFromArgs(t *testing.T) {
	tempDir := helpers.CreateTempDir(t)
	file1 := helpers.WriteTestFile(t, tempDir, "query1.sql", "SELECT 1")
	file2 := helpers.WriteTestFile(t, tempDir, "query2.sql", "SELECT 2")

	tests := []struct {
		name        string
		args        []string
		wantCount   int
		wantError   bool
	}{
		{
			name:        "single direct SQL",
			args:        []string{"SELECT 1"},
			wantCount:   1,
			wantError:   false,
		},
		{
			name:        "multiple direct SQL",
			args:        []string{"SELECT 1", "SELECT 2", "SELECT 3"},
			wantCount:   3,
			wantError:   false,
		},
		{
			name:        "single file",
			args:        []string{file1},
			wantCount:   1,
			wantError:   false,
		},
		{
			name:        "multiple files",
			args:        []string{file1, file2},
			wantCount:   2,
			wantError:   false,
		},
		{
			name:        "mix of files and SQL",
			args:        []string{file1, "SELECT 99", file2},
			wantCount:   3,
			wantError:   false,
		},
		{
			name:        "empty args",
			args:        []string{},
			wantCount:   0,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries, err := getQueriesFromArgs(tt.args)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, queries, tt.wantCount)
				// Verify all queries have ExecuteSQL set
				for _, q := range queries {
					assert.NotEmpty(t, q.ExecuteSQL)
					assert.NotEmpty(t, q.Name)
				}
			}
		})
	}
}

// TestGetQueriesFromArgs_FileError tests error handling when reading files
func TestGetQueriesFromArgs_FileError(t *testing.T) {
	// Non-existent file with .sql extension should error
	args := []string{"/nonexistent/path/query.sql"}
	queries, err := getQueriesFromArgs(args)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
	assert.Nil(t, queries)
}

// TestGetQueryFromFile_EmptyFile tests handling of empty SQL files
func TestGetQueryFromFile_EmptyFile(t *testing.T) {
	tempDir := helpers.CreateTempDir(t)
	emptyFile := helpers.WriteTestFile(t, tempDir, "empty.sql", "")

	result, exists, err := getQueryFromFile(emptyFile)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NotNil(t, result)
	assert.Empty(t, result.ExecuteSQL)
}

// TestGetQueryFromFile_LargeFile tests handling of large SQL files
func TestGetQueryFromFile_LargeFile(t *testing.T) {
	tempDir := helpers.CreateTempDir(t)

	// Create a large query with lots of columns
	largeQuery := "SELECT\n"
	for i := 0; i < 100; i++ {
		if i > 0 {
			largeQuery += ",\n"
		}
		largeQuery += "  column" + string(rune('0'+i%10))
	}
	largeQuery += "\nFROM large_table"

	largeFile := helpers.WriteTestFile(t, tempDir, "large.sql", largeQuery)

	result, exists, err := getQueryFromFile(largeFile)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NotNil(t, result)
	assert.Equal(t, largeQuery, result.ExecuteSQL)
}

// TestResolveQueryAndArgsFromSQLString_SpecialCharacters tests handling special characters
func TestResolveQueryAndArgsFromSQLString_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name      string
		sqlString string
		wantSQL   string
	}{
		{
			name:      "query with quotes",
			sqlString: "SELECT * FROM table WHERE name = 'O''Brien'",
			wantSQL:   "SELECT * FROM table WHERE name = 'O''Brien'",
		},
		{
			name:      "query with semicolon",
			sqlString: "SELECT 1; SELECT 2;",
			wantSQL:   "SELECT 1; SELECT 2;",
		},
		{
			name:      "query with special column names",
			sqlString: "SELECT \"column-with-dash\", \"column.with.dot\" FROM table",
			wantSQL:   "SELECT \"column-with-dash\", \"column.with.dot\" FROM table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveQueryAndArgsFromSQLString(tt.sqlString)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantSQL, result.ExecuteSQL)
		})
	}
}

// TestGetQueryFromFile_WithWhitespace tests handling files with only whitespace
func TestGetQueryFromFile_WithWhitespace(t *testing.T) {
	tempDir := helpers.CreateTempDir(t)
	whitespaceFile := helpers.WriteTestFile(t, tempDir, "whitespace.sql", "   \n  \t  \n  ")

	result, exists, err := getQueryFromFile(whitespaceFile)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NotNil(t, result)
	// Whitespace should be preserved in the query
	assert.Contains(t, result.ExecuteSQL, " ")
}

// TestInitData_Cancel tests the Cancel method
func TestInitData_Cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	initData := &InitData{
		cancelInitialisation: cancel,
	}

	// Cancel should not panic when cancelInitialisation is set
	initData.Cancel()

	// Verify context is cancelled
	assert.Error(t, ctx.Err())

	// Cancel should not panic when called again
	initData.Cancel()
}

// TestInitData_Cancel_NilCancelFunc tests Cancel with nil cancel function
func TestInitData_Cancel_NilCancelFunc(t *testing.T) {
	initData := &InitData{
		cancelInitialisation: nil,
	}

	// Cancel should not panic when cancelInitialisation is nil
	assert.NotPanics(t, func() {
		initData.Cancel()
	})
}

// TestGetQueryFromFile_RelativePath tests handling relative file paths
func TestGetQueryFromFile_RelativePath(t *testing.T) {
	// Create a temp directory and change to it
	tempDir := helpers.CreateTempDir(t)
	originalWd, err := os.Getwd()
	assert.NoError(t, err)

	err = os.Chdir(tempDir)
	assert.NoError(t, err)
	defer func() {
		err := os.Chdir(originalWd)
		assert.NoError(t, err)
	}()

	// Write a file with a relative path
	err = os.WriteFile("relative.sql", []byte("SELECT * FROM relative"), 0644)
	assert.NoError(t, err)

	result, exists, err := getQueryFromFile("relative.sql")

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NotNil(t, result)
	assert.Equal(t, "SELECT * FROM relative", result.ExecuteSQL)
}

// TestResolveQueryAndArgsFromSQLString_ComplexQueries tests complex SQL queries
func TestResolveQueryAndArgsFromSQLString_ComplexQueries(t *testing.T) {
	tests := []struct {
		name      string
		sqlString string
	}{
		{
			name: "query with CTE",
			sqlString: `WITH regional_sales AS (
				SELECT region, SUM(amount) as total_sales
				FROM orders
				GROUP BY region
			)
			SELECT * FROM regional_sales WHERE total_sales > 1000`,
		},
		{
			name: "query with joins",
			sqlString: `SELECT a.*, b.name
				FROM aws_s3_bucket a
				JOIN aws_iam_policy b ON a.policy_arn = b.arn
				WHERE a.region = 'us-east-1'`,
		},
		{
			name: "query with subquery",
			sqlString: `SELECT name FROM aws_s3_bucket
				WHERE name IN (SELECT bucket_name FROM aws_s3_bucket_public)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveQueryAndArgsFromSQLString(tt.sqlString)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.sqlString, result.ExecuteSQL)
			assert.Equal(t, tt.sqlString, result.RawSQL)
		})
	}
}

// TestQueryExporters tests the queryExporters function
func TestQueryExporters(t *testing.T) {
	exporters := queryExporters()

	assert.NotNil(t, exporters)
	assert.Len(t, exporters, 1)
	// Should return a SnapshotExporter
	assert.NotNil(t, exporters[0])
}

// TestGetQueriesFromArgs_PreservesOrder tests that query order is preserved
func TestGetQueriesFromArgs_PreservesOrder(t *testing.T) {
	args := []string{
		"SELECT 1 as first",
		"SELECT 2 as second",
		"SELECT 3 as third",
	}

	queries, err := getQueriesFromArgs(args)

	assert.NoError(t, err)
	assert.Len(t, queries, 3)
	assert.Contains(t, queries[0].ExecuteSQL, "first")
	assert.Contains(t, queries[1].ExecuteSQL, "second")
	assert.Contains(t, queries[2].ExecuteSQL, "third")
}

// TestResolveQueryAndArgsFromSQLString_UnicodeCharacters tests handling Unicode
func TestResolveQueryAndArgsFromSQLString_UnicodeCharacters(t *testing.T) {
	sqlString := "SELECT * FROM table WHERE name = '日本語'"

	result, err := ResolveQueryAndArgsFromSQLString(sqlString)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, sqlString, result.ExecuteSQL)
	assert.Contains(t, result.ExecuteSQL, "日本語")
}

// TestGetQueryFromFile_PermissionDenied tests handling files without read permissions
func TestGetQueryFromFile_PermissionDenied(t *testing.T) {
	// Skip on Windows as permission handling is different
	if os.PathSeparator == '\\' {
		t.Skip("Skipping permission test on Windows")
	}

	tempDir := helpers.CreateTempDir(t)
	noReadFile := filepath.Join(tempDir, "noread.sql")

	err := os.WriteFile(noReadFile, []byte("SELECT 1"), 0000)
	assert.NoError(t, err)
	defer os.Chmod(noReadFile, 0644) // Restore permissions for cleanup

	result, exists, err := getQueryFromFile(noReadFile)

	// Should return exists=true but with an error when trying to read
	assert.Error(t, err)
	assert.True(t, exists)
	assert.Nil(t, result)
}
