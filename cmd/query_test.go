package cmd

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/steampipe/v2/pkg/constants"
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

// TestValidateQueryArgs_ConcurrentCalls tests that validateQueryArgs is thread-safe
// Bug #4706: validateQueryArgs uses global viper state which is not thread-safe
func TestValidateQueryArgs_ConcurrentCalls(t *testing.T) {
	ctx := context.Background()
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Run 100 concurrent calls to validateQueryArgs
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()

			// Create config struct - this is now thread-safe
			// Each goroutine has its own config instance
			cfg := &queryConfig{
				snapshot: false,
				share:    false,
				export:   []string{},
				output:   constants.OutputFormatTable,
			}

			// Call validateQueryArgs with a query argument (non-interactive mode)
			err := validateQueryArgs(ctx, []string{"SELECT 1"}, cfg)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check if any errors occurred
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	// The test should not panic or produce errors
	assert.Empty(t, errs, "validateQueryArgs should handle concurrent calls without errors")
}

// TestValidateQueryArgs_InteractiveModeWithSnapshot tests validation in interactive mode with snapshot
func TestValidateQueryArgs_InteractiveModeWithSnapshot(t *testing.T) {
	ctx := context.Background()

	// Setup config with snapshot enabled
	cfg := &queryConfig{
		snapshot: true,
		share:    false,
		export:   []string{},
		output:   constants.OutputFormatTable,
	}

	// Call with no args (interactive mode)
	err := validateQueryArgs(ctx, []string{}, cfg)

	// Should return error for snapshot in interactive mode
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot share snapshots in interactive mode")
}

// TestValidateQueryArgs_BatchModeWithSnapshot tests validation in batch mode with snapshot
func TestValidateQueryArgs_BatchModeWithSnapshot(t *testing.T) {
	ctx := context.Background()

	// Setup config with snapshot enabled
	cfg := &queryConfig{
		snapshot: true,
		share:    false,
		export:   []string{},
		output:   constants.OutputFormatTable,
	}

	// Call with args (batch mode)
	err := validateQueryArgs(ctx, []string{"SELECT 1"}, cfg)

	// Should not return error for snapshot in batch mode
	// (unless there are other validation errors from cmdconfig.ValidateSnapshotArgs)
	// For this test, we expect it to pass basic validation
	if err != nil {
		// If there's an error, it should not be about interactive mode
		assert.NotContains(t, err.Error(), "cannot share snapshots in interactive mode")
	}
}

// TestValidateQueryArgs_InvalidOutputFormat tests validation with invalid output format
func TestValidateQueryArgs_InvalidOutputFormat(t *testing.T) {
	ctx := context.Background()

	// Setup config with invalid output format
	cfg := &queryConfig{
		snapshot: false,
		share:    false,
		export:   []string{},
		output:   "invalid-format",
	}

	// Call with args (batch mode)
	err := validateQueryArgs(ctx, []string{"SELECT 1"}, cfg)

	// Should return error for invalid output format
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
}
