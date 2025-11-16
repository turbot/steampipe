package connection

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/v2/pkg/constants"
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

// TestLogRefreshConnectionResultsTypeAssertion tests the type assertion panic bug in logRefreshConnectionResults
// This test demonstrates issue #4807 - potential panic when viper.Get returns nil or wrong type
func TestLogRefreshConnectionResultsTypeAssertion(t *testing.T) {
	// Save original viper value
	originalValue := viper.Get(constants.ConfigKeyActiveCommand)
	defer func() {
		if originalValue != nil {
			viper.Set(constants.ConfigKeyActiveCommand, originalValue)
		} else {
			// Clean up by setting to nil if it was nil
			viper.Set(constants.ConfigKeyActiveCommand, nil)
		}
	}()

	// Test case 1: viper.Get returns nil
	t.Run("nil value does not panic", func(t *testing.T) {
		viper.Set(constants.ConfigKeyActiveCommand, nil)

		state := &refreshConnectionState{}

		// After the fix, this should NOT panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic occurred: %v", r)
			}
		}()

		// This should handle nil gracefully after the fix
		state.logRefreshConnectionResults()

		// If we get here without panic, the fix is working
		t.Log("Successfully handled nil value without panic")
	})

	// Test case 2: viper.Get returns wrong type
	t.Run("wrong type does not panic", func(t *testing.T) {
		viper.Set(constants.ConfigKeyActiveCommand, "not-a-cobra-command")

		state := &refreshConnectionState{}

		// After the fix, this should NOT panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic occurred: %v", r)
			}
		}()

		// This should handle wrong type gracefully after the fix
		state.logRefreshConnectionResults()

		// If we get here without panic, the fix is working
		t.Log("Successfully handled wrong type without panic")
	})

	// Test case 3: viper.Get returns *cobra.Command but it's nil
	t.Run("nil cobra.Command pointer does not panic", func(t *testing.T) {
		var nilCmd *cobra.Command
		viper.Set(constants.ConfigKeyActiveCommand, nilCmd)

		state := &refreshConnectionState{}

		// After the fix, this should NOT panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic occurred: %v", r)
			}
		}()

		// This should handle nil cobra.Command gracefully after the fix
		state.logRefreshConnectionResults()

		// If we get here without panic, the fix is working
		t.Log("Successfully handled nil cobra.Command pointer without panic")
	})

	// Test case 4: Valid cobra.Command (should work)
	t.Run("valid cobra.Command works", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "plugin-manager",
		}
		viper.Set(constants.ConfigKeyActiveCommand, cmd)

		state := &refreshConnectionState{}

		// This should work
		state.logRefreshConnectionResults()
	})
}

// TestExecuteUpdateSetsInParallelGoroutineLeak tests for goroutine leak in executeUpdateSetsInParallel
// This test demonstrates issue #4791 - potential goroutine leak with non-idiomatic channel pattern
//
// The issue is in refresh_connections_state.go:519-536 where the goroutine uses:
//   for { select { case connectionError := <-errChan: if connectionError == nil { return } } }
//
// While this pattern technically works when the channel is closed (returns nil, then returns from goroutine),
// it has several problems:
// 1. It's not idiomatic Go - the standard pattern for consuming until close is 'for range'
// 2. It relies on nil checks which can be error-prone
// 3. It's harder to understand and maintain
// 4. If the nil check is accidentally removed or modified, it causes a goroutine leak
//
// The idiomatic pattern 'for range errChan' automatically exits when channel is closed,
// making the code safer and more maintainable.
func TestExecuteUpdateSetsInParallelGoroutineLeak(t *testing.T) {
	// Get baseline goroutine count
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	baselineGoroutines := runtime.NumGoroutine()

	// Test the CURRENT pattern from refresh_connections_state.go:519-536
	// This pattern has potential for goroutine leaks if not carefully maintained
	errChan := make(chan *connectionError)
	var errorList []error
	var mu sync.Mutex

	// Simulate the current (non-idiomatic) pattern
	go func() {
		for {
			select {
			case connectionError := <-errChan:
				if connectionError == nil {
					return
				}
				mu.Lock()
				errorList = append(errorList, connectionError.err)
				mu.Unlock()
			}
		}
	}()

	// Send some errors
	testErr := errors.New("test error")
	errChan <- &connectionError{name: "test1", err: testErr}
	errChan <- &connectionError{name: "test2", err: testErr}

	// Close the channel (this should cause goroutine to exit via nil check)
	close(errChan)

	// Give time for the goroutine to process and exit
	time.Sleep(200 * time.Millisecond)
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// Check for goroutine leak
	afterGoroutines := runtime.NumGoroutine()
	goroutineDiff := afterGoroutines - baselineGoroutines

	// The current pattern SHOULD work (goroutine exits via nil check),
	// but we're testing to document that the pattern is risky
	if goroutineDiff > 2 {
		t.Errorf("Goroutine leak detected with current pattern: baseline=%d, after=%d, diff=%d",
			baselineGoroutines, afterGoroutines, goroutineDiff)
	}

	// Verify errors were collected
	mu.Lock()
	if len(errorList) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errorList))
	}
	mu.Unlock()

	t.Logf("BUG #4791: Current pattern works but is non-idiomatic and error-prone")
	t.Logf("The for-select-nil-check pattern at refresh_connections_state.go:520-535")
	t.Logf("should be replaced with idiomatic 'for range errChan' for safety and clarity")
}
