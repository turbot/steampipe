package queryresult

import (
	"testing"

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

func TestResult_ConcurrentStreamRowAndClose(t *testing.T) {
	// Demonstrates bug #4790 - Race condition between StreamRow() and Close()
	// When StreamRow() and Close() are called concurrently, we can get a
	// "send on closed channel" panic if Close() closes the RowChan while
	// StreamRow() is trying to send to it.

	cols := []*queryresult.ColumnDef{
		{Name: "id", DataType: "integer"},
		{Name: "name", DataType: "text"},
	}

	// Run this test multiple times to increase likelihood of triggering the race
	for i := 0; i < 100; i++ {
		result := NewResult(cols)

		// Start a goroutine that sends rows
		go func() {
			for j := 0; j < 10; j++ {
				result.StreamRow([]interface{}{j, "test"})
			}
		}()

		// Start a goroutine that consumes rows
		go func() {
			for range result.RowChan {
				// Just drain the channel
			}
		}()

		// Close immediately, creating a race with StreamRow
		result.Close()
	}

	// If we get here without panicking, the test passes
	// Run with -race flag to detect the race condition
}
