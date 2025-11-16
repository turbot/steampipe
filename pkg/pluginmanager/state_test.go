package pluginmanager

import (
	"testing"

	"github.com/hashicorp/go-plugin"
)

// TestStateWithNilAddr tests that reattachConfig handles nil Addr gracefully
// This test demonstrates bug #4755
func TestStateWithNilAddr(t *testing.T) {
	state := &State{
		Protocol:        plugin.ProtocolGRPC,
		ProtocolVersion: 1,
		Pid:             12345,
		Executable:      "/usr/local/bin/steampipe",
		Addr:            nil, // Nil address - this will cause panic
	}

	// This should not panic
	config := state.reattachConfig()

	// If we reach here without panic, the bug is fixed
	if config == nil {
		t.Error("Expected non-nil reattach config")
	}
}
