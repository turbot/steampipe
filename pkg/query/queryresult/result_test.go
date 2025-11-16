package queryresult

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/v2/queryresult"
)

func TestResultClose_DoubleClose(t *testing.T) {
	// Create a result with some column definitions
	cols := []*queryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
		{Name: "name", DataType: "text"},
	}
	result := NewResult(cols)

	// Close the result once
	result.Close()

	// Closing again should not panic (idempotent behavior)
	assert.NotPanics(t, func() {
		result.Close()
	}, "Result.Close() should be idempotent and not panic on second call")
}

// TestResult_ConcurrentReadAndClose tests concurrent read from RowChan and Close()
// This test demonstrates bug #4805 - race condition when reading while closing
func TestResult_ConcurrentReadAndClose(t *testing.T) {
	// Run the test multiple times to increase chance of catching race
	for i := 0; i < 100; i++ {
		cols := []*queryresult.ColumnDef{
			{Name: "id", DataType: "integer"},
		}
		result := NewResult(cols)

		var wg sync.WaitGroup
		wg.Add(3)

		// Goroutine 1: Stream rows
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				result.StreamRow([]interface{}{j})
			}
		}()

		// Goroutine 2: Read from RowChan (may race with Close)
		go func() {
			defer wg.Done()
			for range result.RowChan {
				// Consume rows - this read may race with channel close
			}
		}()

		// Goroutine 3: Close while reading is happening (triggers the race)
		go func() {
			defer wg.Done()
			time.Sleep(10 * time.Microsecond) // Let some rows stream first
			result.Close()                     // This may race with goroutine 2 reading
		}()

		wg.Wait()
	}
}

func TestWrapResult_NilResult(t *testing.T) {
	// WrapResult should handle nil input gracefully
	result := WrapResult(nil)

	// Result should be nil, not a wrapper around nil
	assert.Nil(t, result, "WrapResult(nil) should return nil")
}

// TestResult_CloseAfterPartialRead tests the race condition from issue #4790
// This test verifies that StreamRow() is safe to call concurrently with Close()
// When Close() closes the RowChan while StreamRow() is trying to send, it should not panic
func TestResult_CloseAfterPartialRead(t *testing.T) {
	// Run multiple iterations to increase chance of triggering the race
	for iteration := 0; iteration < 50; iteration++ {
		cols := []*queryresult.ColumnDef{
			{Name: "id", DataType: "integer"},
			{Name: "value", DataType: "text"},
		}
		result := NewResult(cols)

		var wg sync.WaitGroup
		wg.Add(2)

		// Goroutine 1: Stream many rows continuously
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				result.StreamRow([]interface{}{i, "value"})
				// No sleep - maximize chance of race
			}
		}()

		// Goroutine 2: Close after partial read
		go func() {
			defer wg.Done()
			// Read a few rows then close while streaming is still active
			count := 0
			for row := range result.RowChan {
				if row == nil {
					break
				}
				count++
				if count > 5 {
					// Close while StreamRow is still being called
					result.Close()
					// Continue draining the channel to prevent deadlock
					for range result.RowChan {
					}
					break
				}
			}
		}()

		wg.Wait()
	}
}
