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
		Addr:            nil, // Nil address - this will cause panic without fix
	}

	// This should not panic - it should return nil gracefully
	config := state.reattachConfig()

	// With nil Addr, we expect nil config (not a panic)
	if config != nil {
		t.Error("Expected nil reattach config when Addr is nil")
	}
}
