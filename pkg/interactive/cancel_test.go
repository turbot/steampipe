package interactive

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/goleak"
)

// TestCreatePromptContext tests prompt context creation
func TestCreatePromptContext(t *testing.T) {
	c := &InteractiveClient{}
	parentCtx := context.Background()

	ctx := c.createPromptContext(parentCtx)

	if ctx == nil {
		t.Fatal("createPromptContext returned nil context")
	}

	if c.cancelPrompt == nil {
		t.Fatal("createPromptContext didn't set cancelPrompt")
	}

	// Verify context can be cancelled
	c.cancelPrompt()

	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Context was not cancelled after calling cancelPrompt")
	}
}

// TestCreatePromptContextReplacesOld tests that creating a new context cancels the old one
func TestCreatePromptContextReplacesOld(t *testing.T) {
	c := &InteractiveClient{}
	parentCtx := context.Background()

	// Create first context
	ctx1 := c.createPromptContext(parentCtx)
	cancel1 := c.cancelPrompt

	// Create second context (should cancel first)
	ctx2 := c.createPromptContext(parentCtx)

	// First context should be cancelled
	select {
	case <-ctx1.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("First context was not cancelled when creating second context")
	}

	// Second context should still be active
	select {
	case <-ctx2.Done():
		t.Error("Second context should not be cancelled yet")
	case <-time.After(10 * time.Millisecond):
		// Expected
	}

	// First cancel function should be different from second
	if &cancel1 == &c.cancelPrompt {
		t.Error("cancelPrompt was not replaced")
	}
}

// TestCreateQueryContext tests query context creation
func TestCreateQueryContext(t *testing.T) {
	c := &InteractiveClient{}
	parentCtx := context.Background()

	ctx := c.createQueryContext(parentCtx)

	if ctx == nil {
		t.Fatal("createQueryContext returned nil context")
	}

	if c.cancelActiveQuery == nil {
		t.Fatal("createQueryContext didn't set cancelActiveQuery")
	}

	// Verify context can be cancelled
	c.cancelActiveQuery()

	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Context was not cancelled after calling cancelActiveQuery")
	}
}

// TestCreateQueryContextDoesNotCancelOld tests that creating a new query context doesn't cancel the old one
func TestCreateQueryContextDoesNotCancelOld(t *testing.T) {
	c := &InteractiveClient{}
	parentCtx := context.Background()

	// Create first context
	ctx1 := c.createQueryContext(parentCtx)
	cancel1 := c.cancelActiveQuery

	// Create second context (should NOT cancel first, just replace the reference)
	ctx2 := c.createQueryContext(parentCtx)

	// First context should still be active (not automatically cancelled)
	select {
	case <-ctx1.Done():
		t.Error("First context was cancelled when creating second context (should not auto-cancel)")
	case <-time.After(10 * time.Millisecond):
		// Expected - first context is NOT cancelled
	}

	// Cancel using the first cancel function
	cancel1()

	// Now first context should be cancelled
	select {
	case <-ctx1.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("First context was not cancelled after calling its cancel function")
	}

	// Second context should still be active
	select {
	case <-ctx2.Done():
		t.Error("Second context should not be cancelled yet")
	case <-time.After(10 * time.Millisecond):
		// Expected
	}
}

// TestCancelActiveQueryIfAnyIdempotent tests that cancellation is idempotent
func TestCancelActiveQueryIfAnyIdempotent(t *testing.T) {
	callCount := 0
	cancelFunc := func() {
		callCount++
	}

	c := &InteractiveClient{
		cancelActiveQuery: cancelFunc,
	}

	// Call multiple times
	for i := 0; i < 5; i++ {
		c.cancelActiveQueryIfAny()
	}

	// Should only be called once
	if callCount != 1 {
		t.Errorf("cancelActiveQueryIfAny() called cancel function %d times, want 1 (should be idempotent)", callCount)
	}

	// Should be nil after first call
	if c.cancelActiveQuery != nil {
		t.Error("cancelActiveQueryIfAny() didn't set cancelActiveQuery to nil")
	}
}

// TestCancelActiveQueryIfAnyNil tests behavior with nil cancel function
func TestCancelActiveQueryIfAnyNil(t *testing.T) {
	c := &InteractiveClient{
		cancelActiveQuery: nil,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("cancelActiveQueryIfAny() panicked with nil cancel function: %v", r)
		}
	}()

	// Should not panic
	c.cancelActiveQueryIfAny()

	// Should remain nil
	if c.cancelActiveQuery != nil {
		t.Error("cancelActiveQueryIfAny() set cancelActiveQuery when it was nil")
	}
}

// TestClosePrompt tests the ClosePrompt method
func TestClosePrompt(t *testing.T) {
	tests := []struct {
		name       string
		afterClose AfterPromptCloseAction
	}{
		{
			name:       "close with exit",
			afterClose: AfterPromptCloseExit,
		},
		{
			name:       "close with restart",
			afterClose: AfterPromptCloseRestart,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cancelled := false
			c := &InteractiveClient{
				cancelPrompt: func() {
					cancelled = true
				},
			}

			c.ClosePrompt(tt.afterClose)

			if !cancelled {
				t.Error("ClosePrompt didn't call cancelPrompt")
			}

			if c.afterClose != tt.afterClose {
				t.Errorf("ClosePrompt set afterClose to %v, want %v", c.afterClose, tt.afterClose)
			}
		})
	}
}

// TestClosePromptNilCancelPanic tests that ClosePrompt doesn't panic
// when cancelPrompt is nil.
//
// This can happen if ClosePrompt is called before the prompt is fully
// initialized or after manual nil assignment.
//
// Bug: #4788
func TestClosePromptNilCancelPanic(t *testing.T) {
	// Create an InteractiveClient with nil cancelPrompt
	c := &InteractiveClient{
		cancelPrompt: nil,
	}

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ClosePrompt() panicked with nil cancelPrompt: %v", r)
		}
	}()

	// Call ClosePrompt with nil cancelPrompt
	// This will panic without the fix
	c.ClosePrompt(AfterPromptCloseExit)
}

// TestContextCancellationPropagation tests that parent context cancellation propagates
func TestContextCancellationPropagation(t *testing.T) {
	c := &InteractiveClient{}
	parentCtx, parentCancel := context.WithCancel(context.Background())

	// Create child context
	childCtx := c.createPromptContext(parentCtx)

	// Cancel parent
	parentCancel()

	// Child should be cancelled too
	select {
	case <-childCtx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Child context was not cancelled when parent was cancelled")
	}
}

// TestContextCancellationTimeout tests context with timeout
func TestContextCancellationTimeout(t *testing.T) {
	c := &InteractiveClient{}
	parentCtx, parentCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer parentCancel()

	// Create child context
	childCtx := c.createPromptContext(parentCtx)

	// Wait for timeout
	select {
	case <-childCtx.Done():
		// Expected after ~50ms
		if childCtx.Err() != context.DeadlineExceeded && childCtx.Err() != context.Canceled {
			t.Errorf("Expected DeadlineExceeded or Canceled error, got %v", childCtx.Err())
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Context did not timeout as expected")
	}
}

// TestRapidContextCreation tests rapid context creation and cancellation
func TestRapidContextCreation(t *testing.T) {
	c := &InteractiveClient{}
	parentCtx := context.Background()

	// Rapidly create and cancel contexts
	for i := 0; i < 1000; i++ {
		ctx := c.createPromptContext(parentCtx)

		// Immediately cancel
		if c.cancelPrompt != nil {
			c.cancelPrompt()
		}

		// Verify cancellation
		select {
		case <-ctx.Done():
			// Expected
		case <-time.After(10 * time.Millisecond):
			t.Errorf("Context %d was not cancelled", i)
			return
		}
	}
}

// TestCancelAfterContextAlreadyCancelled tests cancelling after context is already cancelled
func TestCancelAfterContextAlreadyCancelled(t *testing.T) {
	c := &InteractiveClient{}
	parentCtx, parentCancel := context.WithCancel(context.Background())

	// Create child context
	ctx := c.createQueryContext(parentCtx)

	// Cancel parent first
	parentCancel()

	// Wait for child to be cancelled
	<-ctx.Done()

	// Now try to cancel via cancelActiveQueryIfAny
	// Should not panic even though context is already cancelled
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("cancelActiveQueryIfAny panicked when context already cancelled: %v", r)
		}
	}()

	c.cancelActiveQueryIfAny()
}

// TestContextCancellationTiming verifies that context cancellation propagates
// in a reasonable time across many iterations. This stress test helps identify
// timing issues or deadlocks in the cancellation logic.
func TestContextCancellationTiming(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timing stress test in short mode")
	}

	c := &InteractiveClient{}
	parentCtx := context.Background()

	// Create many query contexts
	for i := 0; i < 10000; i++ {
		ctx := c.createQueryContext(parentCtx)

		// Cancel immediately
		if c.cancelActiveQuery != nil {
			c.cancelActiveQuery()
		}

		// Verify context is cancelled within a reasonable timeout
		// Using 100ms to avoid flakiness on slower CI runners while still
		// catching real deadlocks or cancellation issues
		select {
		case <-ctx.Done():
			// Good - context was cancelled
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Context %d not cancelled within 100ms - possible deadlock or cancellation failure", i)
			return
		}
	}
}

// TestCancelFuncReplacement tests that cancel functions are properly replaced
func TestCancelFuncReplacement(t *testing.T) {
	c := &InteractiveClient{}
	parentCtx := context.Background()

	// Track which cancel function was called
	firstCalled := false
	secondCalled := false

	// Create first query context
	ctx1 := c.createQueryContext(parentCtx)
	firstCancel := c.cancelActiveQuery

	// Wrap the first cancel to track calls
	c.cancelActiveQuery = func() {
		firstCalled = true
		firstCancel()
	}

	// Create second query context (replaces cancelActiveQuery)
	ctx2 := c.createQueryContext(parentCtx)
	secondCancel := c.cancelActiveQuery

	// Wrap the second cancel to track calls
	c.cancelActiveQuery = func() {
		secondCalled = true
		secondCancel()
	}

	// Call cancelActiveQueryIfAny
	c.cancelActiveQueryIfAny()

	// Only the second cancel should be called
	if firstCalled {
		t.Error("First cancel function was called (should have been replaced)")
	}

	if !secondCalled {
		t.Error("Second cancel function was not called")
	}

	// Second context should be cancelled
	select {
	case <-ctx2.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Second context was not cancelled")
	}

	// First context is NOT automatically cancelled (different from prompt context)
	select {
	case <-ctx1.Done():
		// This might happen if parent was cancelled, but shouldn't happen from our cancel
	case <-time.After(10 * time.Millisecond):
		// Expected - first context remains active
	}
}

// TestNoGoroutineLeaks verifies that creating and cancelling query contexts
// doesn't leak goroutines. This uses goleak to detect goroutines that are
// still running after the test completes.
func TestNoGoroutineLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping goroutine leak test in short mode")
	}

	defer goleak.VerifyNone(t)

	c := &InteractiveClient{}
	parentCtx := context.Background()

	// Create and cancel many contexts to stress test for leaks
	for i := 0; i < 1000; i++ {
		ctx := c.createQueryContext(parentCtx)
		if c.cancelActiveQuery != nil {
			c.cancelActiveQuery()
			// Wait for cancellation to complete
			<-ctx.Done()
		}
	}
}

// TestConcurrentCancellation tests that cancelActiveQuery can be accessed
// concurrently without triggering data races.
// This test reproduces the race condition reported in issue #4802.
func TestConcurrentCancellation(t *testing.T) {
	// Create a minimal InteractiveClient
	client := &InteractiveClient{}

	// Simulate concurrent access to cancelActiveQuery from multiple goroutines
	// This mirrors real-world usage where:
	// - createQueryContext() sets cancelActiveQuery
	// - cancelActiveQueryIfAny() reads and clears it
	// - signal handlers may also call cancelActiveQueryIfAny()
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Simulate creating a query context (writes cancelActiveQuery)
			ctx := client.createQueryContext(context.Background())
			_ = ctx
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			// Simulate cancelling the active query (reads and writes cancelActiveQuery)
			client.cancelActiveQueryIfAny()
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// If we get here without panicking or race detector errors, the test passes
	// Note: This test will fail when run with -race flag if cancelActiveQuery access is not synchronized
}

// TestMultipleConcurrentCancellations tests rapid concurrent cancellations
// to stress test the synchronization.
func TestMultipleConcurrentCancellations(t *testing.T) {
	client := &InteractiveClient{}

	var wg sync.WaitGroup
	numIterations := 100

	// Create a query context first
	_ = client.createQueryContext(context.Background())

	// Now try to cancel it from multiple goroutines simultaneously
	for i := 0; i < numIterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.cancelActiveQueryIfAny()
		}()
	}

	wg.Wait()

	// Verify the client is in a consistent state
	if client.cancelActiveQuery != nil {
		t.Error("Expected cancelActiveQuery to be nil after all cancellations")
	}
}
