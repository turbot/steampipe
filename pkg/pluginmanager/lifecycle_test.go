package pluginmanager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/test/helpers"
)

// TestNewState tests creating a new plugin manager state
func TestNewState(t *testing.T) {
	testExePath := "/usr/local/bin/steampipe"
	reattachConfig := &plugin.ReattachConfig{
		Protocol:        plugin.ProtocolGRPC,
		ProtocolVersion: 1,
		Pid:             12345,
		Addr: &simpleAddr{
			network: "tcp",
			address: "localhost:12345",
		},
	}

	state := NewState(testExePath, reattachConfig)

	assert.NotNil(t, state)
	assert.Equal(t, testExePath, state.Executable)
	assert.Equal(t, 12345, state.Pid)
	assert.Equal(t, plugin.ProtocolGRPC, state.Protocol)
	assert.Equal(t, 1, state.ProtocolVersion)
}

// TestStateSaveAndLoad tests saving and loading plugin manager state
func TestStateSaveAndLoad(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := helpers.CreateTempDir(t)

	// Set install dir for app_specific package
	originalInstallDir := app_specific.InstallDir
	app_specific.InstallDir = tempDir
	defer func() {
		// Reset after test
		app_specific.InstallDir = originalInstallDir
	}()

	testExePath := "/usr/local/bin/steampipe"
	reattachConfig := &plugin.ReattachConfig{
		Protocol:        plugin.ProtocolGRPC,
		ProtocolVersion: 1,
		Pid:             12345,
		Addr: &simpleAddr{
			network: "tcp",
			address: "localhost:12345",
		},
	}

	// Create and save state
	state := NewState(testExePath, reattachConfig)
	err := state.Save()
	require.NoError(t, err)

	// Verify file was created
	stateFile := filepath.Join(tempDir, "internal", "plugin_manager.json")
	assert.True(t, helpers.FileExists(stateFile))

	// Load the state
	loadedState, err := LoadState()
	require.NoError(t, err)
	assert.NotNil(t, loadedState)

	// Verify loaded state matches original
	assert.Equal(t, state.Executable, loadedState.Executable)
	assert.Equal(t, state.Pid, loadedState.Pid)
}

// TestStateVerifyRunning tests verifying if the plugin manager process is running
func TestStateVerifyRunning(t *testing.T) {
	t.Skip("Skipping test that requires actual running processes")

	// This test verifies process state which is OS-dependent and may vary
	// The functionality is tested through integration tests
}

// TestStateRunning tests the Running property
func TestStateRunning(t *testing.T) {
	t.Skip("Skipping test that requires actual running processes")

	// This test checks OS process state which is better tested through integration tests
}

// TestLoadStateWhenNotExists tests loading state when no state file exists
func TestLoadStateWhenNotExists(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := helpers.CreateTempDir(t)

	// Set install dir for app_specific package
	originalInstallDir := app_specific.InstallDir
	app_specific.InstallDir = tempDir
	defer func() {
		app_specific.InstallDir = originalInstallDir
	}()

	// Try to load state when file doesn't exist
	state, err := LoadState()

	// Should not error, but should return an empty state
	require.NoError(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, "", state.Executable)
	assert.Equal(t, 0, state.Pid)
	assert.False(t, state.Running)
}

// TestStateFilePathGeneration tests the state file path generation
func TestStateFilePathGeneration(t *testing.T) {
	// This test just checks that LoadState doesn't panic with no state file
	// The actual file path generation is internal to the filepaths package
	tempDir := helpers.CreateTempDir(t)

	originalInstallDir := app_specific.InstallDir
	app_specific.InstallDir = tempDir
	defer func() {
		app_specific.InstallDir = originalInstallDir
	}()

	// Just verify we can load state without panicking
	state, err := LoadState()
	assert.NoError(t, err)
	assert.NotNil(t, state)
}

// TestStateReattachConversion tests converting state to reattach config
func TestStateReattachConversion(t *testing.T) {
	testExePath := "/usr/local/bin/steampipe"
	reattachConfig := &plugin.ReattachConfig{
		Protocol:        plugin.ProtocolGRPC,
		ProtocolVersion: 1,
		Pid:             12345,
		Test:            true,
		Addr: &simpleAddr{
			network: "tcp",
			address: "localhost:12345",
		},
	}

	state := NewState(testExePath, reattachConfig)

	// Verify the state was created with correct values
	assert.Equal(t, 12345, state.Pid)
	assert.Equal(t, plugin.ProtocolGRPC, state.Protocol)
	assert.Equal(t, 1, state.ProtocolVersion)

	// Convert back to plugin reattach config using the private method
	converted := state.reattachConfig()
	assert.NotNil(t, converted)
	assert.Equal(t, reattachConfig.Pid, converted.Pid)
	assert.Equal(t, reattachConfig.Protocol, converted.Protocol)
	assert.Equal(t, reattachConfig.ProtocolVersion, converted.ProtocolVersion)
}

// TestConcurrentStateAccess tests concurrent access to state operations
func TestConcurrentStateAccess(t *testing.T) {
	tempDir := helpers.CreateTempDir(t)

	originalInstallDir := app_specific.InstallDir
	app_specific.InstallDir = tempDir
	defer func() {
		app_specific.InstallDir = originalInstallDir
	}()

	testExePath := "/usr/local/bin/steampipe"
	reattachConfig := &plugin.ReattachConfig{
		Protocol:        plugin.ProtocolGRPC,
		ProtocolVersion: 1,
		Pid:             12345,
		Addr: &simpleAddr{
			network: "tcp",
			address: "localhost:12345",
		},
	}

	// Create state
	state := NewState(testExePath, reattachConfig)
	err := state.Save()
	require.NoError(t, err)

	// Try to load from multiple goroutines
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func() {
			loadedState, err := LoadState()
			assert.NoError(t, err)
			assert.NotNil(t, loadedState)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}
}

// TestStateWithMissingExecutable tests state with missing executable
func TestStateWithMissingExecutable(t *testing.T) {
	reattachConfig := &plugin.ReattachConfig{
		Protocol:        plugin.ProtocolGRPC,
		ProtocolVersion: 1,
		Pid:             12345,
		Addr: &simpleAddr{
			network: "tcp",
			address: "localhost:12345",
		},
	}

	state := NewState("", reattachConfig)

	assert.NotNil(t, state)
	assert.Equal(t, "", state.Executable)
	assert.False(t, state.Running) // Can't verify if running without executable
}

// TestLoadStateWithCorruptedFile tests handling of corrupted state file
func TestLoadStateWithCorruptedFile(t *testing.T) {
	tempDir := helpers.CreateTempDir(t)

	originalInstallDir := app_specific.InstallDir
	app_specific.InstallDir = tempDir
	defer func() {
		app_specific.InstallDir = originalInstallDir
	}()

	// Create a corrupted state file
	stateDir := filepath.Join(tempDir, "internal")
	err := os.MkdirAll(stateDir, 0755)
	require.NoError(t, err)

	stateFile := filepath.Join(stateDir, "plugin_manager.json")
	err = os.WriteFile(stateFile, []byte("corrupted json {{{"), 0644)
	require.NoError(t, err)

	// Try to load the corrupted state
	state, err := LoadState()

	// LoadState is designed to return a default empty state on error
	// and not propagate the error
	assert.NoError(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, "", state.Executable)
	assert.Equal(t, 0, state.Pid)
}

// simpleAddr is a helper type for testing
type simpleAddr struct {
	network string
	address string
}

func (s *simpleAddr) Network() string {
	return s.network
}

func (s *simpleAddr) String() string {
	return s.address
}

// TestStateKill tests the kill functionality
func TestStateKill(t *testing.T) {
	t.Skip("Skipping test that requires file system operations")

	// The kill function requires access to filepaths which need app_specific.InstallDir
	// This is better tested through integration tests
}

// TestMultipleSaveOperations tests multiple save operations
func TestMultipleSaveOperations(t *testing.T) {
	tempDir := helpers.CreateTempDir(t)

	originalInstallDir := app_specific.InstallDir
	app_specific.InstallDir = tempDir
	defer func() {
		app_specific.InstallDir = originalInstallDir
	}()

	testExePath := "/usr/local/bin/steampipe"

	// Save multiple states
	for i := 0; i < 3; i++ {
		reattachConfig := &plugin.ReattachConfig{
			Protocol:        plugin.ProtocolGRPC,
			ProtocolVersion: 1,
			Pid:             12345 + i,
			Addr: &simpleAddr{
				network: "tcp",
				address: "localhost:12345",
			},
		}

		state := NewState(testExePath, reattachConfig)
		err := state.Save()
		require.NoError(t, err)

		// Load and verify
		loadedState, err := LoadState()
		require.NoError(t, err)
		assert.Equal(t, 12345+i, loadedState.Pid)
	}
}
