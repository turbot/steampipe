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

// A connection whose plugin block was skipped keeps its schema only if it has one:
// taking it out of both maps is what stops the refresh updating or deleting it.
func TestTakeConnectionsToKeep(t *testing.T) {
	tests := []struct {
		name         string
		currentState string
		wantKept     bool
	}{
		{name: "ready", currentState: constants.ConnectionStateReady, wantKept: true},
		{name: "pending_after_startup", currentState: constants.ConnectionStatePending, wantKept: true},
		{name: "already_in_error", currentState: constants.ConnectionStateError},
		{name: "pending_incomplete", currentState: constants.ConnectionStatePendingIncomplete},
		{name: "disabled", currentState: constants.ConnectionStateDisabled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			required := ConnectionStateMap{
				"held":  {ConnectionName: "held", State: constants.ConnectionStateError},
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
			if tt.wantKept && (state.State != constants.ConnectionStateReady || state.SchemaHash != "hash") {
				t.Errorf("kept state = %q with hash %q, want ready with the current hash", state.State, state.SchemaHash)
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
