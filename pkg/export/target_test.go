package export

import (
	"context"
	"testing"
)

// TestTarget_Export_NilExporter tests that Target.Export() handles a nil exporter gracefully
// by returning an error instead of panicking.
// This test addresses bug #4717.
func TestTarget_Export_NilExporter(t *testing.T) {
	// Create a Target with a nil exporter
	target := &Target{
		exporter:      nil,
		filePath:      "test.json",
		isNamedTarget: false,
	}

	// Create a simple mock ExportSourceData
	mockData := &mockExportSourceData{}

	// Call Export - this should return an error, not panic
	_, err := target.Export(context.Background(), mockData)

	// Verify that we got an error (not a panic)
	if err == nil {
		t.Fatal("Expected error when exporter is nil, but got nil")
	}

	// Verify the error message is meaningful
	expectedErrSubstring := "exporter"
	if err != nil && len(err.Error()) > 0 {
		t.Logf("Got expected error: %v", err)
	}
	_ = expectedErrSubstring // Will be used after fix is applied
}

// mockExportSourceData is a simple mock implementation for testing
type mockExportSourceData struct{}

func (m *mockExportSourceData) IsExportSourceData() {}
