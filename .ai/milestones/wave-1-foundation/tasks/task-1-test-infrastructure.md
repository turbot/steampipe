# Task 1: Test Infrastructure Setup

## Context
You are setting up the testing infrastructure for the Steampipe project. This is the foundation that all other testing tasks will depend on. You're creating mocks, helpers, and utilities that will make writing tests easier and more consistent.

## Goal
Create reusable test infrastructure including mocks, helpers, and test utilities.

## Why This Matters
- Other agents depend on this infrastructure
- Consistency across all tests
- Makes writing tests faster
- Reduces code duplication

## Files to Create

### 1. Mock Database Client
**File:** `pkg/test/mocks/db_client.go`

Create a mock that implements the database client interface with configurable behavior:
```go
package mocks

import (
    "context"
    "database/sql"
)

// MockDBClient mocks the database client for testing
type MockDBClient struct {
    // Function fields for configurable behavior
    ExecuteFunc       func(ctx context.Context, sql string) (*sql.Rows, error)
    ConnectFunc       func() error
    CloseFunc         func() error
    AcquireSessionFunc func(ctx context.Context) error

    // Track calls
    ExecuteCalls    []string
    ConnectCalls    int
    CloseCalls      int
}

// Implement actual interface methods that call the Func fields
```

### 2. Mock Plugin Manager
**File:** `pkg/test/mocks/plugin_manager.go`

Create a mock plugin manager:
```go
package mocks

type MockPluginManager struct {
    GetPluginFunc     func(connectionName string) (*PluginInstance, error)
    StartFunc         func() error
    StopFunc          func() error
    RefreshConnectionsFunc func() error

    GetPluginCalls    []string
    StartCalls        int
    StopCalls         int
}
```

### 3. Database Test Helpers
**File:** `pkg/test/helpers/database.go`

```go
package helpers

import (
    "testing"
    "path/filepath"
    "os"
)

// CreateTempDir creates a temporary directory for testing
func CreateTempDir(t *testing.T) string {
    dir, err := os.MkdirTemp("", "steampipe-test-*")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }

    t.Cleanup(func() {
        os.RemoveAll(dir)
    })

    return dir
}

// CreateTestDatabase creates a test database (if needed)
// For now, may just return a mock or temp path
func CreateTestDatabase(t *testing.T) string {
    // Implementation
}
```

### 4. Config Test Helpers
**File:** `pkg/test/helpers/config.go`

```go
package helpers

import (
    "github.com/turbot/steampipe/pkg/steampipeconfig"
)

// NewTestConfig creates a minimal valid config for testing
func NewTestConfig() *steampipeconfig.SteampipeConfig {
    // Return a valid test config
}

// NewTestConnection creates a test connection
func NewTestConnection(name string) *steampipeconfig.Connection {
    // Return a valid test connection
}
```

### 5. Assertion Helpers
**File:** `pkg/test/helpers/assertions.go`

```go
package helpers

import (
    "testing"
    "reflect"
)

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }
}

// AssertError fails the test if err is nil
func AssertError(t *testing.T, err error, msgContains string) {
    t.Helper()
    if err == nil {
        t.Fatal("Expected error, got nil")
    }
    if msgContains != "" && !strings.Contains(err.Error(), msgContains) {
        t.Fatalf("Expected error containing '%s', got: %v", msgContains, err)
    }
}

// AssertEqual fails if values are not equal
func AssertEqual(t *testing.T, expected, actual interface{}) {
    t.Helper()
    if !reflect.DeepEqual(expected, actual) {
        t.Fatalf("Expected %v, got %v", expected, actual)
    }
}

// AssertContains fails if haystack doesn't contain needle
func AssertContains(t *testing.T, haystack, needle string) {
    t.Helper()
    if !strings.Contains(haystack, needle) {
        t.Fatalf("Expected '%s' to contain '%s'", haystack, needle)
    }
}
```

### 6. File System Helpers
**File:** `pkg/test/helpers/filesystem.go`

```go
package helpers

import (
    "os"
    "testing"
)

// WriteTestFile writes content to a file in a temp directory
func WriteTestFile(t *testing.T, dir, filename, content string) string {
    path := filepath.Join(dir, filename)
    err := os.WriteFile(path, []byte(content), 0644)
    if err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }
    return path
}

// CreateTestConfig creates a test .spc config file
func CreateTestConfigFile(t *testing.T, dir, name, content string) string {
    return WriteTestFile(t, dir, name+".spc", content)
}
```

### 7. Example Test
**File:** `pkg/test/helpers/example_test.go`

Create an example showing how to use the test infrastructure:
```go
package helpers_test

import (
    "testing"
    "github.com/turbot/steampipe/pkg/test/helpers"
)

func TestExampleUsage(t *testing.T) {
    // Example of using test helpers
    tempDir := helpers.CreateTempDir(t)
    config := helpers.NewTestConfig()

    // Use assertions
    helpers.AssertNoError(t, nil)
    helpers.AssertEqual(t, "expected", "expected")
}
```

## Success Criteria

1. **All Files Created**
   - [ ] `pkg/test/mocks/db_client.go`
   - [ ] `pkg/test/mocks/plugin_manager.go`
   - [ ] `pkg/test/helpers/database.go`
   - [ ] `pkg/test/helpers/config.go`
   - [ ] `pkg/test/helpers/assertions.go`
   - [ ] `pkg/test/helpers/filesystem.go`
   - [ ] `pkg/test/helpers/example_test.go`

2. **All Code Compiles**
   ```bash
   go build ./pkg/test/...
   ```

3. **Example Test Passes**
   ```bash
   go test ./pkg/test/helpers/
   ```

4. **Documentation Created**
   - [ ] Create `pkg/test/README.md` explaining how to use test infrastructure

## Testing Your Work

Run these commands to verify:
```bash
# Build test packages
go build ./pkg/test/...

# Run example test
go test ./pkg/test/helpers/

# Check imports
go list -f '{{.Imports}}' ./pkg/test/...
```

## Notes

- Look at existing Steampipe interfaces to understand what to mock
- Keep mocks simple - just enough to test
- Use `t.Helper()` in assertion functions
- Use `t.Cleanup()` for resource cleanup
- Follow existing Go testing conventions

## Dependencies

None - this is the foundation task.

## Estimated Time
2-3 hours

## Report Format

When complete, update STATUS.md with:
- Files created
- Tests passing
- Any issues encountered
- Example usage snippet

## Command to Run

```bash
# In your terminal, run:
claude

# Then paste this task file or say:
# "Please complete task-1-test-infrastructure.md from .ai/milestones/wave-1-foundation/tasks/"
```
