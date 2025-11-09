package db_local

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRunningDBInstanceInfo_MatchWithGivenListenAddresses tests listen address matching
func TestRunningDBInstanceInfo_MatchWithGivenListenAddresses(t *testing.T) {
	tests := map[string]struct {
		stored []string
		given  []string
		want   bool
	}{
		"exact match": {
			stored: []string{"localhost", "192.168.1.1"},
			given:  []string{"localhost", "192.168.1.1"},
			want:   true,
		},
		"match with different order": {
			stored: []string{"192.168.1.1", "localhost"},
			given:  []string{"localhost", "192.168.1.1"},
			want:   true,
		},
		"different addresses": {
			stored: []string{"localhost"},
			given:  []string{"192.168.1.1"},
			want:   false,
		},
		"subset does not match": {
			stored: []string{"localhost", "192.168.1.1"},
			given:  []string{"localhost"},
			want:   false,
		},
		"superset does not match": {
			stored: []string{"localhost"},
			given:  []string{"localhost", "192.168.1.1"},
			want:   false,
		},
		"empty slices match": {
			stored: []string{},
			given:  []string{},
			want:   true,
		},
		"single address match": {
			stored: []string{"localhost"},
			given:  []string{"localhost"},
			want:   true,
		},
		"wildcard vs localhost": {
			stored: []string{"*"},
			given:  []string{"localhost"},
			want:   false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			info := &RunningDBInstanceInfo{
				GivenListenAddresses: tc.stored,
			}

			got := info.MatchWithGivenListenAddresses(tc.given)
			assert.Equal(t, tc.want, got, "MatchWithGivenListenAddresses should return expected result")
		})
	}
}

// TestRunningDBInstanceInfo_String tests the String method redacts password
func TestRunningDBInstanceInfo_String(t *testing.T) {
	tests := map[string]struct {
		password string
	}{
		"simple password": {
			password: "password123",
		},
		"complex password": {
			password: "P@ssw0rd!#$%^&*()",
		},
		"empty password": {
			password: "",
		},
		"long password": {
			password: "verylongpasswordthatshouldalsoberedactedcorrectly123456789",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			info := &RunningDBInstanceInfo{
				Pid:                     12345,
				ResolvedListenAddresses: []string{"127.0.0.1"},
				GivenListenAddresses:    []string{"localhost"},
				Port:                    9193,
				User:                    "steampipe",
				Password:                tc.password,
				Database:                "steampipe",
				StructVersion:           RunningDBStructVersion,
			}

			str := info.String()

			// Verify password is redacted
			if tc.password != "" {
				assert.NotContains(t, str, tc.password, "Password should be redacted")
			}
			assert.Contains(t, str, "XXXX-XXXX-XXXX", "Should contain redacted password placeholder")

			// Verify other fields are present
			assert.Contains(t, str, "12345", "Should contain PID")
			assert.Contains(t, str, "9193", "Should contain port")
			assert.Contains(t, str, "steampipe", "Should contain user/database")

			// Verify original password is unchanged after String() call
			assert.Equal(t, tc.password, info.Password, "Original password should not be modified")
		})
	}
}

// TestRunningDBInstanceInfo_Save tests the Save method
func TestRunningDBInstanceInfo_Save(t *testing.T) {
	t.Skip("Requires filesystem mocking with temp files")

	// This test would verify:
	// - Instance info is saved as valid JSON
	// - StructVersion is set correctly before saving
	// - File permissions are correct (0644)
	// - All fields are serialized properly including password
	// - Nested arrays are serialized correctly
}

// TestLoadRunningInstanceInfo_ValidJSON tests loading valid JSON
func TestLoadRunningInstanceInfo_ValidJSON(t *testing.T) {
	t.Skip("Requires filesystem mocking with temp files")

	// This test would verify:
	// - Valid JSON file is loaded correctly
	// - All fields are deserialized properly
	// - Password is loaded correctly (not redacted in file)
}

// TestLoadRunningInstanceInfo_MissingFile tests handling of missing file
func TestLoadRunningInstanceInfo_MissingFile(t *testing.T) {
	t.Skip("Requires filesystem mocking")

	// This test would verify:
	// - Missing file returns nil, nil
	// - No error is returned for missing file
}

// TestLoadRunningInstanceInfo_InvalidJSON tests handling of corrupted JSON
func TestLoadRunningInstanceInfo_InvalidJSON(t *testing.T) {
	t.Skip("Requires filesystem mocking with corrupted file")

	// This test would verify:
	// - Invalid JSON returns nil (not error)
	// - Logged error about unmarshaling
	// - System can recover from corrupted state
}

// TestRemoveRunningInstanceInfo_Success tests successful removal
func TestRemoveRunningInstanceInfo_Success(t *testing.T) {
	t.Skip("Requires filesystem mocking")

	// This test would verify:
	// - Running info file is removed
	// - No error on successful removal
}

// TestRemoveRunningInstanceInfo_MissingFile tests removal of non-existent file
func TestRemoveRunningInstanceInfo_MissingFile(t *testing.T) {
	t.Skip("Requires filesystem mocking")

	// This test would verify:
	// - Error is returned when file doesn't exist
	// - This is expected behavior (os.Remove returns error)
}
