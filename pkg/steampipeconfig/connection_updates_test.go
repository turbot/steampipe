package steampipeconfig

import (
	"testing"

	"github.com/turbot/steampipe/v2/pkg/constants"
)

// TestConnectionUpdates_IdentifyMissingComments tests the logic error in IdentifyMissingComments
// Bug #4814: The function uses OR (||) when it should use AND (&&) on line 426
// Current buggy logic: if !updating || deleting
// This means connections being DELETED are still added to MissingComments
// Expected logic: if !updating && !deleting
func TestConnectionUpdates_IdentifyMissingComments(t *testing.T) {
	tests := []struct {
		name                  string
		connectionName        string
		currentState          *ConnectionState
		finalState            *ConnectionState
		isUpdating            bool
		isDeleting            bool
		shouldBeMissing       bool
		description           string
	}{
		{
			name:           "connection being deleted should NOT be in MissingComments",
			connectionName: "conn1",
			currentState: &ConnectionState{
				ConnectionName: "conn1",
				Plugin:         "test_plugin",
				State:          constants.ConnectionStateReady,
				CommentsSet:    false, // Comments not set
			},
			finalState: &ConnectionState{
				ConnectionName: "conn1",
				Plugin:         "test_plugin",
				State:          constants.ConnectionStateReady,
			},
			isUpdating:      false,
			isDeleting:      true, // Being deleted
			shouldBeMissing: false, // Should NOT be in MissingComments (but bug adds it)
			description:     "Deleting connections should be ignored",
		},
		{
			name:           "connection being updated should NOT be in MissingComments",
			connectionName: "conn2",
			currentState: &ConnectionState{
				ConnectionName: "conn2",
				Plugin:         "test_plugin",
				State:          constants.ConnectionStateReady,
				CommentsSet:    false,
			},
			finalState: &ConnectionState{
				ConnectionName: "conn2",
				Plugin:         "test_plugin",
				State:          constants.ConnectionStateReady,
			},
			isUpdating:      true, // Being updated
			isDeleting:      false,
			shouldBeMissing: false, // Should NOT be in MissingComments
			description:     "Updating connections should be ignored",
		},
		{
			name:           "stable connection without comments SHOULD be in MissingComments",
			connectionName: "conn3",
			currentState: &ConnectionState{
				ConnectionName: "conn3",
				Plugin:         "test_plugin",
				State:          constants.ConnectionStateReady,
				CommentsSet:    false, // Comments not set
			},
			finalState: &ConnectionState{
				ConnectionName: "conn3",
				Plugin:         "test_plugin",
				State:          constants.ConnectionStateReady,
			},
			isUpdating:      false, // Not being updated
			isDeleting:      false, // Not being deleted
			shouldBeMissing: true,  // SHOULD be in MissingComments
			description:     "Stable connections without comments should be identified",
		},
		{
			name:           "connection with comments set should NOT be in MissingComments",
			connectionName: "conn4",
			currentState: &ConnectionState{
				ConnectionName: "conn4",
				Plugin:         "test_plugin",
				State:          constants.ConnectionStateReady,
				CommentsSet:    true, // Comments ARE set
			},
			finalState: &ConnectionState{
				ConnectionName: "conn4",
				Plugin:         "test_plugin",
				State:          constants.ConnectionStateReady,
			},
			isUpdating:      false,
			isDeleting:      false,
			shouldBeMissing: false, // Should NOT be in MissingComments
			description:     "Connections with comments set should be ignored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ConnectionUpdates with the test scenario
			updates := &ConnectionUpdates{
				Update:                 make(ConnectionStateMap),
				Delete:                 make(map[string]struct{}),
				MissingComments:        make(ConnectionStateMap),
				CurrentConnectionState: make(ConnectionStateMap),
				FinalConnectionState:   make(ConnectionStateMap),
			}

			// Set up current and final state
			updates.CurrentConnectionState[tt.connectionName] = tt.currentState
			updates.FinalConnectionState[tt.connectionName] = tt.finalState

			// Set up updating/deleting status
			if tt.isUpdating {
				updates.Update[tt.connectionName] = tt.finalState
			}
			if tt.isDeleting {
				updates.Delete[tt.connectionName] = struct{}{}
			}

			// Call the function under test
			updates.IdentifyMissingComments()

			// Check if the connection is in MissingComments
			_, inMissingComments := updates.MissingComments[tt.connectionName]

			if tt.shouldBeMissing != inMissingComments {
				t.Errorf("%s: expected shouldBeMissing=%v, got inMissingComments=%v",
					tt.description, tt.shouldBeMissing, inMissingComments)
			}
		})
	}
}

// A connection whose plugin block was skipped is kept only if its state shows it has a schema:
// taking it out of both maps is what stops the refresh updating or deleting it. A connection
// that cannot be queried is kept in error, with the reason it is in error.
func TestTakeConnectionsToKeep(t *testing.T) {
	tests := []struct {
		name          string
		currentState  string
		requiredState string
		wantKept      bool
		wantState     string
	}{
		{name: "ready", currentState: constants.ConnectionStateReady, requiredState: constants.ConnectionStateError, wantKept: true, wantState: constants.ConnectionStateError},
		{name: "pending_after_startup", currentState: constants.ConnectionStatePending, requiredState: constants.ConnectionStateError, wantKept: true, wantState: constants.ConnectionStateError},
		{name: "resolves_despite_skipped_block", currentState: constants.ConnectionStateReady, requiredState: constants.ConnectionStateReady, wantKept: true, wantState: constants.ConnectionStateReady},
		// a connection kept in error is set to pending-incomplete when the service starts
		{name: "kept_in_error_before_restart", currentState: constants.ConnectionStateError, requiredState: constants.ConnectionStateError, wantKept: true, wantState: constants.ConnectionStateError},
		{name: "kept_in_error_after_restart", currentState: constants.ConnectionStatePendingIncomplete, requiredState: constants.ConnectionStateError, wantKept: true, wantState: constants.ConnectionStateError},
		{name: "error_but_resolves_now", currentState: constants.ConnectionStateError, requiredState: constants.ConnectionStateReady},
		{name: "pending_incomplete_but_resolves_now", currentState: constants.ConnectionStatePendingIncomplete, requiredState: constants.ConnectionStateReady},
		{name: "disabled", currentState: constants.ConnectionStateDisabled, requiredState: constants.ConnectionStateError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := "plugin block 'plugin.held' failed to load"
			required := ConnectionStateMap{
				"held":  {ConnectionName: "held", State: tt.requiredState, ConnectionError: &reason},
				"other": {ConnectionName: "other", State: constants.ConnectionStateReady},
			}
			current := ConnectionStateMap{
				"held":  {ConnectionName: "held", State: tt.currentState, SchemaHash: "hash"},
				"other": {ConnectionName: "other", State: constants.ConnectionStateReady},
			}

			kept := takeConnectionsToKeep(map[string]struct{}{"held": {}, "new": {}}, required, current)

			state, gotKept := kept["held"]
			if gotKept != tt.wantKept {
				t.Fatalf("kept = %v, want %v", gotKept, tt.wantKept)
			}
			_, inRequired := required["held"]
			_, inCurrent := current["held"]
			if inRequired == tt.wantKept || inCurrent == tt.wantKept {
				t.Errorf("in required = %v, in current = %v, want both %v", inRequired, inCurrent, !tt.wantKept)
			}
			if tt.wantKept && (state.State != tt.wantState || state.SchemaHash != "hash") {
				t.Errorf("kept state = %q with hash %q, want %q with the current hash", state.State, state.SchemaHash, tt.wantState)
			}
			if tt.wantKept && tt.wantState == constants.ConnectionStateError && state.Error() != reason {
				t.Errorf("kept error = %q, want %q", state.Error(), reason)
			}
			if _, ok := kept["new"]; ok {
				t.Error("a connection with no current state has no schema to keep")
			}
			if _, ok := required["other"]; !ok {
				t.Error("an unrelated connection must stay in the comparison")
			}
		})
	}
}
