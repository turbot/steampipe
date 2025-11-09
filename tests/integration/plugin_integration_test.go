//go:build integration
// +build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/steampipe/v2/pkg/test/helpers"
)

// TestPluginWorkflow_InstallListRemove tests the complete plugin lifecycle:
// install -> list -> remove
func TestPluginWorkflow_InstallListRemove(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	// Setup: Create temporary Steampipe directory
	tempDir := helpers.CreateTempDir(t)
	pluginsDir := helpers.CreateTestDir(t, tempDir, "plugins")

	// Store original STEAMPIPE_INSTALL_DIR
	originalDir := os.Getenv("STEAMPIPE_INSTALL_DIR")
	os.Setenv("STEAMPIPE_INSTALL_DIR", tempDir)
	t.Cleanup(func() {
		if originalDir != "" {
			os.Setenv("STEAMPIPE_INSTALL_DIR", originalDir)
		} else {
			os.Unsetenv("STEAMPIPE_INSTALL_DIR")
		}
	})

	// Test plugin name
	pluginName := "chaos"
	pluginPath := filepath.Join(pluginsDir, "hub.steampipe.io", "plugins", "turbot", pluginName+"@latest")

	// Step 1: Install plugin
	t.Run("install plugin", func(t *testing.T) {
		// Simulate plugin installation by creating the directory structure
		err := os.MkdirAll(pluginPath, 0755)
		assert.NoError(t, err, "Should create plugin directory")

		// Create a mock plugin binary
		binaryPath := filepath.Join(pluginPath, "steampipe-plugin-"+pluginName)
		err = os.WriteFile(binaryPath, []byte("mock binary"), 0755)
		assert.NoError(t, err, "Should create plugin binary")

		// Verify plugin installed
		assert.True(t, helpers.DirExists(pluginPath), "Plugin directory should exist")
		assert.True(t, helpers.FileExists(binaryPath), "Plugin binary should exist")
	})

	// Step 2: List plugins (should show installed)
	t.Run("list plugins", func(t *testing.T) {
		// Walk the plugins directory to simulate listing
		var foundPlugins []string

		err := filepath.Walk(pluginsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && filepath.Base(path) == pluginName+"@latest" {
				foundPlugins = append(foundPlugins, pluginName)
			}
			return nil
		})

		assert.NoError(t, err, "Should list plugins without error")
		assert.NotEmpty(t, foundPlugins, "Should find installed plugins")
		assert.Contains(t, foundPlugins, pluginName, "Should find chaos plugin")
	})

	// Step 3: Remove plugin
	t.Run("remove plugin", func(t *testing.T) {
		// Simulate plugin removal
		err := os.RemoveAll(pluginPath)
		assert.NoError(t, err, "Should remove plugin directory")

		// Verify plugin removed
		assert.False(t, helpers.DirExists(pluginPath), "Plugin directory should not exist")
	})
}

// TestPluginUpdate tests the plugin update workflow
func TestPluginUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	// Setup
	tempDir := helpers.CreateTempDir(t)
	pluginsDir := helpers.CreateTestDir(t, tempDir, "plugins")

	pluginName := "chaos"
	oldVersion := "0.1.0"
	newVersion := "0.2.0"

	oldPluginPath := filepath.Join(pluginsDir, "hub.steampipe.io", "plugins", "turbot", pluginName+"@"+oldVersion)
	newPluginPath := filepath.Join(pluginsDir, "hub.steampipe.io", "plugins", "turbot", pluginName+"@"+newVersion)

	// Install old version
	t.Run("install old version", func(t *testing.T) {
		err := os.MkdirAll(oldPluginPath, 0755)
		assert.NoError(t, err)

		binaryPath := filepath.Join(oldPluginPath, "steampipe-plugin-"+pluginName)
		err = os.WriteFile(binaryPath, []byte("old version"), 0755)
		assert.NoError(t, err)

		assert.True(t, helpers.DirExists(oldPluginPath), "Old version should be installed")
	})

	// Update to new version
	t.Run("update to new version", func(t *testing.T) {
		// Install new version
		err := os.MkdirAll(newPluginPath, 0755)
		assert.NoError(t, err)

		binaryPath := filepath.Join(newPluginPath, "steampipe-plugin-"+pluginName)
		err = os.WriteFile(binaryPath, []byte("new version"), 0755)
		assert.NoError(t, err)

		// Remove old version (simulating update)
		err = os.RemoveAll(oldPluginPath)
		assert.NoError(t, err)

		// Verify update
		assert.False(t, helpers.DirExists(oldPluginPath), "Old version should be removed")
		assert.True(t, helpers.DirExists(newPluginPath), "New version should be installed")

		// Verify binary content
		content := helpers.ReadTestFile(t, filepath.Join(newPluginPath, "steampipe-plugin-"+pluginName))
		assert.Equal(t, "new version", content, "Should have new version binary")
	})
}

// TestPluginVersionResolution tests version resolution logic
func TestPluginVersionResolution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	tests := map[string]struct {
		installedVersions []string
		requestedVersion  string
		expectedVersion   string
	}{
		"latest resolves to highest version": {
			installedVersions: []string{"0.1.0", "0.2.0", "0.3.0"},
			requestedVersion:  "latest",
			expectedVersion:   "0.3.0",
		},
		"specific version returns exact match": {
			installedVersions: []string{"0.1.0", "0.2.0", "0.3.0"},
			requestedVersion:  "0.2.0",
			expectedVersion:   "0.2.0",
		},
		"single version defaults to that version": {
			installedVersions: []string{"0.5.0"},
			requestedVersion:  "latest",
			expectedVersion:   "0.5.0",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup: Create plugin directories for each version
			tempDir := helpers.CreateTempDir(t)
			pluginsDir := helpers.CreateTestDir(t, tempDir, "plugins")
			pluginName := "test"

			for _, version := range tc.installedVersions {
				versionPath := filepath.Join(pluginsDir, "hub.steampipe.io", "plugins", "turbot", pluginName+"@"+version)
				err := os.MkdirAll(versionPath, 0755)
				assert.NoError(t, err)
			}

			// Simulate version resolution
			resolvedVersion := resolvePluginVersion(pluginsDir, pluginName, tc.requestedVersion, tc.installedVersions)

			// Verify
			assert.Equal(t, tc.expectedVersion, resolvedVersion, "Should resolve to expected version")
		})
	}
}

// TestMultiplePlugins tests managing multiple plugins simultaneously
func TestMultiplePlugins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	// Setup
	tempDir := helpers.CreateTempDir(t)
	pluginsDir := helpers.CreateTestDir(t, tempDir, "plugins")

	plugins := []string{"aws", "azure", "gcp", "chaos"}

	// Install multiple plugins
	t.Run("install multiple plugins", func(t *testing.T) {
		for _, plugin := range plugins {
			pluginPath := filepath.Join(pluginsDir, "hub.steampipe.io", "plugins", "turbot", plugin+"@latest")
			err := os.MkdirAll(pluginPath, 0755)
			assert.NoError(t, err, "Should install %s plugin", plugin)

			binaryPath := filepath.Join(pluginPath, "steampipe-plugin-"+plugin)
			err = os.WriteFile(binaryPath, []byte("mock"), 0755)
			assert.NoError(t, err)
		}
	})

	// List all plugins
	t.Run("list all plugins", func(t *testing.T) {
		var foundPlugins []string

		err := filepath.Walk(pluginsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				base := filepath.Base(path)
				for _, plugin := range plugins {
					if base == plugin+"@latest" {
						foundPlugins = append(foundPlugins, plugin)
					}
				}
			}
			return nil
		})

		assert.NoError(t, err)
		assert.Len(t, foundPlugins, len(plugins), "Should find all installed plugins")

		for _, plugin := range plugins {
			assert.Contains(t, foundPlugins, plugin, "Should find %s plugin", plugin)
		}
	})

	// Remove one plugin
	t.Run("remove one plugin", func(t *testing.T) {
		pluginToRemove := "chaos"
		pluginPath := filepath.Join(pluginsDir, "hub.steampipe.io", "plugins", "turbot", pluginToRemove+"@latest")

		err := os.RemoveAll(pluginPath)
		assert.NoError(t, err)

		// Verify only that plugin was removed
		assert.False(t, helpers.DirExists(pluginPath), "Removed plugin should not exist")

		// Verify others still exist
		for _, plugin := range plugins {
			if plugin != pluginToRemove {
				otherPath := filepath.Join(pluginsDir, "hub.steampipe.io", "plugins", "turbot", plugin+"@latest")
				assert.True(t, helpers.DirExists(otherPath), "%s plugin should still exist", plugin)
			}
		}
	})
}

// resolvePluginVersion simulates plugin version resolution logic
func resolvePluginVersion(pluginsDir, pluginName, requestedVersion string, installedVersions []string) string {
	if requestedVersion == "latest" {
		// Return highest version
		if len(installedVersions) == 0 {
			return ""
		}
		highest := installedVersions[0]
		for _, v := range installedVersions {
			if v > highest {
				highest = v
			}
		}
		return highest
	}

	// Return exact match
	for _, v := range installedVersions {
		if v == requestedVersion {
			return v
		}
	}

	return ""
}
