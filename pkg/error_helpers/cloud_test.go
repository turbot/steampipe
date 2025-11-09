package error_helpers

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsInvalidWorkspaceDatabaseArg(t *testing.T) {
	tests := map[string]struct {
		err      error
		expected bool
	}{
		"404 Not Found": {
			err:      errors.New("404 Not Found"),
			expected: true,
		},
		"nil error": {
			err:      nil,
			expected: false,
		},
		"different error": {
			err:      errors.New("500 Internal Server Error"),
			expected: false,
		},
		"404 with different format": {
			err:      errors.New("404 - Not Found"),
			expected: false,
		},
		"404 not found lowercase": {
			err:      errors.New("404 not found"),
			expected: false,
		},
		"wrapped 404": {
			err:      fmt.Errorf("request failed: %w", errors.New("404 Not Found")),
			expected: false, // Exact match required
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsInvalidWorkspaceDatabaseArg(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsInvalidCloudToken(t *testing.T) {
	tests := map[string]struct {
		err      error
		expected bool
	}{
		"401 Unauthorized": {
			err:      errors.New("401 Unauthorized"),
			expected: true,
		},
		"nil error": {
			err:      nil,
			expected: false,
		},
		"different error": {
			err:      errors.New("403 Forbidden"),
			expected: false,
		},
		"401 with different format": {
			err:      errors.New("401 - Unauthorized"),
			expected: false,
		},
		"401 unauthorized lowercase": {
			err:      errors.New("401 unauthorized"),
			expected: false,
		},
		"wrapped 401": {
			err:      fmt.Errorf("auth failed: %w", errors.New("401 Unauthorized")),
			expected: false, // Exact match required
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsInvalidCloudToken(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCloudErrorsExactMatch(t *testing.T) {
	// Test that the functions require exact matches
	tests := map[string]struct {
		err         error
		is404       bool
		is401       bool
	}{
		"exact 404": {
			err:   errors.New("404 Not Found"),
			is404: true,
			is401: false,
		},
		"exact 401": {
			err:   errors.New("401 Unauthorized"),
			is404: false,
			is401: true,
		},
		"404 with extra spaces": {
			err:   errors.New("404  Not Found"),
			is404: false,
			is401: false,
		},
		"401 with extra spaces": {
			err:   errors.New("401  Unauthorized"),
			is404: false,
			is401: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.is404, IsInvalidWorkspaceDatabaseArg(tc.err), "404 check failed")
			assert.Equal(t, tc.is401, IsInvalidCloudToken(tc.err), "401 check failed")
		})
	}
}

func TestCloudErrorsWithNil(t *testing.T) {
	// Both functions should handle nil gracefully
	assert.False(t, IsInvalidWorkspaceDatabaseArg(nil))
	assert.False(t, IsInvalidCloudToken(nil))
}

func TestCloudErrorsAreDistinct(t *testing.T) {
	// Test that 404 and 401 errors are properly distinguished
	err404 := errors.New("404 Not Found")
	err401 := errors.New("401 Unauthorized")

	assert.True(t, IsInvalidWorkspaceDatabaseArg(err404))
	assert.False(t, IsInvalidCloudToken(err404))

	assert.True(t, IsInvalidCloudToken(err401))
	assert.False(t, IsInvalidWorkspaceDatabaseArg(err401))
}
