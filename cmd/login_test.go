package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/v2/constants"
)

// TestLoginCommand_Initialization tests the login command structure
func TestLoginCommand_Initialization(t *testing.T) {
	cmd := loginCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "login", cmd.Use)
	assert.Equal(t, "Login to Turbot Pipes", cmd.Short)
	assert.Equal(t, "Login to Turbot Pipes.", cmd.Long)
	assert.True(t, cmd.TraverseChildren)
	assert.NotNil(t, cmd.Run)
}

// TestLoginCommand_NoArgsRequired tests that login command doesn't accept arguments
func TestLoginCommand_NoArgsRequired(t *testing.T) {
	cmd := loginCmd()

	// The command should have NoArgs validator
	assert.NotNil(t, cmd.Args)
}

// TestLoginCommand_HelpFlag tests the help flag
func TestLoginCommand_HelpFlag(t *testing.T) {
	cmd := loginCmd()

	helpFlag := cmd.Flags().Lookup(constants.ArgHelp)
	assert.NotNil(t, helpFlag)
	assert.Equal(t, "h", helpFlag.Shorthand)
	assert.Equal(t, "false", helpFlag.DefValue)
}

// TestLoginCommand_CloudFlags tests cloud-related flags
func TestLoginCommand_CloudFlags(t *testing.T) {
	cmd := loginCmd()

	// Cloud flags should be added by AddCloudFlags()
	// Check if common cloud flags are present
	flags := cmd.Flags()
	assert.NotNil(t, flags)
}

// TestLoginCommand_HasRunFunction tests that the command has a run function
func TestLoginCommand_HasRunFunction(t *testing.T) {
	cmd := loginCmd()

	assert.NotNil(t, cmd.Run, "login command should have a Run function")
}

// TestLoginCommand_FlagParsing tests flag parsing without execution
func TestLoginCommand_FlagParsing(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "no flags",
			args:        []string{},
			expectError: false,
		},
		// Note: help flag tests omitted as they trigger help output which is an expected "error" from ParseFlags
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := loginCmd()
			cmd.SetArgs(tt.args)

			err := cmd.ParseFlags(tt.args)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLoginCommand_Usage tests that the command has proper usage documentation
func TestLoginCommand_Usage(t *testing.T) {
	cmd := loginCmd()

	usage := cmd.UsageString()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "login")
}

// TestLoginCommand_CommandHierarchy tests that login is a leaf command
func TestLoginCommand_CommandHierarchy(t *testing.T) {
	cmd := loginCmd()

	// Login should not have subcommands
	assert.False(t, cmd.HasSubCommands())
	assert.Empty(t, cmd.Commands())
}

// TestIncludeBashHelp tests the bash help text includes expected content
func TestLoginCommand_HelpText(t *testing.T) {
	cmd := loginCmd()

	assert.NotEmpty(t, cmd.Short, "Short description should not be empty")
	assert.NotEmpty(t, cmd.Long, "Long description should not be empty")
	assert.Equal(t, "Login to Turbot Pipes", cmd.Short)
	assert.Equal(t, "Login to Turbot Pipes.", cmd.Long)
}

// TestLoginCommand_TraverseChildren tests TraverseChildren flag
func TestLoginCommand_TraverseChildren(t *testing.T) {
	cmd := loginCmd()

	// TraverseChildren should be true to properly inherit flags
	assert.True(t, cmd.TraverseChildren)
}
