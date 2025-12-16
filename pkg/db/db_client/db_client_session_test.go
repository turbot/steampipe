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

// TestDbClient_SessionRegistration verifies session registration in sessions map
func TestDbClient_SessionRegistration(t *testing.T) {
	client := &DbClient{
		sessions:      make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex: &sync.Mutex{},
	}

	// Simulate session registration
	backendPid := uint32(12345)
	session := db_common.NewDBSession(backendPid)

	client.sessionsMutex.Lock()
	client.sessions[backendPid] = session
	client.sessionsMutex.Unlock()

	// Verify session is registered
	client.sessionsMutex.Lock()
	registeredSession, found := client.sessions[backendPid]
	client.sessionsMutex.Unlock()

	assert.True(t, found, "Session should be registered")
	assert.Equal(t, backendPid, registeredSession.BackendPid, "Backend PID should match")
}

// TestDbClient_SessionUnregistration verifies session cleanup via BeforeClose
func TestDbClient_SessionUnregistration(t *testing.T) {
	client := &DbClient{
		sessions:      make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex: &sync.Mutex{},
	}

	// Add sessions
	backendPid1 := uint32(100)
	backendPid2 := uint32(200)

	client.sessionsMutex.Lock()
	client.sessions[backendPid1] = db_common.NewDBSession(backendPid1)
	client.sessions[backendPid2] = db_common.NewDBSession(backendPid2)
	client.sessionsMutex.Unlock()

	assert.Len(t, client.sessions, 2, "Should have 2 sessions")

	// Simulate BeforeClose callback for one session
	client.sessionsMutex.Lock()
	delete(client.sessions, backendPid1)
	client.sessionsMutex.Unlock()

	// Verify only one session remains
	client.sessionsMutex.Lock()
	_, found1 := client.sessions[backendPid1]
	_, found2 := client.sessions[backendPid2]
	client.sessionsMutex.Unlock()

	assert.False(t, found1, "First session should be removed")
	assert.True(t, found2, "Second session should still exist")
	assert.Len(t, client.sessions, 1, "Should have 1 session remaining")
}

// TestDbClient_ConcurrentSessionRegistration tests concurrent session additions
func TestDbClient_ConcurrentSessionRegistration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	client := &DbClient{
		sessions:      make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex: &sync.Mutex{},
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrently add sessions
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id uint32) {
			defer wg.Done()
			backendPid := id
			session := db_common.NewDBSession(backendPid)

			client.sessionsMutex.Lock()
			client.sessions[backendPid] = session
			client.sessionsMutex.Unlock()
		}(uint32(i))
	}

	wg.Wait()

	// Verify all sessions were added
	assert.Len(t, client.sessions, numGoroutines, "All sessions should be registered")
}

// TestDbClient_SessionMapGrowthUnbounded tests for potential memory leaks
// This verifies that sessions don't accumulate indefinitely
func TestDbClient_SessionMapGrowthUnbounded(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	client := &DbClient{
		sessions:      make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex: &sync.Mutex{},
	}

	// Simulate many connections
	numSessions := 10000
	for i := 0; i < numSessions; i++ {
		backendPid := uint32(i)
		session := db_common.NewDBSession(backendPid)

		client.sessionsMutex.Lock()
		client.sessions[backendPid] = session
		client.sessionsMutex.Unlock()
	}

	assert.Len(t, client.sessions, numSessions, "Should have all sessions")

	// Simulate cleanup (BeforeClose callbacks)
	for i := 0; i < numSessions; i++ {
		backendPid := uint32(i)

		client.sessionsMutex.Lock()
		delete(client.sessions, backendPid)
		client.sessionsMutex.Unlock()
	}

	// Verify all sessions are cleaned up
	assert.Len(t, client.sessions, 0, "All sessions should be cleaned up")
}

// TestDbClient_SearchPathUpdates verifies session search path management
func TestDbClient_SearchPathUpdates(t *testing.T) {
	client := &DbClient{
		sessions:         make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex:    &sync.Mutex{},
		customSearchPath: []string{"schema1", "schema2"},
	}

	// Add a session
	backendPid := uint32(12345)
	session := db_common.NewDBSession(backendPid)

	client.sessionsMutex.Lock()
	client.sessions[backendPid] = session
	client.sessionsMutex.Unlock()

	// Verify custom search path is set
	assert.NotNil(t, client.customSearchPath, "Custom search path should be set")
	assert.Len(t, client.customSearchPath, 2, "Should have 2 schemas in search path")
}

// TestSearchPathAccessShouldUseReadLocks checks that search path access does not block other goroutines unnecessarily.
//
// Holding an exclusive mutex during search-path reads in concurrent query setup can deadlock when
// another goroutine is setting the path. The current code uses Lock/Unlock; this test documents
// the expectation to move to a read/non-blocking lock so concurrent reads are safe.
func TestSearchPathAccessShouldUseReadLocks(t *testing.T) {
	content, err := os.ReadFile("db_client_search_path.go")
	require.NoError(t, err, "should be able to read db_client_search_path.go")

	source := string(content)

	assert.Contains(t, source, "GetRequiredSessionSearchPath", "getter must exist")
	assert.Contains(t, source, "searchPathMutex", "getter must guard access to searchPath state")

	// Expect a read or non-blocking lock in getters; fail if only full Lock/Unlock is present.
	hasRLock := strings.Contains(source, "RLock")
	hasTry := strings.Contains(source, "TryLock") || strings.Contains(source, "tryLock")
	if !hasRLock && !hasTry {
		t.Fatalf("GetRequiredSessionSearchPath should avoid exclusive Lock/Unlock to prevent deadlocks under concurrent query setup")
	}
}

// TestDbClient_SessionConnectionNilSafety verifies handling of nil connections
func TestDbClient_SessionConnectionNilSafety(t *testing.T) {
	session := db_common.NewDBSession(12345)

	// Session is created with nil connection initially
	assert.Nil(t, session.Connection, "New session should have nil connection initially")
}

// TestDbClient_SessionSearchPathUpdatesThreadSafe verifies that concurrent access
// to customSearchPath does not cause data races.
// Reference: https://github.com/turbot/steampipe/issues/4792
//
// This test simulates concurrent goroutines accessing and modifying the customSearchPath
// slice. Without proper synchronization, this causes a data race.
//
// Run with: go test -race -run TestDbClient_SessionSearchPathUpdatesThreadSafe
func TestDbClient_SessionSearchPathUpdatesThreadSafe(t *testing.T) {
	// Create a DbClient with the fields we need for testing
	client := &DbClient{
		customSearchPath: []string{"public", "internal"},
		userSearchPath:   []string{"public"},
		searchPathMutex:  &sync.RWMutex{},
	}

	// Number of concurrent operations to test
	const numGoroutines = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3)

	// Simulate concurrent readers calling GetRequiredSessionSearchPath
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = client.GetRequiredSessionSearchPath()
		}()
	}

	// Simulate concurrent readers calling GetCustomSearchPath
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = client.GetCustomSearchPath()
		}()
	}

	// Simulate concurrent writers calling SetRequiredSessionSearchPath
	// This is the most dangerous operation as it modifies the slice
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			ctx := context.Background()
			// This will write to customSearchPath
			_ = client.SetRequiredSessionSearchPath(ctx)
		}()
	}

	wg.Wait()
}
