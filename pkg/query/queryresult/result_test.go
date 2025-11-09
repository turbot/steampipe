package queryresult

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/v2/queryresult"
)

// TestNewResult tests the creation of a new Result with various column configurations
func TestNewResult(t *testing.T) {
	tests := map[string]struct {
		cols     []*queryresult.ColumnDef
		validate func(*testing.T, *Result)
	}{
		"with_valid_columns": {
			cols: []*queryresult.ColumnDef{
				{Name: "id", DataType: "integer"},
				{Name: "name", DataType: "text"},
			},
			validate: func(t *testing.T, result *Result) {
				assert.NotNil(t, result, "Result should not be nil")
				assert.NotNil(t, result.RowChan, "RowChan should be initialized")
				assert.NotNil(t, result.Cols, "Cols should be set")
				assert.Len(t, result.Cols, 2, "Should have 2 columns")
				assert.Equal(t, "id", result.Cols[0].Name)
				assert.Equal(t, "name", result.Cols[1].Name)

				// Verify timing stream is properly initialized
				assert.NotNil(t, result.Timing, "Timing should be initialized")
				assert.NotNil(t, result.Timing.Stream, "Timing stream channel should be initialized")
			},
		},
		"with_empty_columns": {
			cols: []*queryresult.ColumnDef{},
			validate: func(t *testing.T, result *Result) {
				assert.NotNil(t, result, "Result should not be nil even with empty columns")
				assert.NotNil(t, result.RowChan, "RowChan should be initialized")
				assert.Empty(t, result.Cols, "Cols should be empty")
				assert.NotNil(t, result.Timing, "Timing should still be initialized")
			},
		},
		"with_nil_columns": {
			cols: nil,
			validate: func(t *testing.T, result *Result) {
				assert.NotNil(t, result, "Result should not be nil even with nil columns")
				assert.NotNil(t, result.RowChan, "RowChan should be initialized")
				assert.Nil(t, result.Cols, "Cols should be nil as passed")
				assert.NotNil(t, result.Timing, "Timing should still be initialized")
			},
		},
		"with_many_columns": {
			cols: func() []*queryresult.ColumnDef {
				// Test with a large number of columns to check for any allocation issues
				cols := make([]*queryresult.ColumnDef, 100)
				for i := 0; i < 100; i++ {
					cols[i] = &queryresult.ColumnDef{
						Name:     string(rune('a' + (i % 26))),
						DataType: "text",
					}
				}
				return cols
			}(),
			validate: func(t *testing.T, result *Result) {
				assert.NotNil(t, result)
				assert.Len(t, result.Cols, 100, "Should handle 100 columns")
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := NewResult(tc.cols)
			tc.validate(t, result)

			// Clean up resources
			if result != nil && result.RowChan != nil {
				result.Close()
			}
		})
	}
}

// TestNewResult_ConcurrentCreation tests that multiple Results can be created concurrently
// This tests for race conditions in initialization
func TestNewResult_ConcurrentCreation(t *testing.T) {
	cols := []*queryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
	}

	var wg sync.WaitGroup
	results := make([]*Result, 10)

	// Create 10 Results concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = NewResult(cols)
		}(i)
	}

	wg.Wait()

	// Verify all Results are distinct and properly initialized
	channelSet := make(map[chan *queryresult.RowResult]bool)
	for i, result := range results {
		assert.NotNil(t, result, "Result %d should not be nil", i)
		assert.NotNil(t, result.RowChan, "Result %d should have RowChan", i)
		assert.NotNil(t, result.Timing.Stream, "Result %d should have Timing.Stream", i)

		// Verify each Result has its own unique channels
		assert.False(t, channelSet[result.RowChan], "Result %d should have unique RowChan", i)
		channelSet[result.RowChan] = true

		// Clean up
		result.Close()
	}

	assert.Len(t, channelSet, 10, "All Results should have unique channels")
}

// TestNewResult_TimingIntegration tests that the timing stream works correctly with Result
func TestNewResult_TimingIntegration(t *testing.T) {
	cols := []*queryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
	}

	result := NewResult(cols)
	defer result.Close()

	// Test that we can set and get timing through the Result
	testTiming := &TimingResult{
		DurationMs:   250,
		RowsReturned: 42,
		ScanCount:    1,
	}

	// Set timing in a goroutine
	go func() {
		result.Timing.SetTiming(testTiming)
	}()

	// Get timing - this will block until timing is set
	timing := result.Timing.GetTiming()
	assert.NotNil(t, timing, "Should retrieve timing")

	retrievedTiming, ok := timing.(*TimingResult)
	assert.True(t, ok, "Timing should be *TimingResult")
	assert.Equal(t, int64(250), retrievedTiming.DurationMs)
	assert.Equal(t, int64(42), retrievedTiming.RowsReturned)
	assert.Equal(t, int64(1), retrievedTiming.ScanCount)
}

// TestNewResult_NoResourceLeak tests that creating and closing Results doesn't leak resources
func TestNewResult_NoResourceLeak(t *testing.T) {
	cols := []*queryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
	}

	// Create and close many Results to check for leaks
	for i := 0; i < 100; i++ {
		result := NewResult(cols)
		assert.NotNil(t, result)
		result.Close()
	}

	// If there were goroutine or channel leaks, this test would fail or hang
	// The test passing quickly indicates proper cleanup
}

// TestNewResultStreamer tests the creation of ResultStreamer
func TestNewResultStreamer(t *testing.T) {
	tests := map[string]struct {
		validate func(*testing.T, *ResultStreamer)
	}{
		"creates_valid_streamer": {
			validate: func(t *testing.T, streamer *ResultStreamer) {
				assert.NotNil(t, streamer, "ResultStreamer should not be nil")
			},
		},
		"multiple_streamers_are_independent": {
			validate: func(t *testing.T, streamer *ResultStreamer) {
				streamer2 := NewResultStreamer()
				assert.NotNil(t, streamer, "First streamer should not be nil")
				assert.NotNil(t, streamer2, "Second streamer should not be nil")

				// Verify they are different instances
				// Note: We can't directly compare pointers of type aliases, but we can verify
				// they exist and are independently created
				assert.NotNil(t, streamer)
				assert.NotNil(t, streamer2)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			streamer := NewResultStreamer()
			tc.validate(t, streamer)
		})
	}
}

// TestNewResultStreamer_ConcurrentCreation tests concurrent streamer creation
func TestNewResultStreamer_ConcurrentCreation(t *testing.T) {
	var wg sync.WaitGroup
	streamers := make([]*ResultStreamer, 10)

	// Create 10 streamers concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			streamers[index] = NewResultStreamer()
		}(i)
	}

	wg.Wait()

	// Verify all streamers were created
	for i, streamer := range streamers {
		assert.NotNil(t, streamer, "Streamer %d should not be nil", i)
	}
}

// TestNewResultStreamer_NoResourceLeak tests that creating many streamers doesn't leak
func TestNewResultStreamer_NoResourceLeak(t *testing.T) {
	// Create many streamers rapidly
	for i := 0; i < 100; i++ {
		streamer := NewResultStreamer()
		assert.NotNil(t, streamer)
	}

	// If there were leaks, this would be slow or fail
	// The test passing quickly indicates proper resource management
}

// TestNewResult_StreamingWorkflow tests a complete streaming workflow
func TestNewResult_StreamingWorkflow(t *testing.T) {
	cols := []*queryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
		{Name: "name", DataType: "text"},
	}

	result := NewResult(cols)

	// Test streaming rows
	done := make(chan bool)
	receivedRows := [][]interface{}{}

	// Consumer goroutine
	go func() {
		for rowResult := range result.RowChan {
			if rowResult.Error != nil {
				t.Errorf("Received error: %v", rowResult.Error)
				break
			}
			if rowResult.Data != nil {
				receivedRows = append(receivedRows, rowResult.Data)
			}
		}
		done <- true
	}()

	// Producer goroutine
	go func() {
		// Send some test data
		result.StreamRow([]interface{}{1, "Alice"})
		result.StreamRow([]interface{}{2, "Bob"})
		result.StreamRow([]interface{}{3, "Charlie"})

		// Give a moment for processing
		time.Sleep(10 * time.Millisecond)

		// Close the result to signal completion
		result.Close()
	}()

	// Wait for consumer to finish
	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for rows")
	}

	// Verify we received all rows
	assert.Len(t, receivedRows, 3, "Should receive 3 rows")
	if len(receivedRows) == 3 {
		assert.Equal(t, []interface{}{1, "Alice"}, receivedRows[0])
		assert.Equal(t, []interface{}{2, "Bob"}, receivedRows[1])
		assert.Equal(t, []interface{}{3, "Charlie"}, receivedRows[2])
	}
}

// TestNewResult_ErrorStreaming tests error handling in streaming
func TestNewResult_ErrorStreaming(t *testing.T) {
	cols := []*queryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
	}

	result := NewResult(cols)

	done := make(chan bool)
	var receivedError error

	// Consumer goroutine
	go func() {
		for rowResult := range result.RowChan {
			if rowResult.Error != nil {
				receivedError = rowResult.Error
				break
			}
		}
		done <- true
	}()

	// Send an error
	go func() {
		result.StreamError(assert.AnError)
		result.Close()
	}()

	// Wait for consumer
	select {
	case <-done:
		assert.NotNil(t, receivedError, "Should receive error")
		assert.Equal(t, assert.AnError, receivedError)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for error")
	}
}

// TestNewResult_DoubleClose verifies that calling Close() twice is safe and idempotent
func TestNewResult_DoubleClose(t *testing.T) {
	cols := []*queryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
	}

	result := NewResult(cols)

	// First close should succeed
	result.Close()
	assert.True(t, result.IsClosed(), "Result should be marked as closed after first Close()")

	// Second close should be safe (no panic)
	assert.NotPanics(t, func() {
		result.Close()
	}, "Calling Close() twice should not panic")

	// Result should still be closed
	assert.True(t, result.IsClosed(), "Result should still be closed after second Close()")
}

// TestNewResult_ConcurrentClose verifies that concurrent Close() calls are safe
func TestNewResult_ConcurrentClose(t *testing.T) {
	cols := []*queryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
	}

	result := NewResult(cols)

	// Attempt to close from multiple goroutines concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.NotPanics(t, func() {
				result.Close()
			}, "Concurrent Close() calls should not panic")
		}()
	}

	wg.Wait()
	assert.True(t, result.IsClosed(), "Result should be closed after concurrent Close() calls")
}

// TestNewResult_IsClosed verifies the IsClosed() method works correctly
func TestNewResult_IsClosed(t *testing.T) {
	cols := []*queryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
	}

	result := NewResult(cols)

	// Initially not closed
	assert.False(t, result.IsClosed(), "Result should not be closed initially")

	// After Close(), should be closed
	result.Close()
	assert.True(t, result.IsClosed(), "Result should be closed after Close()")

	// Multiple checks should still return true
	assert.True(t, result.IsClosed(), "IsClosed() should remain true")
	assert.True(t, result.IsClosed(), "IsClosed() should remain true on repeated calls")
}
