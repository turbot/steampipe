package error_helpers

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestDecodePgError(t *testing.T) {
	tests := map[string]struct {
		err            error
		expectedString string
		isPgError      bool
	}{
		"regular error": {
			err:            errors.New("regular error"),
			expectedString: "regular error",
			isPgError:      false,
		},
		"nil error": {
			err:            nil,
			expectedString: "",
			isPgError:      false,
		},
		"pg error - syntax error": {
			err: &pgconn.PgError{
				Severity: "ERROR",
				Code:     "42601",
				Message:  "syntax error at or near \"SELCT\"",
			},
			expectedString: "syntax error at or near \"SELCT\"",
			isPgError:      true,
		},
		"pg error - undefined column": {
			err: &pgconn.PgError{
				Severity: "ERROR",
				Code:     "42703",
				Message:  "column \"invalid_col\" does not exist",
			},
			expectedString: "column \"invalid_col\" does not exist",
			isPgError:      true,
		},
		"pg error - undefined table": {
			err: &pgconn.PgError{
				Severity: "ERROR",
				Code:     "42P01",
				Message:  "relation \"invalid_table\" does not exist",
			},
			expectedString: "relation \"invalid_table\" does not exist",
			isPgError:      true,
		},
		"pg error - with detail": {
			err: &pgconn.PgError{
				Severity: "ERROR",
				Code:     "23505",
				Message:  "duplicate key value violates unique constraint",
				Detail:   "Key (id)=(123) already exists.",
			},
			expectedString: "duplicate key value violates unique constraint",
			isPgError:      true,
		},
		"pg error - with hint": {
			err: &pgconn.PgError{
				Severity: "ERROR",
				Code:     "42P01",
				Message:  "relation \"users\" does not exist",
				Hint:     "Perhaps you meant to reference the table \"user\"?",
			},
			expectedString: "relation \"users\" does not exist",
			isPgError:      true,
		},
		"pg error - with position": {
			err: &pgconn.PgError{
				Severity: "ERROR",
				Code:     "42601",
				Message:  "syntax error at or near \"FROM\"",
				Position: 8,
			},
			expectedString: "syntax error at or near \"FROM\"",
			isPgError:      true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := DecodePgError(tc.err)

			if tc.expectedString == "" {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectedString, result.Error())
			}
		})
	}
}

func TestDecodePgErrorTypes(t *testing.T) {
	// Test that DecodePgError properly handles different error types
	tests := map[string]struct {
		err       error
		checkFunc func(*testing.T, error)
	}{
		"wrapped pg error": {
			err: wrapPgError(&pgconn.PgError{
				Message: "connection error",
			}),
			checkFunc: func(t *testing.T, result error) {
				assert.Equal(t, "connection error", result.Error())
			},
		},
		"non-pg error unchanged": {
			err: errors.New("standard error"),
			checkFunc: func(t *testing.T, result error) {
				assert.Equal(t, "standard error", result.Error())
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := DecodePgError(tc.err)
			tc.checkFunc(t, result)
		})
	}
}

func TestDecodePgErrorWithComplexMessages(t *testing.T) {
	tests := map[string]struct {
		pgError        *pgconn.PgError
		expectedString string
	}{
		"multi-line message": {
			pgError: &pgconn.PgError{
				Message: "ERROR: syntax error\nLINE 1: SELCT * FROM test",
			},
			expectedString: "ERROR: syntax error\nLINE 1: SELCT * FROM test",
		},
		"message with quotes": {
			pgError: &pgconn.PgError{
				Message: "column \"test_column\" does not exist",
			},
			expectedString: "column \"test_column\" does not exist",
		},
		"message with special characters": {
			pgError: &pgconn.PgError{
				Message: "syntax error at or near \";\"",
			},
			expectedString: "syntax error at or near \";\"",
		},
		"empty message": {
			pgError: &pgconn.PgError{
				Message: "",
			},
			expectedString: "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := DecodePgError(tc.pgError)
			assert.Equal(t, tc.expectedString, result.Error())
		})
	}
}

func TestDecodePgErrorPreservesMessage(t *testing.T) {
	// Test that the message is extracted correctly without modification
	pgErr := &pgconn.PgError{
		Severity:         "ERROR",
		Code:             "42P01",
		Message:          "relation \"test_table\" does not exist",
		Detail:           "The table you're trying to query doesn't exist",
		Hint:             "Check your table name",
		Position:         0,
		InternalPosition: 0,
		InternalQuery:    "",
		Where:            "",
		SchemaName:       "public",
		TableName:        "test_table",
		ColumnName:       "",
		DataTypeName:     "",
		ConstraintName:   "",
		File:             "parse_relation.c",
		Line:             1234,
		Routine:          "parserOpenTable",
	}

	result := DecodePgError(pgErr)
	assert.Equal(t, "relation \"test_table\" does not exist", result.Error())
	// Verify that additional fields (Detail, Hint, etc.) are not included
	assert.NotContains(t, result.Error(), "The table you're trying to query")
	assert.NotContains(t, result.Error(), "Check your table name")
}

// Helper function to wrap a PgError for testing wrapped errors
func wrapPgError(pgErr *pgconn.PgError) error {
	return pgErr
}

// High-Value Bug-Finding Tests (Wave 1.5 - Task 4, Phase 3)

func TestDecodePgError_MalformedInput(t *testing.T) {
	// Test handling of malformed PgError objects
	// Bug this would find: Crashes with empty/nil fields, unexpected behavior

	tests := map[string]struct {
		err          error
		expectPanic  bool
		checkFunc    func(*testing.T, error)
	}{
		"empty message": {
			err: &pgconn.PgError{
				Code:    "42601",
				Message: "",
			},
			expectPanic: false,
			checkFunc: func(t *testing.T, result error) {
				assert.NotNil(t, result)
				assert.Equal(t, "", result.Error())
			},
		},
		"nil error": {
			err:         nil,
			expectPanic: false,
			checkFunc: func(t *testing.T, result error) {
				// Should handle nil gracefully
				assert.Nil(t, result)
			},
		},
		"only code, no message": {
			err: &pgconn.PgError{
				Code:    "42P01",
				Message: "",
			},
			expectPanic: false,
			checkFunc: func(t *testing.T, result error) {
				assert.NotNil(t, result)
				// Message is empty, should return empty string
				assert.Equal(t, "", result.Error())
			},
		},
		"special characters in message": {
			err: &pgconn.PgError{
				Message: "syntax error: \x00\x01\x02",
			},
			expectPanic: false,
			checkFunc: func(t *testing.T, result error) {
				assert.NotNil(t, result)
				assert.Contains(t, result.Error(), "syntax error")
			},
		},
		"very long message": {
			err: &pgconn.PgError{
				Message: "ERROR: " + string(make([]byte, 100000)),
			},
			expectPanic: false,
			checkFunc: func(t *testing.T, result error) {
				assert.NotNil(t, result)
				// Should handle large messages without crashing
				assert.Greater(t, len(result.Error()), 10000)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.expectPanic {
				assert.Panics(t, func() {
					DecodePgError(tc.err)
				})
			} else {
				assert.NotPanics(t, func() {
					result := DecodePgError(tc.err)
					tc.checkFunc(t, result)
				})
			}
		})
	}
}

func TestDecodePgError_DeeplyNestedWrapping(t *testing.T) {
	// Test deeply nested error wrapping
	// Bug this would find: Stack overflow, performance issues

	originalPgErr := &pgconn.PgError{
		Code:    "42601",
		Message: "syntax error",
	}

	// Wrap the error 50 times
	var err error = originalPgErr
	for i := 0; i < 50; i++ {
		err = fmt.Errorf("layer %d: %w", i, err)
	}

	// Should handle deep nesting without crashing
	assert.NotPanics(t, func() {
		result := DecodePgError(err)
		assert.NotNil(t, result)
		assert.Equal(t, "syntax error", result.Error())
	})
}

func TestDecodePgError_ConcurrentAccess(t *testing.T) {
	// Test concurrent calls to DecodePgError
	// Bug this would find: Race conditions
	// Run with: go test -race

	pgErr := &pgconn.PgError{
		Code:    "42601",
		Message: "syntax error at position 10",
	}

	const numGoroutines = 50
	results := make(chan error, numGoroutines)
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := DecodePgError(pgErr)
			results <- result
		}()
	}

	wg.Wait()
	close(results)

	// Verify all goroutines got consistent results
	count := 0
	for result := range results {
		assert.NotNil(t, result)
		assert.Equal(t, "syntax error at position 10", result.Error())
		count++
	}

	assert.Equal(t, numGoroutines, count)
}
