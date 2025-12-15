package pluginmanager

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/hashicorp/go-plugin"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
	pb "github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
)

// TestStateWithNilAddr tests that reattachConfig handles nil Addr gracefully
// This test demonstrates bug #4755
func TestStateWithNilAddr(t *testing.T) {
	state := &State{
		Protocol:        plugin.ProtocolGRPC,
		ProtocolVersion: 1,
		Pid:             12345,
		Executable:      "/usr/local/bin/steampipe",
		Addr:            nil, // Nil address - this will cause panic without fix
	}

	// This should not panic - it should return nil gracefully
	config := state.reattachConfig()

	// With nil Addr, we expect nil config (not a panic)
	if config != nil {
		t.Error("Expected nil reattach config when Addr is nil")
	}
}

func TestStateFileRaceCondition(t *testing.T) {
	// This test demonstrates the race condition in State.Save()
	// When multiple goroutines call Save() concurrently, they can corrupt the JSON file

	// Setup: Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "steampipe-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize app_specific.InstallDir for the test
	app_specific.InstallDir = filepath.Join(tempDir, ".steampipe")

	// Create multiple states with different data
	concurrency := 50
	iterations := 20
	var wg sync.WaitGroup
	wg.Add(concurrency)

	// Channel to collect errors from goroutines
	errors := make(chan error, concurrency*iterations)

	// Launch concurrent Save() operations to the same file
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()

			// Create a new state with unique data
			addr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080 + id}
			reattach := &plugin.ReattachConfig{
				Protocol:        plugin.ProtocolGRPC,
				ProtocolVersion: 1,
				Addr:            pb.NewSimpleAddr(addr),
				Pid:             1000 + id,
			}

			state := NewState("/test/executable", reattach)

			// Perform multiple saves to increase race window
			for j := 0; j < iterations; j++ {
				if err := state.Save(); err != nil {
					errors <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors during save
	for err := range errors {
		t.Errorf("Failed to save state: %v", err)
	}

	// Verify that the state file is valid JSON
	stateFilePath := filepaths.PluginManagerStateFilePath()
	content, err := os.ReadFile(stateFilePath)
	if err != nil {
		t.Fatalf("Failed to read state file: %v", err)
	}

	// The main test: Can we unmarshal the file without error?
	var state State
	err = json.Unmarshal(content, &state)
	if err != nil {
		t.Fatalf("State file is corrupted (invalid JSON): %v\nContent: %s", err, string(content))
	}

	// Additional validation: ensure required fields are present
	if state.StructVersion != PluginManagerStructVersion {
		t.Errorf("State file missing or has incorrect struct version: got %d, want %d",
			state.StructVersion, PluginManagerStructVersion)
	}
}
