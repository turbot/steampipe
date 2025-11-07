# Task 3: Query Execution Tests

## Context
Query execution is THE core functionality of Steampipe. Users run SQL queries against APIs - if this breaks, Steampipe is useless. The execute.go file has 31 commits and cmd/query.go has 46 commits, making these high-risk files.

## Goal
Create comprehensive unit tests for query execution in `pkg/query/queryexecute/` and related files.

## Why This Matters
- Query execution is Steampipe's primary purpose
- 31 commits to execute.go = very high change frequency
- 46 commits to cmd/query.go = extremely high risk
- Every user interaction involves query execution
- Complex code with batch mode, interactive mode, multiple output formats

## Files to Test

### Primary Files
1. **`pkg/query/queryexecute/execute.go`** (31 changes)
   - Execute() - Main execution logic
   - Batch query execution
   - Interactive query execution
   - Result streaming
   - Error handling

2. **`pkg/query/init_data.go`** (19 changes)
   - Query initialization
   - Database connection setup
   - Configuration loading
   - Workspace setup

3. **`pkg/db/db_client/db_client_execute.go`** (19 changes)
   - ExecuteQuery() - Database execution
   - Result fetching
   - Retry logic
   - Transaction handling

4. **`cmd/query.go`** (46 changes)
   - Command setup
   - Argument parsing
   - Interactive vs batch mode
   - Output format handling

## Test Files to Create

### 1. execute_test.go

```go
package queryexecute

import (
    "testing"
    "context"
    "github.com/turbot/steampipe/pkg/test/helpers"
    "github.com/turbot/steampipe/pkg/test/mocks"
)

func TestExecute_SingleQuery_Success(t *testing.T) {
    tests := map[string]struct {
        query       string
        mockRows    []map[string]interface{}
        wantError   bool
    }{
        "simple select": {
            query: "SELECT 1",
            mockRows: []map[string]interface{}{
                {"?column?": 1},
            },
            wantError: false,
        },
        "select with where": {
            query: "SELECT * FROM aws_s3_bucket WHERE name = 'test'",
            mockRows: []map[string]interface{}{
                {"name": "test", "region": "us-east-1"},
            },
            wantError: false,
        },
        // More cases...
    }

    for name, tc := range tests {
        t.Run(name, func(t *testing.T) {
            // Setup mock DB client
            mockClient := &mocks.MockDBClient{
                ExecuteFunc: func(ctx context.Context, sql string) (*sql.Rows, error) {
                    // Return mock rows
                },
            }

            // Test execution
        })
    }
}

func TestExecute_BatchMode(t *testing.T) {
    // Test batch query execution
}

func TestExecute_InteractiveMode(t *testing.T) {
    // Test interactive execution
}

func TestExecute_QueryTimeout(t *testing.T) {
    // Test query timeout handling
}

func TestExecute_QueryError(t *testing.T) {
    // Test query errors (syntax, etc)
}

func TestExecute_ConnectionFailure(t *testing.T) {
    // Test database connection failures
}

func TestExecute_MultipleQueries(t *testing.T) {
    // Test executing multiple queries
}
```

### 2. init_data_test.go

```go
package query

import "testing"

func TestNewInitData_Success(t *testing.T) {
    // Test successful initialization
}

func TestNewInitData_NoConnection(t *testing.T) {
    // Test when DB connection fails
}

func TestNewInitData_InvalidConfig(t *testing.T) {
    // Test with invalid configuration
}

func TestNewInitData_WorkspaceLoad(t *testing.T) {
    // Test workspace loading
}
```

### 3. db_client_execute_test.go

```go
package db_client

import "testing"

func TestExecuteQuery_Success(t *testing.T) {
    // Test successful query execution
}

func TestExecuteQuery_WithRetry(t *testing.T) {
    // Test retry logic on temporary failures
}

func TestExecuteQuery_MaxRetries(t *testing.T) {
    // Test max retries exceeded
}

func TestExecuteQuery_TransactionHandling(t *testing.T) {
    // Test transaction begin/commit/rollback
}
```

### 4. query_command_test.go

```go
package cmd

import "testing"

// Note: Testing cobra commands requires special setup

func TestQueryCommand_ParseArgs(t *testing.T) {
    // Test argument parsing
}

func TestQueryCommand_OutputFormat(t *testing.T) {
    // Test different output formats (json, csv, table, line)
}

func TestQueryCommand_FileInput(t *testing.T) {
    // Test reading queries from file
}

func TestQueryCommand_StdinInput(t *testing.T) {
    // Test reading queries from stdin
}
```

## Critical Test Cases

### Query Execution
1. ✅ **Simple SELECT** - Basic query works
2. ✅ **Complex query** - WITH, JOIN, subqueries
3. ✅ **Query timeout** - Long-running query timeout
4. ✅ **Syntax error** - Invalid SQL handling
5. ✅ **Empty result** - No rows returned
6. ✅ **Large result set** - Many rows
7. ✅ **Multiple queries** - Batch execution
8. ✅ **Query cancellation** - Ctrl+C handling

### Output Formats
1. ✅ **Table format** - Default pretty table
2. ✅ **JSON format** - JSON output
3. ✅ **CSV format** - CSV output
4. ✅ **Line format** - Line-by-line output
5. ✅ **Output to file** - File export

### Error Handling
1. ✅ **Connection lost** - DB connection drops
2. ✅ **Plugin error** - Plugin returns error
3. ✅ **Invalid connection** - Connection doesn't exist
4. ✅ **Permission denied** - Access control
5. ✅ **Rate limit** - API rate limiting

### Modes
1. ✅ **Interactive mode** - Prompt-based
2. ✅ **Batch mode** - One-off queries
3. ✅ **File input** - Read from .sql file
4. ✅ **Stdin input** - Pipe queries

## Mocking Strategy

### What to Mock
- Database client (use mock from Task 1)
- SQL query results
- File system (for file input)
- Stdin/stdout (for interactive mode)
- Time (for timeout tests)

### What to Keep Real
- Query parsing logic
- Result formatting
- Error message generation
- Configuration structures

## Success Criteria

1. **Coverage Target**
   - 60%+ coverage on execute.go
   - 50%+ coverage on init_data.go
   - 60%+ coverage on db_client_execute.go
   - 40%+ coverage on cmd/query.go (cobra commands harder to test)

2. **Test Quality**
   - [ ] All critical paths tested
   - [ ] Multiple output formats tested
   - [ ] Error scenarios covered
   - [ ] Tests run fast (<50ms each)
   - [ ] No external dependencies

3. **All Tests Pass**
   ```bash
   go test -v ./pkg/query/...
   go test -v ./pkg/db/db_client/
   go test -v ./cmd/
   ```

4. **Existing Tests Still Pass**
   ```bash
   go test ./...
   cd tests/acceptance && ./run.sh chaos_and_query.bats
   ```

## Testing Your Work

```bash
# Run your new tests
go test -v ./pkg/query/queryexecute/
go test -v ./pkg/query/
go test -v ./pkg/db/db_client/ -run Execute

# Check coverage
go test -cover ./pkg/query/...

# Run query-related BATS tests
cd tests/acceptance
./run.sh chaos_and_query.bats

# Run all tests
go test ./...
```

## Dependencies

- **Requires:** Task 1 (test infrastructure) must be complete
- **Parallel with:** Tasks 2, 4-7

## Estimated Time
5-6 hours (complex area with many scenarios)

## Notes

- Query execution has MANY code paths (formats, modes, errors)
- Focus on core execution logic first, then edge cases
- Look at `tests/acceptance/test_files/chaos_and_query.bats` for test scenarios (348 lines, 38 tests!)
- Interactive mode testing is complex - may need special setup
- Consider using golden files for expected output comparison

## Golden Files Example

For testing output formats:
```
pkg/query/queryexecute/testdata/
├── simple_query.golden.json
├── simple_query.golden.csv
└── simple_query.golden.table
```

## Report Format

Update `.ai/milestones/wave-1-foundation/STATUS.md`:
```markdown
## Task 3: Query Execution Tests

**Status:** ✅ Complete
**Coverage Achieved:** X%
**Tests Added:** Y tests

### Files Created
- pkg/query/queryexecute/execute_test.go (X tests)
- pkg/query/init_data_test.go (Y tests)
- pkg/db/db_client/db_client_execute_test.go (Z tests)
- cmd/query_test.go (W tests)

### Test Results
- All new tests passing: ✅
- Existing tests passing: ✅
- Coverage: X%

### Issues Encountered
- [Any issues and how you resolved them]

### Notes
- [Any patterns or insights for other tasks]
```

## Command to Run

```bash
# Requires Task 1 complete first!
# In a new terminal:
claude

# Then:
# "Please complete task-3-query-tests.md from .ai/milestones/wave-1-foundation/tasks/"
```
