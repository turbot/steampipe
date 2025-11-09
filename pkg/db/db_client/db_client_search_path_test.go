package db_client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetRequiredSessionSearchPath tests the GetRequiredSessionSearchPath method
func TestGetRequiredSessionSearchPath(t *testing.T) {
	tests := map[string]struct {
		customSearchPath []string
		userSearchPath   []string
		expected         []string
	}{
		"no custom search path - use user search path": {
			customSearchPath: nil,
			userSearchPath:   []string{"public", "aws", "azure"},
			expected:         []string{"public", "aws", "azure"},
		},
		"custom search path set - use custom": {
			customSearchPath: []string{"custom1", "custom2"},
			userSearchPath:   []string{"public", "aws", "azure"},
			expected:         []string{"custom1", "custom2"},
		},
		"empty custom search path": {
			customSearchPath: []string{},
			userSearchPath:   []string{"public"},
			expected:         []string{},
		},
		"single schema in user path": {
			customSearchPath: nil,
			userSearchPath:   []string{"public"},
			expected:         []string{"public"},
		},
		"multiple schemas in custom path": {
			customSearchPath: []string{"schema1", "schema2", "schema3", "schema4"},
			userSearchPath:   []string{"public"},
			expected:         []string{"schema1", "schema2", "schema3", "schema4"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := &DbClient{
				customSearchPath: tc.customSearchPath,
				userSearchPath:   tc.userSearchPath,
			}

			result := client.GetRequiredSessionSearchPath()

			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestSearchPathInteraction tests interaction between different search path fields
func TestSearchPathInteraction(t *testing.T) {
	tests := map[string]struct {
		client   *DbClient
		expected string
	}{
		"only user search path": {
			client: &DbClient{
				userSearchPath:   []string{"public", "aws"},
				customSearchPath: nil,
				searchPathPrefix: nil,
			},
			expected: "should use user search path",
		},
		"custom overrides user": {
			client: &DbClient{
				userSearchPath:   []string{"public", "aws"},
				customSearchPath: []string{"custom1", "custom2"},
				searchPathPrefix: nil,
			},
			expected: "should use custom search path",
		},
		"prefix with user path": {
			client: &DbClient{
				userSearchPath:   []string{"public", "aws"},
				customSearchPath: nil,
				searchPathPrefix: []string{"prefix1"},
			},
			expected: "should use user search path (prefix stored separately)",
		},
		"all fields set": {
			client: &DbClient{
				userSearchPath:   []string{"public", "aws"},
				customSearchPath: []string{"custom1", "custom2"},
				searchPathPrefix: []string{"prefix1"},
			},
			expected: "should use custom search path",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := tc.client.GetRequiredSessionSearchPath()

			if tc.client.customSearchPath != nil {
				assert.Equal(t, tc.client.customSearchPath, result, tc.expected)
			} else {
				assert.Equal(t, tc.client.userSearchPath, result, tc.expected)
			}
		})
	}
}

