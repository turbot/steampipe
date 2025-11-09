package plugin

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/v2/hclhelpers"
)

// TestPluginConnectionInterfaceCompliance tests that all methods are implemented
func TestPluginConnectionInterfaceCompliance(t *testing.T) {
	mock := &MockPluginConnection{
		name:        "test",
		displayName: "Test",
		declRange: hclhelpers.Range{
			Filename: "test.spc",
			Start:    hclhelpers.Pos{Line: 1, Column: 1},
			End:      hclhelpers.Pos{Line: 10, Column: 1},
		},
	}

	// Verify all interface methods
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "test", mock.GetName())
	})

	t.Run("GetDisplayName", func(t *testing.T) {
		assert.Equal(t, "Test", mock.GetDisplayName())
	})

	t.Run("GetDeclRange", func(t *testing.T) {
		declRange := mock.GetDeclRange()
		assert.Equal(t, "test.spc", declRange.Filename)
		assert.Equal(t, 1, declRange.Start.Line)
		assert.Equal(t, 1, declRange.Start.Column)
		assert.Equal(t, 10, declRange.End.Line)
	})
}

// TestMockPluginConnectionEdgeCases tests edge cases
func TestMockPluginConnectionEdgeCases(t *testing.T) {
	tests := map[string]struct {
		mock     *MockPluginConnection
		validate func(*testing.T, *MockPluginConnection)
	}{
		"unicode characters in name": {
			mock: &MockPluginConnection{
				name:        "test_中文_123",
				displayName: "Test 中文 Connection",
			},
			validate: func(t *testing.T, mock *MockPluginConnection) {
				assert.Equal(t, "test_中文_123", mock.GetName())
				assert.Equal(t, "Test 中文 Connection", mock.GetDisplayName())
			},
		},
		"very long names": {
			mock: &MockPluginConnection{
				name:        "very_long_connection_name_with_many_underscores_and_numbers_12345",
				displayName: "Very Long Connection Name With Many Words And Numbers 12345",
			},
			validate: func(t *testing.T, mock *MockPluginConnection) {
				assert.Greater(t, len(mock.GetName()), 50, "name should be long")
				assert.Contains(t, mock.GetName(), "underscores")
				assert.Contains(t, mock.GetDisplayName(), "Words")
			},
		},
		"special characters": {
			mock: &MockPluginConnection{
				name:        "aws-prod-123",
				displayName: "AWS (Production) #123",
			},
			validate: func(t *testing.T, mock *MockPluginConnection) {
				assert.Contains(t, mock.GetName(), "-")
				assert.Contains(t, mock.GetDisplayName(), "(")
				assert.Contains(t, mock.GetDisplayName(), ")")
			},
		},
		"zero line numbers": {
			mock: &MockPluginConnection{
				name:        "test",
				displayName: "Test",
				declRange: hclhelpers.Range{
					Filename: "test.spc",
					Start:    hclhelpers.Pos{Line: 0, Column: 0},
					End:      hclhelpers.Pos{Line: 0, Column: 0},
				},
			},
			validate: func(t *testing.T, mock *MockPluginConnection) {
				assert.Equal(t, 0, mock.GetDeclRange().Start.Line)
				assert.Equal(t, 0, mock.GetDeclRange().End.Line)
			},
		},
		"large line numbers": {
			mock: &MockPluginConnection{
				name:        "test",
				displayName: "Test",
				declRange: hclhelpers.Range{
					Filename: "test.spc",
					Start:    hclhelpers.Pos{Line: 999999, Column: 1},
					End:      hclhelpers.Pos{Line: 1000000, Column: 1},
				},
			},
			validate: func(t *testing.T, mock *MockPluginConnection) {
				assert.Equal(t, 999999, mock.GetDeclRange().Start.Line)
				assert.Equal(t, 1000000, mock.GetDeclRange().End.Line)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.validate(t, tc.mock)
		})
	}
}

// TestMockPluginConnectionFilenames tests various filename patterns
func TestMockPluginConnectionFilenames(t *testing.T) {
	tests := map[string]struct {
		filename string
		validate func(*testing.T, string)
	}{
		"simple filename": {
			filename: "config.spc",
			validate: func(t *testing.T, filename string) {
				assert.Equal(t, "config.spc", filename)
			},
		},
		"relative path": {
			filename: "config/aws.spc",
			validate: func(t *testing.T, filename string) {
				assert.Contains(t, filename, "/")
				assert.True(t, strings.HasSuffix(filename, ".spc"))
			},
		},
		"absolute path": {
			filename: "/home/user/.steampipe/config/aws.spc",
			validate: func(t *testing.T, filename string) {
				assert.True(t, filename[0] == '/')
				assert.Contains(t, filename, ".steampipe")
			},
		},
		"windows path": {
			filename: "C:\\Users\\user\\.steampipe\\config\\aws.spc",
			validate: func(t *testing.T, filename string) {
				assert.Contains(t, filename, "\\")
				assert.Contains(t, filename, "C:")
			},
		},
		"empty filename": {
			filename: "",
			validate: func(t *testing.T, filename string) {
				assert.Equal(t, "", filename)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mock := &MockPluginConnection{
				name:        "test",
				displayName: "Test",
				declRange: hclhelpers.Range{
					Filename: tc.filename,
					Start:    hclhelpers.Pos{Line: 1, Column: 1},
					End:      hclhelpers.Pos{Line: 10, Column: 1},
				},
			}

			tc.validate(t, mock.GetDeclRange().Filename)
		})
	}
}

// TestPluginConnectionInterfaceNilSafety tests nil safety
func TestPluginConnectionInterfaceNilSafety(t *testing.T) {
	// Test with nil mock pointer should panic (expected Go behavior)
	var nilMock *MockPluginConnection = nil

	assert.Panics(t, func() {
		_ = nilMock.GetName()
	}, "calling method on nil pointer should panic")
}

// TestMultiplePluginConnections tests working with multiple connections
func TestMultiplePluginConnections(t *testing.T) {
	connections := []PluginConnection{
		&MockPluginConnection{
			name:        "aws",
			displayName: "AWS",
			declRange: hclhelpers.Range{
				Filename: "config/aws.spc",
				Start:    hclhelpers.Pos{Line: 1, Column: 1},
				End:      hclhelpers.Pos{Line: 5, Column: 1},
			},
		},
		&MockPluginConnection{
			name:        "azure",
			displayName: "Azure",
			declRange: hclhelpers.Range{
				Filename: "config/azure.spc",
				Start:    hclhelpers.Pos{Line: 1, Column: 1},
				End:      hclhelpers.Pos{Line: 5, Column: 1},
			},
		},
		&MockPluginConnection{
			name:        "gcp",
			displayName: "GCP",
			declRange: hclhelpers.Range{
				Filename: "config/gcp.spc",
				Start:    hclhelpers.Pos{Line: 1, Column: 1},
				End:      hclhelpers.Pos{Line: 5, Column: 1},
			},
		},
	}

	assert.Len(t, connections, 3)

	// Verify each connection
	assert.Equal(t, "aws", connections[0].GetName())
	assert.Equal(t, "azure", connections[1].GetName())
	assert.Equal(t, "gcp", connections[2].GetName())

	// Verify all have different filenames
	filenames := make(map[string]bool)
	for _, conn := range connections {
		filenames[conn.GetDeclRange().Filename] = true
	}
	assert.Len(t, filenames, 3)
}
