package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestGetPipedStdinData_PreservesNewlines(t *testing.T) {
	// Save original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a temporary file to simulate piped input
	tmpFile, err := os.CreateTemp("", "stdin-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Test input with multiple lines - matching the bug report example
	testInput := "SELECT * FROM aws_account\nWHERE account_id = '123'\nAND region = 'us-east-1';"

	// Write test input to the temp file
	if _, err := tmpFile.WriteString(testInput); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Seek back to the beginning
	if _, err := tmpFile.Seek(0, 0); err != nil {
		t.Fatalf("Failed to seek temp file: %v", err)
	}

	// Replace stdin with our temp file
	os.Stdin = tmpFile

	// Call the function
	result := getPipedStdinData()

	// Clean up
	tmpFile.Close()

	// Verify that newlines are preserved
	if result != testInput {
		t.Errorf("getPipedStdinData() did not preserve newlines\nExpected: %q\nGot: %q", testInput, result)

		// Show the difference more clearly
		expectedLines := strings.Split(testInput, "\n")
		resultLines := strings.Split(result, "\n")
		t.Logf("Expected %d lines, got %d lines", len(expectedLines), len(resultLines))
		t.Logf("Expected lines: %v", expectedLines)
		t.Logf("Got lines: %v", resultLines)
	}
}
