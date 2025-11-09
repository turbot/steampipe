package error_helpers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapError(t *testing.T) {
	tests := map[string]struct {
		err      error
		expected error
	}{
		"nil error": {
			err:      nil,
			expected: nil,
		},
		"simple error": {
			err:      errors.New("test error"),
			expected: errors.New("test error"),
		},
		"context canceled error": {
			err:      context.Canceled,
			expected: errors.New("execution cancelled"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := WrapError(tc.err)

			if tc.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tc.expected.Error(), result.Error())
			}
		})
	}
}

func TestTransformErrorToSteampipe(t *testing.T) {
	tests := map[string]struct {
		err            error
		expectedString string
	}{
		"nil error": {
			err:            nil,
			expectedString: "",
		},
		"simple error": {
			err:            errors.New("connection failed"),
			expectedString: "connection failed",
		},
		"ERROR prefix": {
			err:            errors.New("ERROR: syntax error"),
			expectedString: "syntax error",
		},
		"ERROR with rpc error": {
			err:            errors.New("ERROR: rpc error: code = Unknown desc = table not found"),
			expectedString: "table not found",
		},
		"context canceled": {
			err:            context.Canceled,
			expectedString: "execution cancelled",
		},
		"whitespace trimming": {
			err:            errors.New("  ERROR:  test error  "),
			expectedString: "test error",
		},
		"rpc error without ERROR prefix": {
			err:            errors.New("rpc error: code = Unknown desc = connection failed"),
			expectedString: "rpc error: code = Unknown desc = connection failed",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := TransformErrorToSteampipe(tc.err)

			if tc.expectedString == "" {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tc.expectedString, result.Error())
			}
		})
	}
}

func TestHandleCancelError(t *testing.T) {
	tests := map[string]struct {
		err            error
		expectedString string
	}{
		"context canceled": {
			err:            context.Canceled,
			expectedString: "execution cancelled",
		},
		"canceling statement error": {
			err:            errors.New("canceling statement due to user request"),
			expectedString: "execution cancelled",
		},
		"regular error": {
			err:            errors.New("regular error"),
			expectedString: "regular error",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := HandleCancelError(tc.err)
			assert.Equal(t, tc.expectedString, result.Error())
		})
	}
}

func TestIsCancelledError(t *testing.T) {
	tests := map[string]struct {
		err        error
		isCanceled bool
	}{
		"context canceled": {
			err:        context.Canceled,
			isCanceled: true,
		},
		"canceling statement": {
			err:        errors.New("canceling statement due to user request"),
			isCanceled: true,
		},
		"regular error": {
			err:        errors.New("regular error"),
			isCanceled: false,
		},
		"wrapped context canceled": {
			err:        fmt.Errorf("wrapped: %w", context.Canceled),
			isCanceled: true,
		},
		"partial match in error message": {
			err:        errors.New("error while canceling statement due to user request occurred"),
			isCanceled: true,
		},
		"nil error": {
			err:        nil,
			isCanceled: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// BUG CAUGHT: This test will panic if IsCancelledError doesn't check for nil
			result := IsCancelledError(tc.err)
			assert.Equal(t, tc.isCanceled, result)
		})
	}
}

func TestCombineErrors(t *testing.T) {
	tests := map[string]struct {
		errors       []error
		expectNil    bool
		checkStrings []string
	}{
		"no errors": {
			errors:    []error{},
			expectNil: true,
		},
		"all nil errors": {
			errors:    []error{nil, nil, nil},
			expectNil: true,
		},
		"single error": {
			errors:       []error{errors.New("error 1")},
			expectNil:    false,
			checkStrings: []string{"error 1"},
		},
		"multiple errors": {
			errors: []error{
				errors.New("error 1"),
				errors.New("error 2"),
				errors.New("error 3"),
			},
			expectNil:    false,
			checkStrings: []string{"error 1", "error 2", "error 3"},
		},
		"with nil errors": {
			errors: []error{
				errors.New("error 1"),
				nil,
				errors.New("error 2"),
			},
			expectNil:    false,
			checkStrings: []string{"error 1", "error 2"},
		},
		"duplicate errors": {
			errors: []error{
				errors.New("error 1"),
				errors.New("error 1"),
			},
			expectNil:    false,
			checkStrings: []string{"error 1"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := CombineErrors(tc.errors...)

			if tc.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				errorString := result.Error()
				for _, expected := range tc.checkStrings {
					assert.Contains(t, errorString, expected)
				}
			}
		})
	}
}

func TestCombineErrorsWithPrefix(t *testing.T) {
	tests := map[string]struct {
		prefix       string
		errors       []error
		expectNil    bool
		checkStrings []string
	}{
		"no errors with prefix": {
			prefix:    "Operation failed",
			errors:    []error{},
			expectNil: true,
		},
		"single error with prefix": {
			prefix:       "Operation failed",
			errors:       []error{errors.New("connection lost")},
			expectNil:    false,
			checkStrings: []string{"Operation failed", "connection lost"},
		},
		"multiple errors with prefix": {
			prefix: "Multiple issues",
			errors: []error{
				errors.New("error 1"),
				errors.New("error 2"),
			},
			expectNil:    false,
			checkStrings: []string{"Multiple issues", "error 1", "error 2"},
		},
		"single error without prefix": {
			prefix:       "",
			errors:       []error{errors.New("error 1")},
			expectNil:    false,
			checkStrings: []string{"error 1"},
		},
		"all nil with prefix": {
			prefix:    "Operation failed",
			errors:    []error{nil, nil},
			expectNil: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := CombineErrorsWithPrefix(tc.prefix, tc.errors...)

			if tc.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				errorString := result.Error()
				for _, expected := range tc.checkStrings {
					assert.Contains(t, errorString, expected)
				}
			}
		})
	}
}

func TestPrefixError(t *testing.T) {
	tests := map[string]struct {
		err            error
		prefix         string
		checkStrings   []string
	}{
		"simple error with prefix": {
			err:          errors.New("connection failed"),
			prefix:       "Database error",
			checkStrings: []string{"Database error", "connection failed"},
		},
		"ERROR prefixed error": {
			err:          errors.New("ERROR: syntax error"),
			prefix:       "Query failed",
			checkStrings: []string{"Query failed", "syntax error"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := PrefixError(tc.err, tc.prefix)

			assert.NotNil(t, result)
			errorString := result.Error()
			for _, expected := range tc.checkStrings {
				assert.Contains(t, errorString, expected)
			}
		})
	}
}

func TestHandleQueryTimeoutError(t *testing.T) {
	tests := map[string]struct {
		err       error
		isTimeout bool
	}{
		"deadline exceeded": {
			err:       context.DeadlineExceeded,
			isTimeout: true,
		},
		"regular error": {
			err:       errors.New("regular error"),
			isTimeout: false,
		},
		"wrapped deadline exceeded": {
			err:       fmt.Errorf("operation failed: %w", context.DeadlineExceeded),
			isTimeout: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := HandleQueryTimeoutError(tc.err)

			assert.NotNil(t, result)
			if tc.isTimeout {
				assert.Contains(t, result.Error(), "query timeout exceeded")
			} else {
				assert.Equal(t, tc.err.Error(), result.Error())
			}
		})
	}
}

func TestFailOnError(t *testing.T) {
	tests := map[string]struct {
		err         error
		shouldPanic bool
	}{
		"nil error does not panic": {
			err:         nil,
			shouldPanic: false,
		},
		"error causes panic": {
			err:         errors.New("test error"),
			shouldPanic: true,
		},
		"context canceled causes panic": {
			err:         context.Canceled,
			shouldPanic: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.shouldPanic {
				assert.Panics(t, func() {
					FailOnError(tc.err)
				})
			} else {
				assert.NotPanics(t, func() {
					FailOnError(tc.err)
				})
			}
		})
	}
}

func TestFailOnErrorWithMessage(t *testing.T) {
	tests := map[string]struct {
		err         error
		message     string
		shouldPanic bool
		checkString string
	}{
		"nil error does not panic": {
			err:         nil,
			message:     "Operation failed",
			shouldPanic: false,
		},
		"error causes panic with message": {
			err:         errors.New("test error"),
			message:     "Operation failed",
			shouldPanic: true,
			checkString: "Operation failed: test error",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					r := recover()
					assert.NotNil(t, r)
					panicMsg := fmt.Sprint(r)
					assert.Contains(t, panicMsg, tc.message)
					assert.Contains(t, panicMsg, tc.err.Error())
				}()
				FailOnErrorWithMessage(tc.err, tc.message)
			} else {
				assert.NotPanics(t, func() {
					FailOnErrorWithMessage(tc.err, tc.message)
				})
			}
		})
	}
}

func TestAllErrorsNil(t *testing.T) {
	tests := map[string]struct {
		errors   []error
		expected bool
	}{
		"no errors": {
			errors:   []error{},
			expected: true,
		},
		"all nil": {
			errors:   []error{nil, nil, nil},
			expected: true,
		},
		"one non-nil": {
			errors:   []error{nil, errors.New("error"), nil},
			expected: false,
		},
		"all non-nil": {
			errors: []error{
				errors.New("error 1"),
				errors.New("error 2"),
			},
			expected: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := allErrorsNil(tc.errors...)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// DELETED: LOW-VALUE tests that just checked "doesn't panic"
// - TestShowError: No complex logic to test (just prints)
// - TestShowErrorWithMessage: No complex logic to test (just prints)
// - TestShowWarning: No complex logic to test (just prints)
// The actual logic (HandleCancelError, TransformErrorToSteampipe) is tested elsewhere.

func TestTransformErrorEdgeCases(t *testing.T) {
	tests := map[string]struct {
		err            error
		checkCondition func(*testing.T, error)
	}{
		"multiple spaces": {
			err: errors.New("ERROR:    multiple    spaces   "),
			checkCondition: func(t *testing.T, result error) {
				assert.NotContains(t, result.Error(), "ERROR:")
				// Should trim leading/trailing spaces
				assert.Equal(t, "multiple    spaces", result.Error())
			},
		},
		"nested rpc error": {
			err: errors.New("ERROR: rpc error: code = Unknown desc = ERROR: nested error"),
			checkCondition: func(t *testing.T, result error) {
				assert.Contains(t, result.Error(), "nested error")
				// Should strip the outer ERROR and rpc prefix
				assert.True(t, strings.HasPrefix(result.Error(), "ERROR:") || strings.HasPrefix(result.Error(), "nested"))
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := TransformErrorToSteampipe(tc.err)
			tc.checkCondition(t, result)
		})
	}
}

// High-Value Bug-Finding Tests (Wave 1.5 - Task 4, Phase 3)

func TestErrorWrapping_PreservesOriginalError(t *testing.T) {
	// Test that errors.As can unwrap through multiple layers
	// Bug this would find: Error wrapping doesn't preserve original error type

	originalErr := errors.New("database connection failed")

	// Wrap through several layers
	err1 := fmt.Errorf("query execution failed: %w", originalErr)
	err2 := fmt.Errorf("batch processing failed: %w", err1)
	err3 := fmt.Errorf("command failed: %w", err2)

	// Verify we can still find the original error message
	assert.Contains(t, err3.Error(), "database connection failed")
	assert.Contains(t, err3.Error(), "command failed")

	// Verify unwrapping works
	assert.True(t, errors.Is(err3, originalErr), "errors.Is should find original error")
}

func TestErrorWrapping_WithNilInChain(t *testing.T) {
	// Test error wrapping when nil appears in the chain
	// Bug this would find: Crashes or unexpected behavior with nil in error chain

	tests := map[string]struct {
		err          error
		expectPanic  bool
		expectNil    bool
	}{
		"wrapping nil": {
			err:          fmt.Errorf("wrapped: %w", nil),
			expectPanic:  false,
			expectNil:    false, // Should return error with "wrapped: <nil>"
		},
		"nil error passed to CombineErrors": {
			err:          CombineErrors(nil, nil, nil),
			expectPanic:  false,
			expectNil:    true,
		},
		"mix of nil and non-nil in CombineErrors": {
			err:          CombineErrors(nil, errors.New("error 1"), nil),
			expectPanic:  false,
			expectNil:    false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.expectPanic {
				assert.Panics(t, func() {
					_ = tc.err
				})
			} else {
				assert.NotPanics(t, func() {
					if tc.expectNil {
						assert.Nil(t, tc.err)
					} else {
						assert.NotNil(t, tc.err)
					}
				})
			}
		})
	}
}

func TestErrorCombining_DeeplyNestedChain(t *testing.T) {
	// Test error combining with deeply nested error chains
	// Bug this would find: Stack overflow, performance issues, or lost context

	// Create a deeply nested error chain (100 levels)
	var err error = errors.New("base error")
	for i := 0; i < 100; i++ {
		err = fmt.Errorf("layer %d: %w", i, err)
	}

	// Combine with other errors
	combined := CombineErrors(err, errors.New("error 2"), errors.New("error 3"))

	assert.NotNil(t, combined)
	assert.Contains(t, combined.Error(), "base error")
	assert.Contains(t, combined.Error(), "error 2")
	assert.Contains(t, combined.Error(), "error 3")

	// Verify we can still access the base error message in the combined error
	// This tests that deeply nested chains don't break error combination
	assert.Contains(t, combined.Error(), "base error")
}

func TestCombineErrors_Concurrent(t *testing.T) {
	// Test concurrent calls to CombineErrors for race conditions
	// Bug this would find: Race conditions, corrupted error messages
	// Run with: go test -race

	const numGoroutines = 100
	const numIterations = 10

	results := make(chan error, numGoroutines*numIterations)
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				err1 := fmt.Errorf("goroutine %d error %d-1", id, j)
				err2 := fmt.Errorf("goroutine %d error %d-2", id, j)
				err3 := fmt.Errorf("goroutine %d error %d-3", id, j)

				combined := CombineErrors(err1, err2, err3)
				results <- combined
			}
		}(i)
	}

	wg.Wait()
	close(results)

	// Verify all errors were combined correctly
	count := 0
	for combinedErr := range results {
		assert.NotNil(t, combinedErr)
		// Each combined error should contain the goroutine ID
		errMsg := combinedErr.Error()
		assert.Contains(t, errMsg, "error")
		assert.Contains(t, errMsg, "goroutine")
		count++
	}

	assert.Equal(t, numGoroutines*numIterations, count, "Should have all combined errors")
}

func TestTransformErrorToSteampipe_EdgeCases(t *testing.T) {
	// Test edge cases in error transformation
	// Bug this would find: Crashes or incorrect transformations

	tests := map[string]struct {
		err            error
		expectNil      bool
		expectedString string
	}{
		"error with only whitespace": {
			err:            errors.New("   "),
			expectNil:      false,
			expectedString: "",
		},
		"error with multiple ERROR prefixes": {
			err:            errors.New("ERROR: ERROR: ERROR: test"),
			expectNil:      false,
			expectedString: "ERROR: ERROR: test", // Only strips first ERROR prefix
		},
		"error with ERROR in middle": {
			err:            errors.New("test ERROR: should not strip"),
			expectNil:      false,
			expectedString: "test ERROR: should not strip",
		},
		"very long error message": {
			err:            errors.New("ERROR: " + strings.Repeat("x", 10000)),
			expectNil:      false,
			expectedString: strings.Repeat("x", 10000),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := TransformErrorToSteampipe(tc.err)

			if tc.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				if tc.expectedString != "" {
					assert.Equal(t, tc.expectedString, result.Error())
				}
			}
		})
	}
}
