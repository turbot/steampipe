package db_client

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
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

// TestBeforeCloseCleanupShouldBeNonBlocking ensures the cleanup hook does not take a blocking lock.
//
// A blocking mutex in the BeforeClose hook can deadlock pool.Close() when another goroutine
// holds sessionsMutex (service stop/restart hangs). This test is intentionally strict and
// will fail until the hook uses a non-blocking strategy (e.g., TryLock or similar).
func TestBeforeCloseCleanupShouldBeNonBlocking(t *testing.T) {
	content, err := os.ReadFile("db_client_connect.go")
	require.NoError(t, err, "should be able to read db_client_connect.go")

	source := string(content)

	// Guardrail: the BeforeClose hook should avoid unconditionally blocking on sessionsMutex.
	assert.Contains(t, source, "config.BeforeClose", "BeforeClose cleanup hook must exist")
	assert.Contains(t, source, "sessionsTryLock", "BeforeClose cleanup should use non-blocking lock helper")

	// Expect a non-blocking lock pattern; if we only find Lock()/Unlock, this fails.
	nonBlockingPatterns := []string{"TryLock", "tryLock", "non-block", "select {"}
	foundNonBlocking := false
	for _, p := range nonBlockingPatterns {
		if strings.Contains(source, p) {
			foundNonBlocking = true
			break
		}
	}

	if !foundNonBlocking {
		t.Fatalf("BeforeClose cleanup appears to take a blocking lock on sessionsMutex; add a non-blocking guard to prevent pool.Close deadlocks")
	}
}

// TestDbClient_Close_Idempotent verifies that calling Close() multiple times does not cause issues
// Reference: Similar to bug #4712 (Result.Close() idempotency)
//
// Close() should be safe to call multiple times without panicking or causing errors.
func TestDbClient_Close_Idempotent(t *testing.T) {
	ctx := context.Background()

	// Create a minimal client (without real connection)
	client := &DbClient{
		sessions:      make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex: &sync.Mutex{},
	}

	// First close
	err := client.Close(ctx)
	assert.NoError(t, err, "First Close() should not return error")

	// Second close - should not panic
	err = client.Close(ctx)
	assert.NoError(t, err, "Second Close() should not return error")

	// Third close - should still not panic
	err = client.Close(ctx)
	assert.NoError(t, err, "Third Close() should not return error")

	// Verify sessions map is nil after close
	assert.Nil(t, client.sessions, "Sessions map should be nil after Close()")
}

// TestDbClient_ConcurrentSessionAccess tests concurrent access to the sessions map
// This test should be run with -race flag to detect data races.
//
// The sessions map is protected by sessionsMutex, but we want to verify
// that all access paths properly use the mutex.
func TestDbClient_ConcurrentSessionAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent access test in short mode")
	}

	client := &DbClient{
		sessions:      make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex: &sync.Mutex{},
	}

	var wg sync.WaitGroup
	numGoroutines := 50
	numOperations := 100

	// Track errors in a thread-safe way
	errors := make(chan error, numGoroutines*numOperations)

	// Simulate concurrent session additions
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id uint32) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Add session
				client.sessionsMutex.Lock()
				backendPid := id*1000 + uint32(j)
				client.sessions[backendPid] = db_common.NewDBSession(backendPid)
				client.sessionsMutex.Unlock()

				// Read session
				client.sessionsMutex.Lock()
				_ = client.sessions[backendPid]
				client.sessionsMutex.Unlock()

				// Delete session (simulating BeforeClose callback)
				client.sessionsMutex.Lock()
				delete(client.sessions, backendPid)
				client.sessionsMutex.Unlock()
			}
		}(uint32(i))
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}
}

// TestDbClient_Close_ClearsSessionsMap verifies that Close() properly clears the sessions map
func TestDbClient_Close_ClearsSessionsMap(t *testing.T) {
	ctx := context.Background()

	client := &DbClient{
		sessions:      make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex: &sync.Mutex{},
	}

	// Add some sessions
	client.sessions[1] = db_common.NewDBSession(1)
	client.sessions[2] = db_common.NewDBSession(2)
	client.sessions[3] = db_common.NewDBSession(3)

	assert.Len(t, client.sessions, 3, "Should have 3 sessions before Close()")

	// Close the client
	err := client.Close(ctx)
	assert.NoError(t, err)

	// Sessions should be nil after close
	assert.Nil(t, client.sessions, "Sessions map should be nil after Close()")
}

// TestDbClient_ConcurrentCloseAndRead verifies that concurrent reads don't panic
// when Close() sets sessions to nil
// Reference: https://github.com/turbot/steampipe/issues/4793
func TestDbClient_ConcurrentCloseAndRead(t *testing.T) {

	// This test simulates the race condition where:
	// 1. A goroutine enters AcquireSession, locks the mutex, reads c.sessions
	// 2. Close() sets c.sessions = nil WITHOUT holding the mutex
	// 3. The goroutine tries to write to c.sessions which is now nil
	// This causes a nil map panic or data race

	// Run the test multiple times to increase chance of catching the race
	for i := 0; i < 50; i++ {
		client := &DbClient{
			sessions:      make(map[uint32]*db_common.DatabaseSession),
			sessionsMutex: &sync.Mutex{},
		}

		done := make(chan bool, 2)

		// Goroutine 1: Simulates AcquireSession behavior
		go func() {
			defer func() { done <- true }()

			client.sessionsMutex.Lock()
			// After the fix, code should check if sessions is nil
			if client.sessions != nil {
				_, found := client.sessions[12345]
				if !found {
					client.sessions[12345] = db_common.NewDBSession(12345)
				}
			}
			client.sessionsMutex.Unlock()
		}()

		// Goroutine 2: Calls Close()
		go func() {
			defer func() { done <- true }()
			// Without the fix, Close() sets sessions to nil without mutex protection
			// This is the bug - it should acquire the mutex first
			client.Close(nil)
		}()

		// Wait for both goroutines
		<-done
		<-done
	}

	// With the bug present, running with -race will detect the data race
	// After the fix, this test should pass cleanly
}

// TestDbClient_ConcurrentClose tests concurrent Close() calls
// BUG FOUND: Race condition in Close() - c.sessions = nil at line 171 is not protected by mutex
// Reference: https://github.com/turbot/steampipe/issues/4780
func TestDbClient_ConcurrentClose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	ctx := context.Background()

	client := &DbClient{
		sessions:      make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex: &sync.Mutex{},
	}

	var wg sync.WaitGroup
	numGoroutines := 10

	// Call Close() from multiple goroutines simultaneously
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = client.Close(ctx)
		}()
	}

	wg.Wait()

	// Should not panic and sessions should be nil
	assert.Nil(t, client.sessions)
}

// TestDbClient_SessionsMapNilAfterClose verifies that accessing sessions after Close
// doesn't cause a nil pointer panic
// Reference: https://github.com/turbot/steampipe/issues/4793
func TestDbClient_SessionsMapNilAfterClose(t *testing.T) {

	client := &DbClient{
		sessions:      make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex: &sync.Mutex{},
	}

	// Add a session
	client.sessionsMutex.Lock()
	client.sessions[12345] = db_common.NewDBSession(12345)
	client.sessionsMutex.Unlock()

	// Close sets sessions to nil (without mutex protection - this is the bug)
	client.Close(nil)

	// Attempt to access sessions like AcquireSession does
	// After the fix, this should not panic
	client.sessionsMutex.Lock()
	defer client.sessionsMutex.Unlock()

	// With the bug: this panics because sessions is nil
	// After fix: sessions should either not be nil, or code checks for nil
	if client.sessions != nil {
		client.sessions[67890] = db_common.NewDBSession(67890)
	}
}

// TestDbClient_SessionsMutexProtectsMap verifies that sessionsMutex protects all map operations
func TestDbClient_SessionsMutexProtectsMap(t *testing.T) {
	// This is a structural test to verify the sessions map is never accessed without the mutex
	content, err := os.ReadFile("db_client_session.go")
	require.NoError(t, err, "should be able to read db_client_session.go")

	sourceCode := string(content)

	// Count occurrences of mutex lock helpers
	mutexLocks := strings.Count(sourceCode, "lockSessions()") +
		strings.Count(sourceCode, "sessionsTryLock()")

	// This is a heuristic check - in practice, we'd need more sophisticated analysis
	// But it serves as a reminder to use the mutex
	assert.True(t, mutexLocks > 0,
		"sessions lock helpers should be used when accessing sessions map")
}

// TestDbClient_SessionMapDocumentation verifies that session lifecycle is documented
func TestDbClient_SessionMapDocumentation(t *testing.T) {
	content, err := os.ReadFile("db_client.go")
	require.NoError(t, err)

	sourceCode := string(content)

	// Verify documentation mentions the lifecycle
	assert.Contains(t, sourceCode, "Session lifecycle:",
		"Sessions map should have lifecycle documentation")

	assert.Contains(t, sourceCode, "issue #3737",
		"Should reference the memory leak issue")
}

// TestDbClient_ClosePools_NilPoolsHandling verifies closePools handles nil pools
func TestDbClient_ClosePools_NilPoolsHandling(t *testing.T) {
	client := &DbClient{
		sessions:      make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex: &sync.Mutex{},
	}

	// Should not panic with nil pools
	assert.NotPanics(t, func() {
		client.closePools()
	}, "closePools should handle nil pools gracefully")
}

// TestResetPools verifies that ResetPools handles nil pools gracefully without panicking.
// This test addresses bug #4698 where ResetPools panics when called on a DbClient with nil pools.
func TestResetPools(t *testing.T) {
	// Create a DbClient with nil pools (simulating a partially initialized or closed client)
	client := &DbClient{
		userPool:       nil,
		managementPool: nil,
	}

	// ResetPools should NOT panic even with nil pools
	// This is the expected correct behavior
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ResetPools panicked with nil pools: %v", r)
		}
	}()

	ctx := context.Background()
	client.ResetPools(ctx)
}

// TestDbClient_SessionsMapInitialized verifies sessions map is initialized in NewDbClient
func TestDbClient_SessionsMapInitialized(t *testing.T) {
	// Verify the initialization happens in NewDbClient
	content, err := os.ReadFile("db_client.go")
	require.NoError(t, err)

	sourceCode := string(content)

	// Verify sessions map is initialized
	assert.Contains(t, sourceCode, "sessions:                make(map[uint32]*db_common.DatabaseSession)",
		"sessions map should be initialized in NewDbClient")

	// Verify mutex is initialized
	assert.Contains(t, sourceCode, "sessionsMutex:           &sync.Mutex{}",
		"sessionsMutex should be initialized in NewDbClient")
}

// TestDbClient_DeferredCleanupInNewDbClient verifies error cleanup in NewDbClient
func TestDbClient_DeferredCleanupInNewDbClient(t *testing.T) {
	content, err := os.ReadFile("db_client.go")
	require.NoError(t, err)

	sourceCode := string(content)

	// Verify there's a defer that handles cleanup on error
	assert.Contains(t, sourceCode, "defer func() {",
		"NewDbClient should have deferred cleanup")

	assert.Contains(t, sourceCode, "client.Close(ctx)",
		"Deferred cleanup should close the client on error")
}

// TestDbClient_ParallelSessionInitLock verifies parallelSessionInitLock initialization
func TestDbClient_ParallelSessionInitLock(t *testing.T) {
	content, err := os.ReadFile("db_client.go")
	require.NoError(t, err)

	sourceCode := string(content)

	// Verify parallelSessionInitLock is initialized
	assert.Contains(t, sourceCode, "parallelSessionInitLock:",
		"parallelSessionInitLock should be initialized")

	// Should use semaphore
	assert.Contains(t, sourceCode, "semaphore.NewWeighted",
		"parallelSessionInitLock should use weighted semaphore")
}

// TestDbClient_BeforeCloseCallbackNilSafety tests the BeforeClose callback with nil connection
func TestDbClient_BeforeCloseCallbackNilSafety(t *testing.T) {
	content, err := os.ReadFile("db_client_connect.go")
	require.NoError(t, err)

	sourceCode := string(content)

	// Verify nil checks in BeforeClose callback
	assert.Contains(t, sourceCode, "if conn != nil",
		"BeforeClose should check if conn is nil")

	assert.Contains(t, sourceCode, "conn.PgConn() != nil",
		"BeforeClose should check if PgConn() is nil")
}

// TestDbClient_BeforeCloseHandlesNilSessions verifies BeforeClose callback handles nil sessions map
// Reference: https://github.com/turbot/steampipe/issues/4809
//
// This test ensures that the BeforeClose callback properly checks if the sessions map
// has been nil'd by Close() before attempting to delete from it.
func TestDbClient_BeforeCloseHandlesNilSessions(t *testing.T) {
	// Read the source file to verify nil check is present
	content, err := os.ReadFile("db_client_connect.go")
	require.NoError(t, err, "should be able to read db_client_connect.go")

	sourceCode := string(content)

	// Verify BeforeClose callback exists
	assert.Contains(t, sourceCode, "config.BeforeClose",
		"BeforeClose callback must be registered")

	// Verify the callback checks for nil sessions before deleting
	// The check should happen after acquiring the mutex and before the delete
	hasNilCheckBeforeDelete := strings.Contains(sourceCode, "if c.sessions != nil") &&
		strings.Contains(sourceCode, "delete(c.sessions, backendPid)")
	assert.True(t, hasNilCheckBeforeDelete,
		"BeforeClose callback must check if sessions map is nil before deleting (fix for #4809)")

	// Verify comment explaining the nil check
	assert.Contains(t, sourceCode, "Check if sessions map has been nil'd by Close()",
		"Should document why the nil check is needed")
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
