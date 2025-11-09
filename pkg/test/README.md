# Steampipe Test Infrastructure

This directory contains the test infrastructure for Steampipe, including mocks, helpers, and utilities that make writing tests easier and more consistent.

## Directory Structure

```
pkg/test/
├── README.md                    # This file
├── mocks/                       # Mock implementations
│   ├── db_client.go            # Mock database client
│   └── plugin_manager.go       # Mock plugin manager
└── helpers/                     # Test helpers and utilities
    ├── config.go               # Config creation helpers
    ├── database.go             # Database test helpers
    ├── filesystem.go           # File system helpers
    └── example_test.go         # Example usage tests
```

## Testing Library

We use **[testify/assert](https://github.com/stretchr/testify)** for assertions. Testify is the industry-standard Go testing library with excellent assertion helpers.

### Common Assertions

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestExample(t *testing.T) {
    // Error assertions
    assert.NoError(t, err)
    assert.Error(t, err)
    assert.ErrorContains(t, err, "expected message")

    // Equality assertions
    assert.Equal(t, expected, actual)
    assert.NotEqual(t, value1, value2)
    assert.EqualValues(t, 42, int64(42))  // Type-insensitive comparison

    // String assertions
    assert.Contains(t, "hello world", "world")
    assert.NotContains(t, "hello world", "foo")

    // Boolean assertions
    assert.True(t, condition)
    assert.False(t, condition)

    // Nil assertions
    assert.Nil(t, value)
    assert.NotNil(t, value)

    // Length and emptiness
    assert.Len(t, slice, 3)
    assert.Empty(t, []int{})
    assert.NotEmpty(t, slice)

    // Numeric comparisons
    assert.Greater(t, 10, 5)
    assert.Less(t, 5, 10)
    assert.GreaterOrEqual(t, 10, 10)
}
```

For a complete list of assertions, see the [testify documentation](https://pkg.go.dev/github.com/stretchr/testify/assert).

## Mocks

### MockClient (Database Client)

A mock implementation of `db_common.Client` for testing database operations.

**Usage:**
```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/turbot/steampipe/v2/pkg/test/helpers"
)

func TestSomething(t *testing.T) {
    client := helpers.CreateTestDatabaseClient(t)

    // Customize behavior if needed
    client.ExecuteSyncFunc = func(ctx context.Context, sql string, args ...any) (*pqueryresult.SyncQueryResult, error) {
        // Custom behavior
        return &pqueryresult.SyncQueryResult{}, nil
    }

    // Use the client
    result := client.AcquireSession(context.Background())

    // Verify calls using testify
    assert.Equal(t, 1, client.AcquireSessionCalls)
}
```

**Key Features:**
- Tracks all method calls
- Configurable behavior via function fields
- Sensible defaults for common operations
- Thread-safe for concurrent tests

### MockPluginManager

A mock implementation of the plugin manager interface.

**Usage:**
```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/turbot/steampipe/v2/pkg/test/mocks"
)

func TestPlugin(t *testing.T) {
    pm := &mocks.MockPluginManager{
        GetFunc: func(req *pb.GetRequest) (*pb.GetResponse, error) {
            return &pb.GetResponse{}, nil
        },
    }

    // Use the plugin manager
    resp, err := pm.Get(&pb.GetRequest{Connections: []string{"test"}})

    // Verify calls
    assert.NoError(t, err)
    assert.Equal(t, 1, len(pm.GetCalls))
}
```

## Helpers

### Config Helpers

Create test configurations easily.

**Usage:**
```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/turbot/steampipe/v2/pkg/test/helpers"
)

// Create a basic config
config := helpers.NewTestConfig()

// Create a config with a connection
config := helpers.NewTestConfigWithConnection(t, "my_connection")

// Create a connection
conn := helpers.NewTestConnection("my_connection")

// Create an aggregator connection
aggregator := helpers.NewTestAggregatorConnection("all", []string{"conn1", "conn2"})

// Add connections to existing config
helpers.AddConnectionToConfig(config, conn)
```

### Database Helpers

Helpers for database-related testing.

**Usage:**
```go
import "github.com/turbot/steampipe/v2/pkg/test/helpers"

// Create a mock database client
client := helpers.CreateTestDatabaseClient(t)

// Create a test database session
session := helpers.CreateTestDatabaseSession(t)

// Create a test database path with directories
dbPath := helpers.CreateTestDatabasePath(t)
```

### Filesystem Helpers

Manage temporary files and directories in tests.

**Usage:**
```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/turbot/steampipe/v2/pkg/test/helpers"
)

// Create a temporary directory (auto-cleanup)
tempDir := helpers.CreateTempDir(t)

// Create a subdirectory
subDir := helpers.CreateTestDir(t, tempDir, "subdir")

// Write a test file
filePath := helpers.WriteTestFile(t, tempDir, "test.txt", "content")

// Write a config file
configPath := helpers.CreateTestConfigFile(t, tempDir, "myconfig", "config content")

// Read a test file
content := helpers.ReadTestFile(t, filePath)

// Check existence
if helpers.FileExists(filePath) {
    // File exists
}

if helpers.DirExists(tempDir) {
    // Directory exists
}
```

## Best Practices

### 1. Use t.Helper()

All helper functions call `t.Helper()` to ensure error messages point to the correct line:

```go
func MyHelper(t *testing.T) {
    t.Helper()
    // ... helper code
}
```

### 2. Use t.Cleanup()

Use cleanup functions for resource management:

```go
func TestSomething(t *testing.T) {
    resource := createResource()
    t.Cleanup(func() {
        resource.Close()
    })
}
```

### 3. Mock Only What You Need

Don't configure every mock function - only set the ones your test needs:

```go
client := helpers.CreateTestDatabaseClient(t)
// Only customize the method you need to test
client.ExecuteSyncFunc = func(ctx context.Context, sql string, args ...any) (*pqueryresult.SyncQueryResult, error) {
    return customResult, nil
}
```

### 4. Verify Behavior

Use call tracking to verify your code calls the expected methods:

```go
client := helpers.CreateTestDatabaseClient(t)
// ... run code that uses client
assert.Equal(t, 2, len(client.ExecuteSyncCalls))
assert.Equal(t, "SELECT * FROM test", client.ExecuteSyncCalls[0].SQL)
```

### 5. Table-Driven Tests

Use table-driven tests with testify for comprehensive coverage:

```go
func TestMultipleScenarios(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case1", "input1", "output1"},
        {"case2", "input2", "output2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Process(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## Examples

See `pkg/test/helpers/example_test.go` for complete working examples of all test infrastructure features.

## Running Tests

```bash
# Run all helper tests
go test ./pkg/test/helpers/

# Run with verbose output
go test -v ./pkg/test/helpers/

# Run specific test
go test -v ./pkg/test/helpers/ -run TestExampleUsage

# Build test packages
go build ./pkg/test/...
```

## Contributing

When adding new test infrastructure:

1. Add mocks in `pkg/test/mocks/`
2. Add helpers in `pkg/test/helpers/`
3. Document in this README
4. Add examples to `example_test.go`
5. Use testify/assert for all assertions
6. Ensure all code uses `t.Helper()` appropriately
7. Use `t.Cleanup()` for resource management

## Testing Conventions

- Test files end with `_test.go`
- Test functions start with `Test`
- Use descriptive test names: `TestConnectionValidation_WithInvalidConfig`
- Group related tests with subtests using `t.Run()`
- Use table-driven tests for multiple scenarios
- Always use testify/assert for assertions
- Import testify as: `"github.com/stretchr/testify/assert"`
