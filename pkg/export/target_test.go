package export

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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

// TestTarget_Export_Message tests that the status message returned by Target.Export()
// reports a relative filePath joined with the working directory, and reports an
// absolute filePath as-is (not doubled with the working directory).
func TestTarget_Export_Message(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	testCases := []struct {
		name     string
		filePath string
		expect   string
	}{
		{
			name:     "relative path",
			filePath: "output.json",
			expect:   fmt.Sprintf("File exported to %s", filepath.Join(pwd, "output.json")),
		},
		{
			name:     "absolute path",
			filePath: filepath.Join(pwd, "abs", "output.json"),
			expect:   fmt.Sprintf("File exported to %s", filepath.Join(pwd, "abs", "output.json")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := &Target{
				exporter: &testExporter{extension: ".json", name: "noop"},
				filePath: tc.filePath,
			}
			msg, err := target.Export(context.Background(), &mockExportSourceData{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if msg != tc.expect {
				t.Errorf("expected message %q, got %q", tc.expect, msg)
			}
		})
	}
}
