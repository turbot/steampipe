package error_helpers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsContextCanceled(t *testing.T) {
	tests := map[string]struct {
		ctx      context.Context
		expected bool
	}{
		"canceled context": {
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			expected: true,
		},
		"active context": {
			ctx:      context.Background(),
			expected: false,
		},
		"timed out context": {
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 0)
				defer cancel()
				<-ctx.Done()
				return ctx
			}(),
			expected: false, // Timeout is not the same as cancellation
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsContextCanceled(tc.ctx)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsContextCancelledError(t *testing.T) {
	tests := map[string]struct {
		err      error
		expected bool
	}{
		"context.Canceled": {
			err:      context.Canceled,
			expected: true,
		},
		"nil error": {
			err:      nil,
			expected: false,
		},
		"regular error": {
			err:      errors.New("regular error"),
			expected: false,
		},
		"context.DeadlineExceeded": {
			err:      context.DeadlineExceeded,
			expected: false,
		},
		"wrapped context.Canceled": {
			err:      errors.Join(errors.New("operation failed"), context.Canceled),
			expected: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsContextCancelledError(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsContextCanceledWithContextError(t *testing.T) {
	// Test with a context that has an error
	ctx, cancel := context.WithCancel(context.Background())

	// Before canceling
	assert.False(t, IsContextCanceled(ctx))

	// After canceling
	cancel()
	assert.True(t, IsContextCanceled(ctx))
}

func TestIsContextCancelledErrorConsistency(t *testing.T) {
	// Test that IsContextCanceled and IsContextCancelledError are consistent
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Both should return true for a canceled context
	assert.True(t, IsContextCanceled(ctx))
	assert.True(t, IsContextCancelledError(ctx.Err()))
}

func TestIsContextCancelledErrorWithDifferentErrors(t *testing.T) {
	// Test various error types to ensure proper detection
	tests := map[string]struct {
		err      error
		expected bool
	}{
		"simple string error": {
			err:      errors.New("connection failed"),
			expected: false,
		},
		"error with 'cancel' in text": {
			err:      errors.New("please cancel the operation"),
			expected: false,
		},
		"actual context cancellation": {
			err:      context.Canceled,
			expected: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsContextCancelledError(tc.err)
			assert.Equal(t, tc.expected, result, "Error: %v", tc.err)
		})
	}
}
