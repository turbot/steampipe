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

func TestWrapResult_NilResult(t *testing.T) {
	// WrapResult should handle nil input gracefully
	result := WrapResult(nil)

	// Result should be nil, not a wrapper around nil
	assert.Nil(t, result, "WrapResult(nil) should return nil")
}
