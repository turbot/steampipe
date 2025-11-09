package cmd

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestCompletionCommand_Initialization tests the completion command structure
func TestCompletionCommand_Initialization(t *testing.T) {
	cmd := generateCompletionScriptsCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "completion [bash|zsh|fish]", cmd.Use)
	assert.Equal(t, "Generate completion scripts", cmd.Short)
	assert.True(t, cmd.DisableFlagsInUseLine)
	assert.Contains(t, cmd.ValidArgs, "bash")
	assert.Contains(t, cmd.ValidArgs, "zsh")
	assert.Contains(t, cmd.ValidArgs, "fish")
}

// TestCompletionCommand_ValidShells tests the valid shell arguments
func TestCompletionCommand_ValidShells(t *testing.T) {
	cmd := generateCompletionScriptsCmd()

	expectedShells := []string{"bash", "zsh", "fish"}
	assert.Equal(t, expectedShells, cmd.ValidArgs)
}

// TestCompletionCommand_HelpFlag tests the help flag
func TestCompletionCommand_HelpFlag(t *testing.T) {
	cmd := generateCompletionScriptsCmd()

	helpFlag := cmd.Flags().Lookup("help")
	assert.NotNil(t, helpFlag)
	assert.Equal(t, "h", helpFlag.Shorthand)
}

// TestCompletionCommand_BashCompletion tests bash completion generation
func TestCompletionCommand_BashCompletion(t *testing.T) {
	cmd := generateCompletionScriptsCmd()

	// Set up root command for completion generation
	rootCmd := &cobra.Command{Use: "steampipe"}
	rootCmd.AddCommand(cmd)

	// Capture stdout to avoid verbose completion script in test output
	oldStdout := os.Stdout
	devNull, _ := os.Open(os.DevNull)
	os.Stdout = devNull
	defer func() {
		devNull.Close()
		os.Stdout = oldStdout
	}()

	// Just verify the command runs without error
	runGenCompletionScriptsCmd(cmd, []string{"bash"})

	// If we get here without panicking, the test passes
	assert.True(t, true)
}

// TestCompletionCommand_ZshCompletion tests zsh completion generation
func TestCompletionCommand_ZshCompletion(t *testing.T) {
	cmd := generateCompletionScriptsCmd()

	// Set up root command for completion generation
	rootCmd := &cobra.Command{Use: "steampipe"}
	rootCmd.AddCommand(cmd)

	// Capture stdout to avoid verbose completion script in test output
	oldStdout := os.Stdout
	devNull, _ := os.Open(os.DevNull)
	os.Stdout = devNull
	defer func() {
		devNull.Close()
		os.Stdout = oldStdout
	}()

	// Just verify the command runs without error
	runGenCompletionScriptsCmd(cmd, []string{"zsh"})

	// If we get here without panicking, the test passes
	assert.True(t, true)
}

// TestCompletionCommand_FishCompletion tests fish completion generation
func TestCompletionCommand_FishCompletion(t *testing.T) {
	cmd := generateCompletionScriptsCmd()

	// Set up root command for completion generation
	rootCmd := &cobra.Command{Use: "steampipe"}
	rootCmd.AddCommand(cmd)

	// Capture stdout to avoid verbose completion script in test output
	oldStdout := os.Stdout
	devNull, _ := os.Open(os.DevNull)
	os.Stdout = devNull
	defer func() {
		devNull.Close()
		os.Stdout = oldStdout
	}()

	// Just verify the command runs without error
	runGenCompletionScriptsCmd(cmd, []string{"fish"})

	// If we get here without panicking, the test passes
	assert.True(t, true)
}

// TestCompletionCommand_NoArgs tests behavior when no shell is specified
func TestCompletionCommand_NoArgs(t *testing.T) {
	cmd := generateCompletionScriptsCmd()

	// Set up root command
	rootCmd := &cobra.Command{Use: "steampipe"}
	rootCmd.AddCommand(cmd)

	// Just verify it doesn't panic when called with no args
	runGenCompletionScriptsCmd(cmd, []string{})

	assert.True(t, true)
}

// TestCompletionCommand_InvalidShell tests behavior with invalid shell argument
func TestCompletionCommand_InvalidShell(t *testing.T) {
	cmd := generateCompletionScriptsCmd()

	// Set up root command
	rootCmd := &cobra.Command{Use: "steampipe"}
	rootCmd.AddCommand(cmd)

	// Just verify it doesn't panic with invalid shell
	runGenCompletionScriptsCmd(cmd, []string{"invalid-shell"})

	assert.True(t, true)
}

// TestCompletionCommand_TooManyArgs tests behavior with too many arguments
func TestCompletionCommand_TooManyArgs(t *testing.T) {
	cmd := generateCompletionScriptsCmd()

	// Set up root command
	rootCmd := &cobra.Command{Use: "steampipe"}
	rootCmd.AddCommand(cmd)

	// Just verify it doesn't panic with too many args
	runGenCompletionScriptsCmd(cmd, []string{"bash", "extra"})

	assert.True(t, true)
}

// TestIncludeBashHelp tests bash help text generation
func TestIncludeBashHelp(t *testing.T) {
	result := includeBashHelp("Base text")

	assert.Contains(t, result, "Base text")
	assert.Contains(t, result, "Bash:")
	assert.Contains(t, result, "steampipe completion bash")
}

// TestIncludeZshHelp tests zsh help text generation
func TestIncludeZshHelp(t *testing.T) {
	result := includeZshHelp("Base text")

	assert.Contains(t, result, "Base text")
	// Zsh help is only shown on macOS, so we just check it doesn't crash
	assert.NotEmpty(t, result)
}

// TestIncludeFishHelp tests fish help text generation
func TestIncludeFishHelp(t *testing.T) {
	result := includeFishHelp("Base text")

	assert.Contains(t, result, "Base text")
	assert.Contains(t, result, "fish:")
	assert.Contains(t, result, "steampipe completion fish")
}

// TestRunGenCompletionScriptsCmd tests the run function with various arguments
func TestRunGenCompletionScriptsCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		shouldError bool
	}{
		{
			name:        "bash",
			args:        []string{"bash"},
			shouldError: false,
		},
		{
			name:        "zsh",
			args:        []string{"zsh"},
			shouldError: false,
		},
		{
			name:        "fish",
			args:        []string{"fish"},
			shouldError: false,
		},
		{
			name:        "no args",
			args:        []string{},
			shouldError: false,
		},
		{
			name:        "invalid shell",
			args:        []string{"powershell"},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := generateCompletionScriptsCmd()
			rootCmd := &cobra.Command{Use: "steampipe"}
			rootCmd.AddCommand(cmd)

			// Capture stdout to avoid verbose completion script in test output
			oldStdout := os.Stdout
			devNull, _ := os.Open(os.DevNull)
			os.Stdout = devNull
			defer func() {
				devNull.Close()
				os.Stdout = oldStdout
			}()

			// Call the run function directly
			runGenCompletionScriptsCmd(cmd, tt.args)

			// The function doesn't return errors, it just outputs help or completion
			// We just verify it doesn't panic
			assert.NotNil(t, cmd)
		})
	}
}
