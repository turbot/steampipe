package helpers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/steampipe/v2/pkg/test/helpers"
)

// TestExampleUsage demonstrates how to use the test infrastructure
func TestExampleUsage(t *testing.T) {
	// Example 1: Using filesystem helpers
	tempDir := helpers.CreateTempDir(t)
	testFile := helpers.WriteTestFile(t, tempDir, "test.txt", "test content")

	content := helpers.ReadTestFile(t, testFile)
	assert.Equal(t, "test content", content)

	// Example 2: Using config helpers
	config := helpers.NewTestConfig()
	assert.NotNil(t, config)
	assert.NotNil(t, config.Connections)

	// Example 3: Using testify assertions
	assert.NoError(t, nil)
	assert.Equal(t, "expected", "expected")
	assert.Contains(t, "hello world", "world")
	assert.True(t, true)
}

// TestMockDatabaseClient demonstrates how to use the mock database client
func TestMockDatabaseClient(t *testing.T) {
	// Create a test database client
	client := helpers.CreateTestDatabaseClient(t)
	assert.NotNil(t, client)

	// Acquire a session
	ctx := context.Background()
	result := client.AcquireSession(ctx)

	// Verify the session
	assert.NotNil(t, result)
	assert.NotNil(t, result.Session)
	assert.Equal(t, uint32(12345), result.Session.BackendPid)

	// Verify call tracking
	assert.Equal(t, 1, client.AcquireSessionCalls)
}

// TestFileSystemHelpers demonstrates file system helper usage
func TestFileSystemHelpers(t *testing.T) {
	// Create a temp directory
	tempDir := helpers.CreateTempDir(t)

	// Create a subdirectory
	subDir := helpers.CreateTestDir(t, tempDir, "subdir")
	assert.True(t, helpers.DirExists(subDir), "subdirectory should exist")

	// Write a test file
	filePath := helpers.WriteTestFile(t, subDir, "test.txt", "test content")
	assert.True(t, helpers.FileExists(filePath), "test file should exist")

	// Read the file back
	content := helpers.ReadTestFile(t, filePath)
	assert.Equal(t, "test content", content)
}

// TestConfigHelpers demonstrates config helper usage
func TestConfigHelpers(t *testing.T) {
	// Create a basic test config
	config := helpers.NewTestConfig()
	assert.NotNil(t, config)
	assert.Len(t, config.Connections, 0)

	// Create a connection
	conn := helpers.NewTestConnection("test_connection")
	assert.Equal(t, "test_connection", conn.Name)

	// Add the connection to the config
	helpers.AddConnectionToConfig(config, conn)
	assert.Len(t, config.Connections, 1)

	// Create a config with a connection in one step
	configWithConn := helpers.NewTestConfigWithConnection(t, "my_connection")
	assert.Len(t, configWithConn.Connections, 1)
}

// TestAssertionHelpers demonstrates testify assertion usage
func TestAssertionHelpers(t *testing.T) {
	// Error assertions
	assert.NoError(t, nil)
	// assert.Error(t, someError) // would fail if error is nil

	// Equality assertions
	assert.Equal(t, 42, 42)
	assert.NotEqual(t, 42, 43)

	// String assertions
	assert.Contains(t, "hello world", "world")
	assert.NotContains(t, "hello world", "foo")

	// Boolean assertions
	assert.True(t, true)
	assert.False(t, false)

	// Nil assertions
	assert.Nil(t, nil)
	assert.NotNil(t, "not nil")

	// Length assertions
	slice := []int{1, 2, 3}
	assert.Len(t, slice, 3)

	// Empty/NotEmpty
	assert.Empty(t, []int{})
	assert.NotEmpty(t, slice)

	// Greater/Less than
	assert.Greater(t, 10, 5)
	assert.Less(t, 5, 10)
}
