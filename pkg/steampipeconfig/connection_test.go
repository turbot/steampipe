package steampipeconfig

import (
	"testing"
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
