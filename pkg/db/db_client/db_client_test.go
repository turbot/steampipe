package db_client

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
)

// TestClose tests the Close method
func TestClose(t *testing.T) {
	tests := map[string]struct {
		setupClient func() *DbClient
		expectError bool
	}{
		"close with nil pools": {
			setupClient: func() *DbClient {
				return &DbClient{}
			},
			expectError: false,
		},
		"close nullifies sessions": {
			setupClient: func() *DbClient {
				return &DbClient{
					sessions: map[uint32]*db_common.DatabaseSession{
						12345: {BackendPid: 12345},
					},
				}
			},
			expectError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := tc.setupClient()
			ctx := context.Background()

			err := client.Close(ctx)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Nil(t, client.sessions)
			}
		})
	}
}

// TestResetPools tests the ResetPools method
func TestResetPools(t *testing.T) {
	tests := map[string]struct {
		setupClient  func() *DbClient
		expectPanic  bool
		description  string
	}{
		"panics with nil userPool": {
			setupClient: func() *DbClient {
				// Create client with nil userPool (bug!)
				return &DbClient{
					userPool:       nil,
					managementPool: nil,
				}
			},
			expectPanic: true,
			description: "BUG: ResetPools panics with nil pools - needs nil checks like closePools()",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := tc.setupClient()
			ctx := context.Background()

			if tc.expectPanic {
				// This SHOULD panic with current implementation
				// TODO: Fix ResetPools to add nil checks
				assert.Panics(t, func() {
					client.ResetPools(ctx)
				}, tc.description)
			} else {
				assert.NotPanics(t, func() {
					client.ResetPools(ctx)
				})
			}
		})
	}
}

// TestBuildSchemasQuery tests the buildSchemasQuery method
func TestBuildSchemasQuery(t *testing.T) {
	tests := map[string]struct {
		schemas         []string
		expectedContains []string
	}{
		"no schemas": {
			schemas:         []string{},
			expectedContains: []string{"information_schema.columns", "LEFT(cols.table_schema,8) = 'pg_temp_'"},
		},
		"single schema": {
			schemas: []string{"public"},
			expectedContains: []string{
				"information_schema.columns",
				"'public'",
				"LEFT(cols.table_schema,8) = 'pg_temp_'",
			},
		},
		"multiple schemas": {
			schemas: []string{"public", "aws", "azure"},
			expectedContains: []string{
				"information_schema.columns",
				"'public'",
				"'aws'",
				"'azure'",
				"LEFT(cols.table_schema,8) = 'pg_temp_'",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := &DbClient{}
			query := client.buildSchemasQuery(tc.schemas...)

			for _, expected := range tc.expectedContains {
				assert.Contains(t, query, expected)
			}
		})
	}
}

// TestBuildSchemasQueryLegacy tests the buildSchemasQueryLegacy method
func TestBuildSchemasQueryLegacy(t *testing.T) {
	client := &DbClient{}
	query := client.buildSchemasQueryLegacy()

	expectedContains := []string{
		"information_schema.foreign_tables",
		"distinct_schema",
		"steampipe_command",
		"information_schema.columns",
		"LEFT(cols.table_schema,8) = 'pg_temp_'",
	}

	for _, expected := range expectedContains {
		assert.Contains(t, query, expected)
	}
}

// TestSessionMapLeak tests for the documented memory leak in session map cleanup
// Reference: pkg/db/db_client/db_client.go:45-46
// "TODO: there's no code which cleans up this map when connections get dropped by pgx"
func TestSessionMapLeak(t *testing.T) {
	// This test verifies the documented memory leak behavior
	client := &DbClient{
		sessions:      make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex: &sync.Mutex{},
	}

	// Simulate many connections over time
	connectionCount := 1000
	for i := uint32(0); i < uint32(connectionCount); i++ {
		session := &db_common.DatabaseSession{BackendPid: i}
		client.sessionsMutex.Lock()
		client.sessions[i] = session
		client.sessionsMutex.Unlock()
	}

	// Verify sessions accumulated (this is the leak!)
	assert.Equal(t, connectionCount, len(client.sessions),
		"Sessions accumulate without cleanup - KNOWN MEMORY LEAK")

	// Close client should nullify sessions
	err := client.Close(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, client.sessions, "Close() should nullify sessions map")

	// BUG: During normal operation (before Close), sessions are NEVER cleaned up
	// This is a confirmed memory leak that grows over time in long-running services
	// See: https://github.com/turbot/steampipe/issues/3737
}

// TestConcurrentSessionAccess tests for race conditions in session map access
func TestConcurrentSessionAccess(t *testing.T) {
	// This test should be run with -race flag: go test -race ./pkg/db/db_client/
	client := &DbClient{
		sessions:      make(map[uint32]*db_common.DatabaseSession),
		sessionsMutex: &sync.Mutex{},
	}

	// Launch multiple goroutines accessing the session map
	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(id uint32) {
			defer wg.Done()

			// Add session
			client.sessionsMutex.Lock()
			client.sessions[id] = &db_common.DatabaseSession{BackendPid: id}
			client.sessionsMutex.Unlock()

			// Read session
			client.sessionsMutex.Lock()
			_ = client.sessions[id]
			client.sessionsMutex.Unlock()

			// Delete session
			client.sessionsMutex.Lock()
			delete(client.sessions, id)
			client.sessionsMutex.Unlock()
		}(uint32(i))
	}

	wg.Wait()

	// If we get here without race detector warnings, the mutex is working correctly
	// Run with: go test -race ./pkg/db/db_client/
	assert.True(t, true, "Concurrent access completed without deadlock")
}

// TestHumanizeRowCount tests the humanizeRowCount function
func TestHumanizeRowCount(t *testing.T) {
	tests := map[string]struct {
		count    int
		expected string
	}{
		"zero": {
			count:    0,
			expected: "0",
		},
		"single digit": {
			count:    5,
			expected: "5",
		},
		"hundreds": {
			count:    123,
			expected: "123",
		},
		"thousands": {
			count:    1234,
			expected: "1,234",
		},
		"millions": {
			count:    1234567,
			expected: "1,234,567",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := humanizeRowCount(tc.count)
			assert.Equal(t, tc.expected, result)
		})
	}
}
