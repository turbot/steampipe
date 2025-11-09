package plugin

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/v2/versionfile"
	"github.com/turbot/steampipe/v2/pkg/test/helpers"
)

// TestPluginListItem removed - trivial struct field comparison tests
// Only validated that item == expected with no business logic
// Documented in cleanup report

// TestDetectLocalPlugin tests the detectLocalPlugin function
func TestDetectLocalPlugin(t *testing.T) {
	tests := map[string]struct {
		setup          func(t *testing.T) (installation *versionfile.InstalledVersion, pluginPath string)
		expectedResult bool
	}{
		"plugin modified after installation": {
			setup: func(t *testing.T) (*versionfile.InstalledVersion, string) {
				tempDir := helpers.CreateTempDir(t)
				pluginPath := filepath.Join(tempDir, "plugin.so")

				// Create plugin file
				err := os.WriteFile(pluginPath, []byte("fake plugin"), 0644)
				assert.NoError(t, err)

				// Installation date is 1 hour ago
				installDate := time.Now().Add(-1 * time.Hour)

				installation := &versionfile.InstalledVersion{
					Name:        "turbot/aws",
					Version:     "1.0.0",
					InstallDate: installDate.Format(time.RFC3339),
				}

				// Touch the file to update mod time to now
				now := time.Now()
				err = os.Chtimes(pluginPath, now, now)
				assert.NoError(t, err)

				return installation, pluginPath
			},
			expectedResult: true,
		},
		"plugin not modified after installation": {
			setup: func(t *testing.T) (*versionfile.InstalledVersion, string) {
				tempDir := helpers.CreateTempDir(t)
				pluginPath := filepath.Join(tempDir, "plugin.so")

				// Installation date is 1 hour in the future (plugin was modified before "installation")
				installDate := time.Now().Add(1 * time.Hour)

				installation := &versionfile.InstalledVersion{
					Name:        "turbot/aws",
					Version:     "1.0.0",
					InstallDate: installDate.Format(time.RFC3339),
				}

				// Create plugin file with current time
				err := os.WriteFile(pluginPath, []byte("fake plugin"), 0644)
				assert.NoError(t, err)

				return installation, pluginPath
			},
			expectedResult: false,
		},
		"invalid install date format": {
			setup: func(t *testing.T) (*versionfile.InstalledVersion, string) {
				tempDir := helpers.CreateTempDir(t)
				pluginPath := filepath.Join(tempDir, "plugin.so")

				installation := &versionfile.InstalledVersion{
					Name:        "turbot/aws",
					Version:     "1.0.0",
					InstallDate: "invalid-date",
				}

				// Create plugin file
				err := os.WriteFile(pluginPath, []byte("fake plugin"), 0644)
				assert.NoError(t, err)

				return installation, pluginPath
			},
			expectedResult: false,
		},
		"plugin file does not exist": {
			setup: func(t *testing.T) (*versionfile.InstalledVersion, string) {
				tempDir := helpers.CreateTempDir(t)
				pluginPath := filepath.Join(tempDir, "nonexistent.so")

				installDate := time.Now().Add(-1 * time.Hour)

				installation := &versionfile.InstalledVersion{
					Name:        "turbot/aws",
					Version:     "1.0.0",
					InstallDate: installDate.Format(time.RFC3339),
				}

				return installation, pluginPath
			},
			expectedResult: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			installation, pluginPath := tc.setup(t)
			result := detectLocalPlugin(installation, pluginPath)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

// TestDetectLocalPluginTimeTruncation tests that time truncation works correctly
func TestDetectLocalPluginTimeTruncation(t *testing.T) {
	tempDir := helpers.CreateTempDir(t)
	pluginPath := filepath.Join(tempDir, "plugin.so")

	// Create a timestamp with nanoseconds
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 123456789, time.UTC)

	// Installation date with nanoseconds
	installation := &versionfile.InstalledVersion{
		Name:        "turbot/aws",
		Version:     "1.0.0",
		InstallDate: baseTime.Format(time.RFC3339),
	}

	// Create plugin file
	err := os.WriteFile(pluginPath, []byte("fake plugin"), 0644)
	assert.NoError(t, err)

	// Set mod time to slightly after (within same second)
	modTime := baseTime.Add(500 * time.Millisecond)
	err = os.Chtimes(pluginPath, modTime, modTime)
	assert.NoError(t, err)

	// Should return false because times are truncated to seconds and would be equal
	result := detectLocalPlugin(installation, pluginPath)
	assert.False(t, result, "times within same second should be considered equal after truncation")
}

// TestDetectLocalPluginEdgeCases tests edge cases
func TestDetectLocalPluginEdgeCases(t *testing.T) {
	tests := map[string]struct {
		setup          func(t *testing.T) (*versionfile.InstalledVersion, string)
		expectedResult bool
		description    string
	}{
		"exactly same time": {
			setup: func(t *testing.T) (*versionfile.InstalledVersion, string) {
				tempDir := helpers.CreateTempDir(t)
				pluginPath := filepath.Join(tempDir, "plugin.so")

				sameTime := time.Now().Truncate(time.Second)

				installation := &versionfile.InstalledVersion{
					Name:        "turbot/aws",
					Version:     "1.0.0",
					InstallDate: sameTime.Format(time.RFC3339),
				}

				// Create plugin file
				err := os.WriteFile(pluginPath, []byte("fake plugin"), 0644)
				assert.NoError(t, err)

				// Set exact same time
				err = os.Chtimes(pluginPath, sameTime, sameTime)
				assert.NoError(t, err)

				return installation, pluginPath
			},
			expectedResult: false,
			description:    "same time should not be detected as local",
		},
		"one second difference": {
			setup: func(t *testing.T) (*versionfile.InstalledVersion, string) {
				tempDir := helpers.CreateTempDir(t)
				pluginPath := filepath.Join(tempDir, "plugin.so")

				installTime := time.Now().Truncate(time.Second)
				modTime := installTime.Add(1 * time.Second)

				installation := &versionfile.InstalledVersion{
					Name:        "turbot/aws",
					Version:     "1.0.0",
					InstallDate: installTime.Format(time.RFC3339),
				}

				// Create plugin file
				err := os.WriteFile(pluginPath, []byte("fake plugin"), 0644)
				assert.NoError(t, err)

				// Set mod time 1 second later
				err = os.Chtimes(pluginPath, modTime, modTime)
				assert.NoError(t, err)

				return installation, pluginPath
			},
			expectedResult: true,
			description:    "one second later should be detected as local",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			installation, pluginPath := tc.setup(t)
			result := detectLocalPlugin(installation, pluginPath)
			assert.Equal(t, tc.expectedResult, result, tc.description)
		})
	}
}
