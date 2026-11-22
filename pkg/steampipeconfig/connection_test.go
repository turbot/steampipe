package steampipeconfig

import (
	"slices"
	"testing"
	"time"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

func TestConnectionsUpdateEqual(t *testing.T) {
	testCases := []struct {
		name     string
		data1    *ConnectionState
		data2    *ConnectionState
		expected bool
	}{
		{
			name: "equal",
			data1: &ConnectionState{
				ConnectionName: "test1",
				Plugin:         "test_plugin",
				State:          "ready",
			},
			data2: &ConnectionState{
				ConnectionName: "test1",
				Plugin:         "test_plugin",
				State:          "ready",
			},
			expected: true,
		},
		{
			name: "different plugin",
			data1: &ConnectionState{
				ConnectionName: "test1",
				Plugin:         "test_plugin",
				State:          "ready",
			},
			data2: &ConnectionState{
				ConnectionName: "test1",
				Plugin:         "different_plugin",
				State:          "ready",
			},
			expected: false,
		},
		{
			name: "different type",
			data1: &ConnectionState{
				ConnectionName: "test1",
				Plugin:         "test_plugin",
				Type:           typehelpers.String("aggregator"),
				State:          "ready",
			},
			data2: &ConnectionState{
				ConnectionName: "test1",
				Plugin:         "test_plugin",
				Type:           nil,
				State:          "ready",
			},
			expected: false,
		},
		{
			name: "different import schema",
			data1: &ConnectionState{
				ConnectionName: "test1",
				Plugin:         "test_plugin",
				ImportSchema:   "enabled",
				State:          "ready",
			},
			data2: &ConnectionState{
				ConnectionName: "test1",
				Plugin:         "test_plugin",
				ImportSchema:   "disabled",
				State:          "ready",
			},
			expected: false,
		},
		{
			name: "different error",
			data1: &ConnectionState{
				ConnectionName:  "test1",
				Plugin:          "test_plugin",
				ConnectionError: typehelpers.String("error1"),
				State:           "error",
			},
			data2: &ConnectionState{
				ConnectionName:  "test1",
				Plugin:          "test_plugin",
				ConnectionError: typehelpers.String("error2"),
				State:           "error",
			},
			expected: false,
		},
		{
			name: "plugin mod time within tolerance",
			data1: &ConnectionState{
				ConnectionName: "test1",
				Plugin:         "test_plugin",
				PluginModTime:  time.Now(),
				State:          "ready",
			},
			data2: &ConnectionState{
				ConnectionName: "test1",
				Plugin:         "test_plugin",
				PluginModTime:  time.Now().Add(500 * time.Microsecond),
				State:          "ready",
			},
			expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.data1.Equals(testCase.data2)
			if result != testCase.expected {
				t.Errorf("Expected %v, got %v", testCase.expected, result)
			}
		})
	}
}

func TestConnectionStateLoaded(t *testing.T) {
	testCases := []struct {
		name     string
		state    *ConnectionState
		expected bool
	}{
		{
			name: "ready state is loaded",
			state: &ConnectionState{
				ConnectionName: "test1",
				State:          constants.ConnectionStateReady,
			},
			expected: true,
		},
		{
			name: "error state is loaded",
			state: &ConnectionState{
				ConnectionName: "test1",
				State:          constants.ConnectionStateError,
			},
			expected: true,
		},
		{
			name: "disabled state is loaded",
			state: &ConnectionState{
				ConnectionName: "test1",
				State:          constants.ConnectionStateDisabled,
			},
			expected: true,
		},
		{
			name: "pending state is not loaded",
			state: &ConnectionState{
				ConnectionName: "test1",
				State:          constants.ConnectionStatePending,
			},
			expected: false,
		},
		{
			name: "updating state is not loaded",
			state: &ConnectionState{
				ConnectionName: "test1",
				State:          constants.ConnectionStateUpdating,
			},
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.state.Loaded()
			if result != testCase.expected {
				t.Errorf("Expected %v, got %v", testCase.expected, result)
			}
		})
	}
}

func TestConnectionStateDisabled(t *testing.T) {
	testCases := []struct {
		name     string
		state    *ConnectionState
		expected bool
	}{
		{
			name: "disabled state",
			state: &ConnectionState{
				ConnectionName: "test1",
				State:          constants.ConnectionStateDisabled,
			},
			expected: true,
		},
		{
			name: "ready state is not disabled",
			state: &ConnectionState{
				ConnectionName: "test1",
				State:          constants.ConnectionStateReady,
			},
			expected: false,
		},
		{
			name: "error state is not disabled",
			state: &ConnectionState{
				ConnectionName: "test1",
				State:          constants.ConnectionStateError,
			},
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.state.Disabled()
			if result != testCase.expected {
				t.Errorf("Expected %v, got %v", testCase.expected, result)
			}
		})
	}
}

func TestConnectionStateGetType(t *testing.T) {
	testCases := []struct {
		name     string
		state    *ConnectionState
		expected string
	}{
		{
			name: "aggregator type",
			state: &ConnectionState{
				ConnectionName: "test1",
				Type:           typehelpers.String("aggregator"),
			},
			expected: "aggregator",
		},
		{
			name: "nil type returns empty string",
			state: &ConnectionState{
				ConnectionName: "test1",
				Type:           nil,
			},
			expected: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.state.GetType()
			if result != testCase.expected {
				t.Errorf("Expected %v, got %v", testCase.expected, result)
			}
		})
	}
}

func TestConnectionStateError(t *testing.T) {
	testCases := []struct {
		name     string
		state    *ConnectionState
		expected string
	}{
		{
			name: "error message",
			state: &ConnectionState{
				ConnectionName:  "test1",
				ConnectionError: typehelpers.String("test error"),
			},
			expected: "test error",
		},
		{
			name: "nil error returns empty string",
			state: &ConnectionState{
				ConnectionName:  "test1",
				ConnectionError: nil,
			},
			expected: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.state.Error()
			if result != testCase.expected {
				t.Errorf("Expected %v, got %v", testCase.expected, result)
			}
		})
	}
}

func TestConnectionStateSetError(t *testing.T) {
	state := &ConnectionState{
		ConnectionName: "test1",
		State:          constants.ConnectionStateReady,
	}

	state.SetError("test error")

	if state.State != constants.ConnectionStateError {
		t.Errorf("Expected state to be %s, got %s", constants.ConnectionStateError, state.State)
	}

	if state.Error() != "test error" {
		t.Errorf("Expected error to be 'test error', got %s", state.Error())
	}
}

func TestConnectionStateEqualsAggregatorChildren(t *testing.T) {
	aggregator := func(children ...string) *ConnectionState {
		aggType := "aggregator"
		return &ConnectionState{
			ConnectionName: "agg",
			Plugin:         "test_plugin",
			Type:           &aggType,
			State:          "ready",
			Connections:    children,
		}
	}
	testCases := []struct {
		name     string
		left     *ConnectionState
		right    *ConnectionState
		expected bool
	}{
		{"two children, same order", aggregator("aws_a", "aws_b"), aggregator("aws_a", "aws_b"), true},
		{"two children, different order", aggregator("aws_b", "aws_a"), aggregator("aws_a", "aws_b"), true},
		{"three children, same", aggregator("c", "a", "b"), aggregator("b", "c", "a"), true},
		{"different child", aggregator("aws_a", "aws_b"), aggregator("aws_a", "aws_c"), false},
		{"extra child", aggregator("aws_a", "aws_b"), aggregator("aws_a", "aws_b", "aws_c"), false},
		{"one child, same", aggregator("aws_a"), aggregator("aws_a"), true},
		{"one child, different", aggregator("aws_a"), aggregator("aws_b"), false},
		{"no children, nil and empty", aggregator(), &ConnectionState{ConnectionName: "agg", Plugin: "test_plugin", State: "ready", Type: aggregator().Type, Connections: []string{}}, true},
		{"none versus one", aggregator(), aggregator("aws_a"), false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.left.Equals(tc.right); got != tc.expected {
				t.Errorf("Equals = %v, want %v", got, tc.expected)
			}
			if got := tc.right.Equals(tc.left); got != tc.expected {
				t.Errorf("reversed Equals = %v, want %v", got, tc.expected)
			}
		})
	}
}

// Equals must not reorder the Connections of either state
func TestConnectionStateEqualsKeepsChildOrder(t *testing.T) {
	left := &ConnectionState{Plugin: "p", Connections: []string{"c", "a", "b"}}
	right := &ConnectionState{Plugin: "p", Connections: []string{"b", "c", "a"}}
	if !left.Equals(right) {
		t.Fatal("expected states with the same children to be equal")
	}
	if got := left.Connections; !slices.Equal(got, []string{"c", "a", "b"}) {
		t.Errorf("left Connections reordered: %v", got)
	}
	if got := right.Connections; !slices.Equal(got, []string{"b", "c", "a"}) {
		t.Errorf("right Connections reordered: %v", got)
	}
}

// an unchanged aggregator with two or more children must not need an update
func TestConnectionRequiresUpdateUnchangedAggregator(t *testing.T) {
	aggType := "aggregator"
	modTime := time.Now()
	newState := func(children ...string) *ConnectionState {
		return &ConnectionState{
			ConnectionName: "agg",
			Plugin:         "test_plugin",
			Type:           &aggType,
			State:          constants.ConnectionStateReady,
			PluginModTime:  modTime,
			Connections:    children,
		}
	}
	current := ConnectionStateMap{"agg": newState("aws_a", "aws_b")}

	if res := connectionRequiresUpdate(nil, "agg", current, newState("aws_b", "aws_a")); res.requiresUpdate {
		t.Error("unchanged aggregator with two children should not require an update")
	}
	if res := connectionRequiresUpdate(nil, "agg", current, newState("aws_a", "aws_c")); !res.requiresUpdate {
		t.Error("aggregator with a changed child should require an update")
	}
}
