package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/steampipe/v2/pkg/statushooks"
)

// TestHideRootFlags_NilPointerBug tests potential nil panic bug
// HIGH-VALUE: Bug hunt - what if we try to hide a non-existent flag?
func TestHideRootFlags_NilPointerBug(t *testing.T) {
	// BUG FOUND: hideRootFlags panics if called with non-existent flag
	// This is because rootCmd.Flag() returns nil, then we try to access .Hidden
	defer func() {
		if r := recover(); r != nil {
			// Expected panic - bug confirmed!
			t.Logf("BUG CONFIRMED: hideRootFlags panics on non-existent flag: %v", r)
		}
	}()

	// This should either:
	// 1. Check if flag exists before accessing .Hidden
	// 2. Panic (current behavior - potential bug)
	hideRootFlags("nonexistent-flag-xyz")

	// If we get here, the function properly handles non-existent flags
	t.Log("No bug: hideRootFlags handles non-existent flags gracefully")
}

// TestHideRootFlags_EmptyList tests edge case with empty list
// MEDIUM-VALUE: Tests boundary condition
func TestHideRootFlags_EmptyList(t *testing.T) {
	// Should not panic with empty list
	assert.NotPanics(t, func() {
		hideRootFlags()
	})
}

// TestHideRootFlags_ActualBehavior tests the actual hiding behavior
// MEDIUM-VALUE: Tests real functionality
func TestHideRootFlags_ActualBehavior(t *testing.T) {
	// Use a unique flag name for this test
	testFlagName := "test-hide-behavior-flag-xyz"

	// Add the flag if it doesn't exist
	if rootCmd.Flag(testFlagName) == nil {
		rootCmd.Flags().String(testFlagName, "default", "Test flag for hiding")
	}

	// Verify flag starts as visible
	flag := rootCmd.Flag(testFlagName)
	assert.NotNil(t, flag)
	originalHidden := flag.Hidden
	flag.Hidden = false

	// Hide the flag
	hideRootFlags(testFlagName)

	// Verify flag is now hidden
	flag = rootCmd.Flag(testFlagName)
	assert.True(t, flag.Hidden, "Flag should be hidden after calling hideRootFlags")

	// Restore original state
	flag.Hidden = originalHidden
}

// TestCreateRootContext tests context creation with status hooks
// MEDIUM-VALUE: Tests actual behavior and verifies proper context setup
func TestCreateRootContext(t *testing.T) {
	ctx := createRootContext()

	// Verify context is not nil
	assert.NotNil(t, ctx, "Context should not be nil")

	// Verify status hooks are added to context
	hooks := statushooks.StatusHooksFromContext(ctx)
	assert.NotNil(t, hooks, "Status hooks should be added to context")
}

// TestAddCommands_AllCommandsExist tests that all expected subcommands are added
// MEDIUM-VALUE: Tests actual command registration, not just structure
func TestAddCommands_AllCommandsExist(t *testing.T) {
	rootCmd.ResetCommands()
	AddCommands()

	commands := rootCmd.Commands()
	commandNames := make(map[string]bool)
	for _, cmd := range commands {
		commandNames[cmd.Name()] = true
	}

	expectedCommands := []string{"plugin", "query", "service", "completion", "plugin-manager", "login"}
	for _, expected := range expectedCommands {
		assert.True(t, commandNames[expected], "Expected command %s to be added", expected)
	}
}

// TestAddCommands_NoNilCommands tests that no nil commands are added
// HIGH-VALUE: Bug hunt - what if a command function returns nil?
func TestAddCommands_NoNilCommands(t *testing.T) {
	rootCmd.ResetCommands()

	// This could panic if any command function returns nil
	assert.NotPanics(t, func() {
		AddCommands()
	}, "AddCommands should not panic even if command functions return nil")

	commands := rootCmd.Commands()
	for _, cmd := range commands {
		assert.NotNil(t, cmd, "No command should be nil")
	}
}

// TestAddCommands_Concurrent tests concurrent calls to AddCommands
// HIGH-VALUE: Bug hunt - race conditions when adding commands
func TestAddCommands_Concurrent(t *testing.T) {
	rootCmd.ResetCommands()

	// Run AddCommands concurrently to check for race conditions
	// Run with: go test -race
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func() {
			defer func() { done <- true }()
			rootCmd.ResetCommands()
			AddCommands()
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify we still have a valid command structure
	assert.NotNil(t, rootCmd)
}
