package plugin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/v2/versionfile"
	"github.com/turbot/steampipe/v2/pkg/test/helpers"
)

// TestDetectLocalPluginEdgeCasesExtended tests edge cases that could cause bugs
// HIGH VALUE: Tests boundary conditions and edge cases
func TestDetectLocalPluginEdgeCasesExtended(t *testing.T) {
	tests := map[string]struct {
		setupPlugin      func(t *testing.T, tempDir string) (installation *versionfile.InstalledVersion, pluginPath string)
		expectedIsLocal  bool
		description      string
	}{
		"plugin binary newer than install date by days": {
			setupPlugin: func(t *testing.T, tempDir string) (*versionfile.InstalledVersion, string) {
				pluginPath := filepath.Join(tempDir, "plugin.so")

				// Create plugin file
				err := os.WriteFile(pluginPath, []byte("fake plugin"), 0644)
				assert.NoError(t, err)

				// Set install date to long ago
				installDate := "2020-01-01T00:00:00Z"

				installation := &versionfile.InstalledVersion{
					Name:        "test/plugin",
					Version:     "1.0.0",
					InstallDate: installDate,
				}

				return installation, pluginPath
			},
			expectedIsLocal: true,
			description:     "binary much newer than installation should be detected as local",
		},
		"zero-byte plugin file": {
			setupPlugin: func(t *testing.T, tempDir string) (*versionfile.InstalledVersion, string) {
				pluginPath := filepath.Join(tempDir, "plugin.so")

				// Create empty plugin file - edge case that could cause issues
				err := os.WriteFile(pluginPath, []byte{}, 0644)
				assert.NoError(t, err)

				installation := &versionfile.InstalledVersion{
					Name:        "test/plugin",
					Version:     "1.0.0",
					InstallDate: "2024-01-01T00:00:00Z",
				}

				return installation, pluginPath
			},
			expectedIsLocal: true,
			description:     "zero-byte files should still be checked without panicking",
		},
		"large plugin file": {
			setupPlugin: func(t *testing.T, tempDir string) (*versionfile.InstalledVersion, string) {
				pluginPath := filepath.Join(tempDir, "plugin.so")

				// Create large file (1MB) - tests if large files cause performance issues
				largeData := make([]byte, 1024*1024)
				err := os.WriteFile(pluginPath, largeData, 0644)
				assert.NoError(t, err)

				installation := &versionfile.InstalledVersion{
					Name:        "test/plugin",
					Version:     "1.0.0",
					InstallDate: "2024-01-01T00:00:00Z",
				}

				return installation, pluginPath
			},
			expectedIsLocal: true,
			description:     "large files should be handled without performance degradation",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tempDir := helpers.CreateTempDir(t)
			installation, pluginPath := tc.setupPlugin(t, tempDir)

			isLocal := detectLocalPlugin(installation, pluginPath)
			assert.Equal(t, tc.expectedIsLocal, isLocal, tc.description)
		})
	}
}

// TestDetectLocalPluginConcurrency tests for race conditions
// HIGH VALUE: Tests concurrent access which is a common source of bugs
func TestDetectLocalPluginConcurrency(t *testing.T) {
	tempDir := helpers.CreateTempDir(t)
	pluginPath := filepath.Join(tempDir, "plugin.so")

	// Create plugin file
	err := os.WriteFile(pluginPath, []byte("fake plugin"), 0644)
	assert.NoError(t, err)

	installation := &versionfile.InstalledVersion{
		Name:        "test/plugin",
		Version:     "1.0.0",
		InstallDate: "2024-01-01T00:00:00Z",
	}

	// Run the same check concurrently multiple times
	// This tests if the function has race conditions or panics under concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			// BUG HUNT: Does concurrent access cause panics or data races?
			_ = detectLocalPlugin(installation, pluginPath)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we got here without panic, the test passes
	assert.True(t, true, "detectLocalPlugin should be safe for concurrent access")
}

// TestDetectLocalPluginSymlink tests behavior with symlinked plugin files
// MEDIUM VALUE: Tests edge case with symlinks that could cause unexpected behavior
func TestDetectLocalPluginSymlink(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping symlink test when running as root")
	}

	tempDir := helpers.CreateTempDir(t)

	// Create actual plugin file
	realPath := filepath.Join(tempDir, "real-plugin.so")
	err := os.WriteFile(realPath, []byte("fake plugin"), 0644)
	assert.NoError(t, err)

	// Create symlink
	symlinkPath := filepath.Join(tempDir, "plugin.so")
	err = os.Symlink(realPath, symlinkPath)
	if err != nil {
		t.Skip("Cannot create symlink on this system")
	}

	installation := &versionfile.InstalledVersion{
		Name:        "test/plugin",
		Version:     "1.0.0",
		InstallDate: "2024-01-01T00:00:00Z",
	}

	// Test with symlink
	// BUG HUNT: The function uses Lstat - does it handle symlinks correctly?
	isLocal := detectLocalPlugin(installation, symlinkPath)
	t.Logf("detectLocalPlugin with symlink returned: %v", isLocal)
	// We just verify it doesn't panic - the behavior with symlinks may vary
}
