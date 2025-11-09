package db_local

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStartListenType_ToListenAddresses tests the conversion of listen types to addresses
func TestStartListenType_ToListenAddresses(t *testing.T) {
	tests := map[string]struct {
		listenType StartListenType
		want       []string
	}{
		"network binding": {
			listenType: ListenTypeNetwork,
			want:       []string{"*"},
		},
		"local binding": {
			listenType: ListenTypeLocal,
			want:       []string{"localhost"},
		},
		"custom addresses": {
			listenType: StartListenType("192.168.1.1,10.0.0.1"),
			want:       []string{"192.168.1.1", "10.0.0.1"},
		},
		"single custom address": {
			listenType: StartListenType("192.168.1.1"),
			want:       []string{"192.168.1.1"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.listenType.ToListenAddresses()
			assert.Equal(t, tc.want, got, "ToListenAddresses() should return expected addresses")
		})
	}
}

// TestStartResult_SetError tests the SetError method
func TestStartResult_SetError(t *testing.T) {
	tests := map[string]struct {
		initialStatus StartDbStatus
		err           error
		wantStatus    StartDbStatus
	}{
		"set error on started service": {
			initialStatus: ServiceStarted,
			err:           assert.AnError,
			wantStatus:    ServiceFailedToStart,
		},
		"set error on already running service": {
			initialStatus: ServiceAlreadyRunning,
			err:           assert.AnError,
			wantStatus:    ServiceFailedToStart,
		},
		"set error on failed service": {
			initialStatus: ServiceFailedToStart,
			err:           assert.AnError,
			wantStatus:    ServiceFailedToStart,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res := &StartResult{
				Status: tc.initialStatus,
			}

			result := res.SetError(tc.err)

			assert.Equal(t, tc.wantStatus, result.Status, "Status should be ServiceFailedToStart")
			assert.Equal(t, tc.err, result.Error, "Error should be set")
			assert.Same(t, res, result, "SetError should return the same instance")
		})
	}
}

// TestGetListenAddresses tests the getListenAddresses function
func TestGetListenAddresses(t *testing.T) {
	tests := map[string]struct {
		input       []string
		expectLocal bool // Whether to expect localhost/loopback addresses
		expectOther bool // Whether to expect other addresses
	}{
		"localhost binding": {
			input:       []string{"localhost"},
			expectLocal: true,
			expectOther: false,
		},
		"wildcard binding": {
			input:       []string{"*"},
			expectLocal: true,
			expectOther: true, // May include public addresses
		},
		"specific IP addresses": {
			input:       []string{"192.168.1.1", "10.0.0.1"},
			expectLocal: false,
			expectOther: true,
		},
		"mixed addresses": {
			input:       []string{"localhost", "192.168.1.1"},
			expectLocal: true,
			expectOther: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := getListenAddresses(tc.input)

			// Verify we got addresses back
			if tc.expectLocal || tc.expectOther {
				assert.NotEmpty(t, got, "Should return addresses")
			}

			// Check for localhost/loopback addresses
			if tc.expectLocal {
				hasLocal := false
				for _, addr := range got {
					if addr == "127.0.0.1" || addr == "::1" || addr == "localhost" {
						hasLocal = true
						break
					}
				}
				assert.True(t, hasLocal, "Should contain localhost/loopback addresses")
			}

			// Verify addresses are unique
			seen := make(map[string]bool)
			for _, addr := range got {
				assert.False(t, seen[addr], "Addresses should be unique: %s", addr)
				seen[addr] = true
			}
		})
	}
}

// TestEnsurePluginManager tests plugin manager initialization logic
func TestEnsurePluginManager(t *testing.T) {
	t.Skip("Skipping test that requires actual plugin manager - needs mocking infrastructure")

	// This test would require:
	// 1. Mocking pluginmanager.LoadState()
	// 2. Mocking pluginmanager.StartNewInstance()
	// 3. Mocking pluginmanager.NewPluginManagerClient()
	//
	// These would be good candidates for future enhancement
}

// TestStartServices_Integration would test the full StartServices flow
// This is intentionally skipped as it requires:
// - Database installation
// - Port availability
// - File system permissions
// - Process management
func TestStartServices_Integration(t *testing.T) {
	t.Skip("Integration test - requires full environment setup")

	// This would be a good candidate for acceptance testing
	// The existing BATS tests in tests/acceptance/test_files/service.bats
	// already cover this integration scenario
}

// TestPostServiceStart tests post-startup setup logic
func TestPostServiceStart(t *testing.T) {
	t.Skip("Requires database connection - would benefit from mock database client")

	// This test would verify:
	// - setupInternal() is called
	// - initializeConnectionStateTable() is called
	// - PopulatePluginTable() is called
	// - setupServerSettingsTable() is called
	// - restoreDBBackup() is called
	// - SQL functions are created
}

// TestStartDB_ErrorHandling tests that startDB properly cleans up on errors
func TestStartDB_ErrorHandling(t *testing.T) {
	t.Skip("Requires process mocking - complex integration test")

	// This test would verify:
	// - When startDB fails, it calls StopServices
	// - It removes the running instance info file
	// - It kills the postgres process if started
}
