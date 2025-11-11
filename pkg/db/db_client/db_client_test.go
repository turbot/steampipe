package db_client

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionMapCleanupImplemented verifies that the session map memory leak is fixed
// Reference: https://github.com/turbot/steampipe/issues/3737
//
// This test verifies that a BeforeClose callback is registered to clean up
// session map entries when connections are dropped by pgx.
//
// Without this fix, sessions accumulate indefinitely causing a memory leak.
func TestSessionMapCleanupImplemented(t *testing.T) {
	// Read the db_client_connect.go file to verify BeforeClose callback exists
	content, err := os.ReadFile("db_client_connect.go")
	require.NoError(t, err, "should be able to read db_client_connect.go")

	sourceCode := string(content)

	// Verify BeforeClose callback is registered
	assert.Contains(t, sourceCode, "config.BeforeClose",
		"BeforeClose callback must be registered to clean up sessions when connections close")

	// Verify the callback deletes from sessions map
	assert.Contains(t, sourceCode, "delete(c.sessions, backendPid)",
		"BeforeClose callback must delete session entries to prevent memory leak")

	// Verify the comment in db_client.go documents automatic cleanup
	clientContent, err := os.ReadFile("db_client.go")
	require.NoError(t, err, "should be able to read db_client.go")

	clientCode := string(clientContent)

	// The comment should document automatic cleanup, not a TODO
	assert.NotContains(t, clientCode, "TODO: there's no code which cleans up this map",
		"TODO comment should be removed after implementing the fix")

	// Should document the automatic cleanup mechanism
	hasCleanupComment := strings.Contains(clientCode, "automatically cleaned up") ||
		strings.Contains(clientCode, "automatic cleanup") ||
		strings.Contains(clientCode, "BeforeClose")
	assert.True(t, hasCleanupComment,
		"Comment should document automatic cleanup mechanism")
}

// TestDbClient_DisableTimingFlag tests for race conditions on the disableTiming field
// Reference: https://github.com/turbot/steampipe/issues/4808
//
// This test demonstrates that the disableTiming boolean is accessed from multiple
// goroutines without synchronization, which can cause data races.
//
// The race occurs between:
// - shouldFetchTiming() reading disableTiming (db_client.go:138)
// - getQueryTiming() writing disableTiming (db_client_execute.go:190, 194)
func TestDbClient_DisableTimingFlag(t *testing.T) {
	// Read the db_client.go file to check the field type
	content, err := os.ReadFile("db_client.go")
	require.NoError(t, err, "should be able to read db_client.go")

	sourceCode := string(content)

	// Verify that disableTiming uses atomic.Bool instead of plain bool
	// The field declaration should be: disableTiming atomic.Bool
	assert.Contains(t, sourceCode, "disableTiming        atomic.Bool",
		"disableTiming must use atomic.Bool to prevent race conditions")

	// Verify the atomic import exists
	assert.Contains(t, sourceCode, "\"sync/atomic\"",
		"sync/atomic package must be imported for atomic.Bool")

	// Check that db_client_execute.go uses atomic operations
	executeContent, err := os.ReadFile("db_client_execute.go")
	require.NoError(t, err, "should be able to read db_client_execute.go")

	executeCode := string(executeContent)

	// Verify atomic Store operations are used instead of direct assignment
	assert.Contains(t, executeCode, ".Store(true)",
		"disableTiming writes must use atomic Store(true)")
	assert.Contains(t, executeCode, ".Store(false)",
		"disableTiming writes must use atomic Store(false)")

	// The old non-atomic assignments should not be present
	assert.NotContains(t, executeCode, "c.disableTiming = true",
		"direct assignment to disableTiming creates race condition")
	assert.NotContains(t, executeCode, "c.disableTiming = false",
		"direct assignment to disableTiming creates race condition")

	// Verify that shouldFetchTiming uses atomic Load
	shouldFetchTimingLine := "if c.disableTiming.Load() {"
	assert.Contains(t, sourceCode, shouldFetchTimingLine,
		"disableTiming reads must use atomic Load()")
}
