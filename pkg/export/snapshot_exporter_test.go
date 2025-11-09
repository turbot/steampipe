package export

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/steampipe/v2/pkg/test/helpers"
)

// TestSnapshotExporter_ExportInvalidInput tests that Export returns error for invalid input
func TestSnapshotExporter_ExportInvalidInput(t *testing.T) {
	exporter := &SnapshotExporter{}
	tempDir := helpers.CreateTempDir(t)
	outputPath := filepath.Join(tempDir, "output.sps")

	// Try to export with wrong type
	invalidInput := &testExportSourceData{}
	err := exporter.Export(context.Background(), invalidInput, outputPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SnapshotExporter input must be")
}

// TestTargetExport tests the Target Export method
func TestTargetExport(t *testing.T) {
	tests := map[string]struct {
		exporter    Exporter
		filePath    string
		expectError bool
	}{
		"successful export": {
			exporter: &testExporter{
				name:      "test",
				extension: ".test",
			},
			filePath:    "output.test",
			expectError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tempDir := helpers.CreateTempDir(t)
			_ = filepath.Join(tempDir, tc.filePath)

			// Change to temp dir to test relative path handling
			oldWd, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(oldWd)

			target := &Target{
				exporter: tc.exporter,
				filePath: tc.filePath,
			}

			testData := &testExportSourceData{}
			msg, err := target.Export(context.Background(), testData)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, msg, "File exported to")
				assert.Contains(t, msg, tc.filePath)
			}
		})
	}
}

// TestTargetExport_NamedVsUnnamed tests isNamedTarget flag
func TestTargetExport_NamedVsUnnamed(t *testing.T) {
	m := NewManager()
	m.Register(&dummyJSONExporter)

	// Named target (file path)
	namedTarget, err := m.getExportTarget("output.json", "test")
	assert.NoError(t, err)
	assert.True(t, namedTarget.isNamedTarget)
	assert.Equal(t, "output.json", namedTarget.filePath)

	// Unnamed target (format name)
	unnamedTarget, err := m.getExportTarget("json", "test")
	assert.NoError(t, err)
	assert.False(t, unnamedTarget.isNamedTarget)
	// Filename should be auto-generated with timestamp
	assert.Contains(t, unnamedTarget.filePath, "test")
	assert.Contains(t, unnamedTarget.filePath, ".json")
}

// TestWrite_Helper tests the Write helper function
func TestWrite_Helper(t *testing.T) {
	// This is tested indirectly through the export tests, but we can add specific tests
	// Testing is done in helpers_test.go
}
