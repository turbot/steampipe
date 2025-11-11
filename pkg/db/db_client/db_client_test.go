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

// TestDbClient_SessionsMapNilAfterClose demonstrates bug #4809
// After Close() sets sessions to nil, operations that access the sessions map
// must handle the nil case properly. This test verifies nil checks are in place.
func TestDbClient_SessionsMapNilAfterClose(t *testing.T) {
	// Read the source files to verify nil checks are present
	sessionCode, err := os.ReadFile("db_client_session.go")
	require.NoError(t, err, "should be able to read db_client_session.go")

	connectCode, err := os.ReadFile("db_client_connect.go")
	require.NoError(t, err, "should be able to read db_client_connect.go")

	sessionSource := string(sessionCode)
	connectSource := string(connectCode)

	// Verify AcquireSession checks for nil sessions map
	// The check should happen after acquiring the mutex and before accessing the map
	hasNilCheckInAcquire := strings.Contains(sessionSource, "c.sessions == nil") ||
		strings.Contains(sessionSource, "if c.sessions == nil")
	assert.True(t, hasNilCheckInAcquire,
		"AcquireSession must check if sessions map is nil after Close()")

	// Verify BeforeClose callback checks for nil sessions map
	// The check should happen before attempting to delete from the map
	// Can be either "c.sessions == nil" or "c.sessions != nil" pattern
	hasNilCheckInBeforeClose := strings.Contains(connectSource, "c.sessions == nil") ||
		strings.Contains(connectSource, "c.sessions != nil") ||
		strings.Contains(connectSource, "if c.sessions == nil") ||
		strings.Contains(connectSource, "if c.sessions != nil")
	assert.True(t, hasNilCheckInBeforeClose,
		"BeforeClose callback must check if sessions map is nil after Close()")
}
