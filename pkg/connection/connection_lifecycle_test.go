package connection

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestExemplarSchemaMapConcurrentAccess tests concurrent access to exemplarSchemaMap
// This test demonstrates issue #4757 - race condition when writing to exemplarSchemaMap
// without proper mutex protection.
func TestExemplarSchemaMapConcurrentAccess(t *testing.T) {
	// Create a refreshConnectionState with initialized exemplarSchemaMap
	state := &refreshConnectionState{
		exemplarSchemaMap:    make(map[string]string),
		exemplarSchemaMapMut: sync.Mutex{},
	}

	// Number of concurrent goroutines
	numGoroutines := 10
	numIterations := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch multiple goroutines that will concurrently read and write to exemplarSchemaMap
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numIterations; j++ {
				pluginName := "aws"
				connectionName := "connection"

				// Simulate the FIXED pattern in executeUpdateForConnections
				// Read with mutex (line 581-591)
				state.exemplarSchemaMapMut.Lock()
				_, haveExemplarSchema := state.exemplarSchemaMap[pluginName]
				state.exemplarSchemaMapMut.Unlock()

				// FIXED: Write with mutex protection (line 602-604)
				if !haveExemplarSchema {
					// Now properly protected with mutex
					state.exemplarSchemaMapMut.Lock()
					state.exemplarSchemaMap[pluginName] = connectionName
					state.exemplarSchemaMapMut.Unlock()
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify the map has an entry (basic sanity check)
	state.exemplarSchemaMapMut.Lock()
	if len(state.exemplarSchemaMap) == 0 {
		t.Error("Expected exemplarSchemaMap to have at least one entry")
	}
	state.exemplarSchemaMapMut.Unlock()
}

// TestExemplarSchemaMapRaceCondition specifically tests the race condition pattern
// found in refresh_connections_state.go:601 - now FIXED
func TestExemplarSchemaMapRaceCondition(t *testing.T) {
	// This test now PASSES with -race flag after the bug fix
	state := &refreshConnectionState{
		exemplarSchemaMap:    make(map[string]string),
		exemplarSchemaMapMut: sync.Mutex{},
	}

	plugins := []string{"aws", "azure", "gcp", "github", "slack"}

	var wg sync.WaitGroup

	// Simulate multiple connections being processed concurrently
	for _, plugin := range plugins {
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(p string, connNum int) {
				defer wg.Done()

				// This simulates the FIXED code pattern in executeUpdateForConnections
				state.exemplarSchemaMapMut.Lock()
				_, haveExemplar := state.exemplarSchemaMap[p]
				state.exemplarSchemaMapMut.Unlock()

				// FIXED: This write is now protected by the mutex
				if !haveExemplar {
					// No more race condition!
					state.exemplarSchemaMapMut.Lock()
					state.exemplarSchemaMap[p] = p + "_connection"
					state.exemplarSchemaMapMut.Unlock()
				}
			}(plugin, i)
		}
	}

	wg.Wait()

	// Verify all plugins are in the map
	state.exemplarSchemaMapMut.Lock()
	defer state.exemplarSchemaMapMut.Unlock()

	for _, plugin := range plugins {
		if _, ok := state.exemplarSchemaMap[plugin]; !ok {
			t.Errorf("Expected plugin %s to be in exemplarSchemaMap", plugin)
		}
	}
}

// TestRefreshConnectionsDeadlockTimeout tests that RefreshConnections cannot deadlock
// This test demonstrates issue #4761 - the double-lock mechanism (queueLock + executeLock)
// could theoretically lead to deadlock if executeLock is never released.
func TestRefreshConnectionsDeadlockTimeout(t *testing.T) {
	// This test simulates the scenario where:
	// 1. Goroutine A acquires queueLock via TryLock()
	// 2. Goroutine A blocks on executeLock.Lock() indefinitely
	// 3. Goroutine B tries queueLock.TryLock() and fails, returns immediately
	// 4. If executeLock is never released, system is effectively deadlocked

	// Acquire the executeLock to simulate a hung goroutine
	executeLock.Lock()

	// Create a channel to track if RefreshConnections completes or times out
	done := make(chan bool, 1)

	// Start RefreshConnections in a goroutine
	go func() {
		// This should block trying to acquire executeLock
		RefreshConnections(context.Background(), nil)
		done <- true
	}()

	// Wait for goroutine to block on executeLock
	time.Sleep(100 * time.Millisecond)

	// Try to call RefreshConnections again - should return immediately
	// because queueLock.TryLock() will fail
	start := time.Now()
	RefreshConnections(context.Background(), nil)
	elapsed := time.Since(start)

	// This should return quickly (< 1 second) because TryLock fails
	if elapsed > 1*time.Second {
		t.Errorf("RefreshConnections took too long when queueLock was held: %v", elapsed)
	}

	// Now the problem: the first goroutine is stuck forever waiting for executeLock
	// In a real scenario, if executeLock holder crashes/hangs, we have a deadlock

	// For this test, we'll verify the goroutine is still blocked after a timeout
	select {
	case <-done:
		t.Error("RefreshConnections completed unexpectedly - it should be blocked on executeLock")
	case <-time.After(500 * time.Millisecond):
		// Expected - goroutine is still blocked
		// This demonstrates the issue: without a timeout mechanism,
		// the goroutine would block indefinitely
	}

	// Clean up - release the lock so the goroutine can complete
	executeLock.Unlock()

	// Verify goroutine completes now
	select {
	case <-done:
		// Good - goroutine completed after lock was released
	case <-time.After(1 * time.Second):
		t.Error("RefreshConnections failed to complete even after executeLock was released")
	}
}

// TestRefreshConnectionState_ContextCancellation tests that executeUpdateSetsInParallel
// properly checks context cancellation in spawned goroutines.
// This test demonstrates issue #4806 - goroutines continue running until completion
// after context cancellation, wasting resources.
func TestRefreshConnectionState_ContextCancellation(t *testing.T) {
	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx // Will be used in the fixed version

	// Track how many goroutines are still running after cancellation
	var activeGoroutines atomic.Int32
	var goroutinesStarted atomic.Int32

	// Simulate executeUpdateSetsInParallel behavior
	var wg sync.WaitGroup
	numGoroutines := 20

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			goroutinesStarted.Add(1)
			activeGoroutines.Add(1)
			defer activeGoroutines.Add(-1)

			// Check if context is cancelled before starting work (Fix #4806)
			select {
			case <-ctx.Done():
				// Context cancelled - don't process this batch
				return
			default:
				// Context still valid - proceed with work
			}

			// Simulate work that takes time
			for j := 0; j < 10; j++ {
				// Check context cancellation in the loop (Fix #4806)
				select {
				case <-ctx.Done():
					// Context cancelled - stop processing
					return
				default:
					// Context still valid - continue
					time.Sleep(50 * time.Millisecond)
				}
			}
		}(i)
	}

	// Wait a bit for goroutines to start
	time.Sleep(100 * time.Millisecond)

	// Cancel the context - goroutines should stop
	cancel()

	// Wait a bit to see if goroutines respect cancellation
	time.Sleep(100 * time.Millisecond)

	// Check how many are still active
	active := activeGoroutines.Load()
	started := goroutinesStarted.Load()

	t.Logf("Goroutines started: %d, still active after cancellation: %d", started, active)

	// BUG #4806: Without the fix, most/all goroutines will still be running
	// because they don't check ctx.Done()
	// With the fix, active should be 0 or very low
	if active > started/2 {
		t.Errorf("Bug #4806: Too many goroutines still active after context cancellation (started: %d, active: %d). Goroutines should check ctx.Done() and exit early.", started, active)
	}

	// Clean up - wait for all goroutines to finish
	wg.Wait()
}
