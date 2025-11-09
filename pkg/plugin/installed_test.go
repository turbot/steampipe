package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPluginConnectionInterface tests that the PluginConnection interface is properly defined
func TestPluginConnectionInterface(t *testing.T) {
	// Test that our mock properly implements the interface
	var conn PluginConnection = &MockPluginConnection{
		name:        "test",
		displayName: "Test Plugin",
	}

	assert.NotNil(t, conn)
	assert.Equal(t, "test", conn.GetName())
	assert.Equal(t, "Test Plugin", conn.GetDisplayName())
	assert.NotNil(t, conn.GetDeclRange())
}

// TestPluginConnectionMock verifies mock behavior
func TestPluginConnectionMock(t *testing.T) {
	tests := map[string]struct {
		mock     *MockPluginConnection
		validate func(*testing.T, *MockPluginConnection)
	}{
		"basic mock": {
			mock: &MockPluginConnection{
				name:        "aws",
				displayName: "AWS",
			},
			validate: func(t *testing.T, mock *MockPluginConnection) {
				assert.Equal(t, "aws", mock.GetName())
				assert.Equal(t, "AWS", mock.GetDisplayName())
			},
		},
		"mock with special characters": {
			mock: &MockPluginConnection{
				name:        "aws_prod_123",
				displayName: "AWS Production (123)",
			},
			validate: func(t *testing.T, mock *MockPluginConnection) {
				assert.Equal(t, "aws_prod_123", mock.GetName())
				assert.Equal(t, "AWS Production (123)", mock.GetDisplayName())
			},
		},
		"mock with empty values": {
			mock: &MockPluginConnection{
				name:        "",
				displayName: "",
			},
			validate: func(t *testing.T, mock *MockPluginConnection) {
				assert.Equal(t, "", mock.GetName())
				assert.Equal(t, "", mock.GetDisplayName())
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.validate(t, tc.mock)
		})
	}
}
