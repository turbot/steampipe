package statushooks

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestSpinnerCancelChannelNeverInitialized tests that the cancel channel is never initialized
// BUG: The cancel channel field exists but is never initialized or used - it's dead code
func TestSpinnerCancelChannelNeverInitialized(t *testing.T) {
	spinner := NewStatusSpinnerHook()

	if spinner.cancel != nil {
		t.Error("BUG: Cancel channel should be nil (it's never initialized)")
	}

	// Even after showing and hiding, cancel is never used
	spinner.Show()
	spinner.Hide()

	// The cancel field exists but serves no purpose - this is dead code
	t.Log("CONFIRMED: Cancel channel field exists but is completely unused (dead code)")
}

// TestSpinnerConcurrentShowHide tests concurrent Show/Hide calls for race conditions
// BUG: This exposes a race condition on the 'visible' field
func TestSpinnerConcurrentShowHide(t *testing.T) {
	t.Skip("Demonstrates bugs #4743, #4744 - Race condition in concurrent Show/Hide. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
	spinner := NewStatusSpinnerHook()

	var wg sync.WaitGroup
	iterations := 100

	// Run with: go test -race
	for i := 0; i < iterations; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			spinner.Show() // BUG: Race on 'visible' field
		}()
		go func() {
			defer wg.Done()
			spinner.Hide() // BUG: Race on 'visible' field
		}()
	}

	wg.Wait()
	t.Log("Test completed - check for race detector warnings")
}

// TestSpinnerConcurrentUpdate tests concurrent message updates for race conditions
// BUG: This exposes a race condition on spinner.Suffix field
func TestSpinnerConcurrentUpdate(t *testing.T) {
	// t.Skip("Demonstrates bugs #4743, #4744 - Race condition in concurrent Update. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
	spinner := NewStatusSpinnerHook()
	spinner.Show()
	defer spinner.Hide()

	var wg sync.WaitGroup
	iterations := 100

	// Run with: go test -race
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			spinner.UpdateSpinnerMessage(fmt.Sprintf("msg-%d", n)) // BUG: Race on spinner.Suffix
		}(i)
	}

	wg.Wait()
	t.Log("Test completed - check for race detector warnings")
}

// TestSpinnerMessageDeferredRestart tests that Message() can restart a hidden spinner
// BUG: This exposes a bug where deferred Start() can restart a hidden spinner
func TestSpinnerMessageDeferredRestart(t *testing.T) {
	spinner := NewStatusSpinnerHook()
	spinner.UpdateSpinnerMessage("test message")
	spinner.Show()

	// Start a goroutine that will call Hide() while Message() is executing
	done := make(chan struct{})
	go func() {
		time.Sleep(10 * time.Millisecond)
		spinner.Hide()
		close(done)
	}()

	// Message() stops the spinner and defers Start()
	spinner.Message("test output")

	<-done
	time.Sleep(50 * time.Millisecond)

	// BUG: Spinner might be restarted even though Hide() was called
	if spinner.spinner.Active() {
		t.Error("BUG FOUND: Spinner was restarted after Hide() due to deferred Start() in Message()")
	}
}

// TestSpinnerWarnDeferredRestart tests that Warn() can restart a hidden spinner
// BUG: Similar to Message(), Warn() has the same deferred restart bug
func TestSpinnerWarnDeferredRestart(t *testing.T) {
	spinner := NewStatusSpinnerHook()
	spinner.UpdateSpinnerMessage("test message")
	spinner.Show()

	// Start a goroutine that will call Hide() while Warn() is executing
	done := make(chan struct{})
	go func() {
		time.Sleep(10 * time.Millisecond)
		spinner.Hide()
		close(done)
	}()

	// Warn() stops the spinner and defers Start()
	spinner.Warn("test warning")

	<-done
	time.Sleep(50 * time.Millisecond)

	// BUG: Spinner might be restarted even though Hide() was called
	if spinner.spinner.Active() {
		t.Error("BUG FOUND: Spinner was restarted after Hide() due to deferred Start() in Warn()")
	}
}

// TestSpinnerConcurrentMessageAndHide tests concurrent Message/Warn and Hide calls
// BUG: This exposes race conditions and the deferred restart bug
func TestSpinnerConcurrentMessageAndHide(t *testing.T) {
	t.Skip("Demonstrates bugs #4743, #4744 - Race condition in concurrent Message and Hide. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
	spinner := NewStatusSpinnerHook()
	spinner.UpdateSpinnerMessage("initial message")
	spinner.Show()

	var wg sync.WaitGroup
	iterations := 50

	// Run with: go test -race
	for i := 0; i < iterations; i++ {
		wg.Add(3)
		go func(n int) {
			defer wg.Done()
			spinner.Message(fmt.Sprintf("message-%d", n))
		}(i)
		go func(n int) {
			defer wg.Done()
			spinner.Warn(fmt.Sprintf("warning-%d", n))
		}(i)
		go func() {
			defer wg.Done()
			if i%10 == 0 {
				spinner.Hide()
			} else {
				spinner.Show()
			}
		}()
	}

	wg.Wait()
	t.Log("Test completed - check for race detector warnings and restart bugs")
}

// TestProgressReporterConcurrentUpdates tests concurrent updates to progress reporter
// This should be safe due to mutex, but we verify no races occur
func TestProgressReporterConcurrentUpdates(t *testing.T) {
	ctx := context.Background()
	ctx = AddStatusHooksToContext(ctx, NewStatusSpinnerHook())

	reporter := NewSnapshotProgressReporter("test-snapshot")

	var wg sync.WaitGroup
	iterations := 100

	// Run with: go test -race
	for i := 0; i < iterations; i++ {
		wg.Add(2)
		go func(n int) {
			defer wg.Done()
			reporter.UpdateRowCount(ctx, n)
		}(i)
		go func(n int) {
			defer wg.Done()
			reporter.UpdateErrorCount(ctx, 1)
		}(i)
	}

	wg.Wait()
	t.Logf("Final counts: rows=%d, errors=%d", reporter.rows, reporter.errors)
}

// TestSpinnerGoroutineLeak tests for goroutine leaks in spinner lifecycle
func TestSpinnerGoroutineLeak(t *testing.T) {
	// Allow some warm-up
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	initialGoroutines := runtime.NumGoroutine()

	// Create and destroy many spinners
	for i := 0; i < 100; i++ {
		spinner := NewStatusSpinnerHook()
		spinner.UpdateSpinnerMessage("test message")
		spinner.Show()
		time.Sleep(1 * time.Millisecond)
		spinner.Hide()
	}

	// Allow cleanup
	runtime.GC()
	time.Sleep(200 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()

	// Allow some tolerance (5 goroutines)
	if finalGoroutines > initialGoroutines+5 {
		t.Errorf("Possible goroutine leak: started with %d, ended with %d goroutines",
			initialGoroutines, finalGoroutines)
	}
}


// TestSpinnerUpdateAfterHide tests updating spinner message after Hide()
func TestSpinnerUpdateAfterHide(t *testing.T) {
	spinner := NewStatusSpinnerHook()
	spinner.Show()
	spinner.UpdateSpinnerMessage("initial message")
	spinner.Hide()

	// Update after hide - should not start spinner
	spinner.UpdateSpinnerMessage("updated message")

	if spinner.spinner.Active() {
		t.Error("Spinner should not be active after Hide() even if message is updated")
	}
}

// TestSpinnerSetStatusRace tests concurrent SetStatus calls
func TestSpinnerSetStatusRace(t *testing.T) {
	// t.Skip("Demonstrates bugs #4743, #4744 - Race condition in SetStatus. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
	spinner := NewStatusSpinnerHook()
	spinner.Show()

	var wg sync.WaitGroup
	iterations := 100

	// Run with: go test -race
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			spinner.SetStatus(fmt.Sprintf("status-%d", n))
		}(i)
	}

	wg.Wait()
	spinner.Hide()
}

// TestContextFunctionsNilContext tests that context helper functions handle nil context
func TestContextFunctionsNilContext(t *testing.T) {
	// These should not panic with nil context
	hooks := StatusHooksFromContext(nil)
	if hooks != NullHooks {
		t.Error("Expected NullHooks for nil context")
	}

	progress := SnapshotProgressFromContext(nil)
	if progress != NullProgress {
		t.Error("Expected NullProgress for nil context")
	}

	renderer := MessageRendererFromContext(nil)
	if renderer == nil {
		t.Error("Expected non-nil renderer for nil context")
	}
}

// TestSnapshotProgressHelperFunctions tests the helper functions for snapshot progress
func TestSnapshotProgressHelperFunctions(t *testing.T) {
	ctx := context.Background()
	reporter := NewSnapshotProgressReporter("test")
	ctx = AddSnapshotProgressToContext(ctx, reporter)

	// These should not panic
	UpdateSnapshotProgress(ctx, 10)
	SnapshotError(ctx)

	if reporter.rows != 10 {
		t.Errorf("Expected 10 rows, got %d", reporter.rows)
	}
	if reporter.errors != 1 {
		t.Errorf("Expected 1 error, got %d", reporter.errors)
	}
}

// TestSpinnerShowWithoutMessage tests showing spinner without setting a message first
func TestSpinnerShowWithoutMessage(t *testing.T) {
	spinner := NewStatusSpinnerHook()
	// Show without message - spinner should not start
	spinner.Show()

	if spinner.spinner.Active() {
		t.Error("Spinner should not be active when shown without a message")
	}
}

// TestSpinnerMultipleStartStopCycles tests multiple start/stop cycles
func TestSpinnerMultipleStartStopCycles(t *testing.T) {
	spinner := NewStatusSpinnerHook()
	spinner.UpdateSpinnerMessage("test message")

	for i := 0; i < 100; i++ {
		spinner.Show()
		time.Sleep(1 * time.Millisecond)
		spinner.Hide()
	}

	// Should not crash or leak resources
	t.Log("Multiple start/stop cycles completed successfully")
}

// TestSpinnerConcurrentSetStatusAndHide tests race between SetStatus and Hide
func TestSpinnerConcurrentSetStatusAndHide(t *testing.T) {
	// t.Skip("Demonstrates bugs #4743, #4744 - Race condition in concurrent SetStatus and Hide. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
	spinner := NewStatusSpinnerHook()
	spinner.Show()

	var wg sync.WaitGroup
	done := make(chan struct{})

	// Continuously set status
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				spinner.SetStatus("updating status")
			}
		}
	}()

	// Continuously hide/show
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			spinner.Hide()
			spinner.Show()
		}
	}()

	time.Sleep(100 * time.Millisecond)
	close(done)
	wg.Wait()
}
