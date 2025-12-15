package steampipeconfig

import (
	"testing"
	"time"

	"github.com/turbot/steampipe/v2/pkg/constants"
)

func TestConnectionStateMapGetSummary(t *testing.T) {
	stateMap := ConnectionStateMap{
		"conn1": &ConnectionState{
			ConnectionName: "conn1",
			State:          constants.ConnectionStateReady,
		},
		"conn2": &ConnectionState{
			ConnectionName: "conn2",
			State:          constants.ConnectionStateReady,
		},
		"conn3": &ConnectionState{
			ConnectionName: "conn3",
			State:          constants.ConnectionStateError,
		},
		"conn4": &ConnectionState{
			ConnectionName: "conn4",
			State:          constants.ConnectionStatePending,
		},
	}

	summary := stateMap.GetSummary()

	if summary[constants.ConnectionStateReady] != 2 {
		t.Errorf("Expected 2 ready connections, got %d", summary[constants.ConnectionStateReady])
	}

	if summary[constants.ConnectionStateError] != 1 {
		t.Errorf("Expected 1 error connection, got %d", summary[constants.ConnectionStateError])
	}

	if summary[constants.ConnectionStatePending] != 1 {
		t.Errorf("Expected 1 pending connection, got %d", summary[constants.ConnectionStatePending])
	}
}

func TestConnectionStateMapPending(t *testing.T) {
	testCases := []struct {
		name     string
		stateMap ConnectionStateMap
		expected bool
	}{
		{
			name: "has pending connections",
			stateMap: ConnectionStateMap{
				"conn1": &ConnectionState{
					ConnectionName: "conn1",
					State:          constants.ConnectionStatePending,
				},
				"conn2": &ConnectionState{
					ConnectionName: "conn2",
					State:          constants.ConnectionStateReady,
				},
			},
			expected: true,
		},
		{
			name: "has pending incomplete connections",
			stateMap: ConnectionStateMap{
				"conn1": &ConnectionState{
					ConnectionName: "conn1",
					State:          constants.ConnectionStatePendingIncomplete,
				},
			},
			expected: true,
		},
		{
			name: "no pending connections",
			stateMap: ConnectionStateMap{
				"conn1": &ConnectionState{
					ConnectionName: "conn1",
					State:          constants.ConnectionStateReady,
				},
				"conn2": &ConnectionState{
					ConnectionName: "conn2",
					State:          constants.ConnectionStateError,
				},
			},
			expected: false,
		},
		{
			name:     "empty map",
			stateMap: ConnectionStateMap{},
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.stateMap.Pending()
			if result != testCase.expected {
				t.Errorf("Expected %v, got %v", testCase.expected, result)
			}
		})
	}
}

func TestConnectionStateMapLoaded(t *testing.T) {
	testCases := []struct {
		name        string
		stateMap    ConnectionStateMap
		connections []string
		expected    bool
	}{
		{
			name: "all connections loaded",
			stateMap: ConnectionStateMap{
				"conn1": &ConnectionState{
					ConnectionName: "conn1",
					State:          constants.ConnectionStateReady,
				},
				"conn2": &ConnectionState{
					ConnectionName: "conn2",
					State:          constants.ConnectionStateError,
				},
			},
			connections: []string{},
			expected:    true,
		},
		{
			name: "some connections not loaded",
			stateMap: ConnectionStateMap{
				"conn1": &ConnectionState{
					ConnectionName: "conn1",
					State:          constants.ConnectionStateReady,
				},
				"conn2": &ConnectionState{
					ConnectionName: "conn2",
					State:          constants.ConnectionStatePending,
				},
			},
			connections: []string{},
			expected:    false,
		},
		{
			name: "specific connections loaded",
			stateMap: ConnectionStateMap{
				"conn1": &ConnectionState{
					ConnectionName: "conn1",
					State:          constants.ConnectionStateReady,
				},
				"conn2": &ConnectionState{
					ConnectionName: "conn2",
					State:          constants.ConnectionStatePending,
				},
			},
			connections: []string{"conn1"},
			expected:    true,
		},
		{
			name: "disabled connections are loaded",
			stateMap: ConnectionStateMap{
				"conn1": &ConnectionState{
					ConnectionName: "conn1",
					State:          constants.ConnectionStateDisabled,
				},
			},
			connections: []string{},
			expected:    true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.stateMap.Loaded(testCase.connections...)
			if result != testCase.expected {
				t.Errorf("Expected %v, got %v", testCase.expected, result)
			}
		})
	}
}

func TestConnectionStateMapConnectionsInState(t *testing.T) {
	stateMap := ConnectionStateMap{
		"conn1": &ConnectionState{
			ConnectionName: "conn1",
			State:          constants.ConnectionStateReady,
		},
		"conn2": &ConnectionState{
			ConnectionName: "conn2",
			State:          constants.ConnectionStateError,
		},
		"conn3": &ConnectionState{
			ConnectionName: "conn3",
			State:          constants.ConnectionStatePending,
		},
	}

	testCases := []struct {
		name     string
		states   []string
		expected bool
	}{
		{
			name:     "has ready connections",
			states:   []string{constants.ConnectionStateReady},
			expected: true,
		},
		{
			name:     "has error or pending connections",
			states:   []string{constants.ConnectionStateError, constants.ConnectionStatePending},
			expected: true,
		},
		{
			name:     "no updating connections",
			states:   []string{constants.ConnectionStateUpdating},
			expected: false,
		},
		{
			name:     "no deleting connections",
			states:   []string{constants.ConnectionStateDeleting},
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := stateMap.ConnectionsInState(testCase.states...)
			if result != testCase.expected {
				t.Errorf("Expected %v, got %v", testCase.expected, result)
			}
		})
	}
}

func TestConnectionStateMapEquals(t *testing.T) {
	testCases := []struct {
		name     string
		map1     ConnectionStateMap
		map2     ConnectionStateMap
		expected bool
	}{
		{
			name: "equal maps",
			map1: ConnectionStateMap{
				"conn1": &ConnectionState{
					ConnectionName: "conn1",
					Plugin:         "plugin1",
					State:          constants.ConnectionStateReady,
				},
			},
			map2: ConnectionStateMap{
				"conn1": &ConnectionState{
					ConnectionName: "conn1",
					Plugin:         "plugin1",
					State:          constants.ConnectionStateReady,
				},
			},
			expected: true,
		},
		{
			name: "different plugins",
			map1: ConnectionStateMap{
				"conn1": &ConnectionState{
					ConnectionName: "conn1",
					Plugin:         "plugin1",
					State:          constants.ConnectionStateReady,
				},
			},
			map2: ConnectionStateMap{
				"conn1": &ConnectionState{
					ConnectionName: "conn1",
					Plugin:         "plugin2",
					State:          constants.ConnectionStateReady,
				},
			},
			expected: false,
		},
		{
			name: "different keys",
			map1: ConnectionStateMap{
				"conn1": &ConnectionState{
					ConnectionName: "conn1",
					Plugin:         "plugin1",
					State:          constants.ConnectionStateReady,
				},
			},
			map2: ConnectionStateMap{
				"conn2": &ConnectionState{
					ConnectionName: "conn2",
					Plugin:         "plugin1",
					State:          constants.ConnectionStateReady,
				},
			},
			expected: false,
		},
		{
			name: "nil vs non-nil",
			map1: nil,
			map2: ConnectionStateMap{
				"conn1": &ConnectionState{
					ConnectionName: "conn1",
					Plugin:         "plugin1",
					State:          constants.ConnectionStateReady,
				},
			},
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.map1.Equals(testCase.map2)
			if result != testCase.expected {
				t.Errorf("Expected %v, got %v", testCase.expected, result)
			}
		})
	}
}

func TestConnectionStateMapConnectionModTime(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-1 * time.Hour)
	later := now.Add(1 * time.Hour)

	stateMap := ConnectionStateMap{
		"conn1": &ConnectionState{
			ConnectionName:    "conn1",
			ConnectionModTime: earlier,
		},
		"conn2": &ConnectionState{
			ConnectionName:    "conn2",
			ConnectionModTime: later,
		},
		"conn3": &ConnectionState{
			ConnectionName:    "conn3",
			ConnectionModTime: now,
		},
	}

	result := stateMap.ConnectionModTime()

	if !result.Equal(later) {
		t.Errorf("Expected latest mod time %v, got %v", later, result)
	}
}

func TestConnectionStateMapConnectionModTimeEmpty(t *testing.T) {
	stateMap := ConnectionStateMap{}

	result := stateMap.ConnectionModTime()

	if !result.IsZero() {
		t.Errorf("Expected zero time for empty map, got %v", result)
	}
}

func TestConnectionStateMapGetPluginToConnectionMap(t *testing.T) {
	stateMap := ConnectionStateMap{
		"conn1": &ConnectionState{
			ConnectionName: "conn1",
			Plugin:         "plugin1",
		},
		"conn2": &ConnectionState{
			ConnectionName: "conn2",
			Plugin:         "plugin1",
		},
		"conn3": &ConnectionState{
			ConnectionName: "conn3",
			Plugin:         "plugin2",
		},
	}

	result := stateMap.GetPluginToConnectionMap()

	if len(result["plugin1"]) != 2 {
		t.Errorf("Expected 2 connections for plugin1, got %d", len(result["plugin1"]))
	}

	if len(result["plugin2"]) != 1 {
		t.Errorf("Expected 1 connection for plugin2, got %d", len(result["plugin2"]))
	}
}

func TestConnectionStateMapSetConnectionsToPendingOrIncomplete(t *testing.T) {
	stateMap := ConnectionStateMap{
		"conn1": &ConnectionState{
			ConnectionName: "conn1",
			State:          constants.ConnectionStateReady,
		},
		"conn2": &ConnectionState{
			ConnectionName: "conn2",
			State:          constants.ConnectionStateError,
		},
		"conn3": &ConnectionState{
			ConnectionName: "conn3",
			State:          constants.ConnectionStateDisabled,
		},
	}

	stateMap.SetConnectionsToPendingOrIncomplete()

	if stateMap["conn1"].State != constants.ConnectionStatePending {
		t.Errorf("Expected conn1 to be pending, got %s", stateMap["conn1"].State)
	}

	if stateMap["conn2"].State != constants.ConnectionStatePendingIncomplete {
		t.Errorf("Expected conn2 to be pending incomplete, got %s", stateMap["conn2"].State)
	}

	if stateMap["conn3"].State != constants.ConnectionStateDisabled {
		t.Errorf("Expected conn3 to remain disabled, got %s", stateMap["conn3"].State)
	}
}
