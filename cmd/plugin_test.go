package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/v2/constants"
)

// TestPluginCommand_Initialization tests the plugin command structure
func TestPluginCommand_Initialization(t *testing.T) {
	cmd := pluginCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "plugin [command]", cmd.Use)
	assert.Equal(t, "Steampipe plugin management", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.Contains(t, cmd.Long, "Plugins extend Steampipe")
	assert.NotNil(t, cmd.PersistentPostRun)
}

// TestPluginCommand_Subcommands tests that all expected subcommands are present
func TestPluginCommand_Subcommands(t *testing.T) {
	cmd := pluginCmd()

	expectedSubcommands := []string{"install", "list", "uninstall", "update"}

	assert.True(t, cmd.HasSubCommands())
	subcommands := cmd.Commands()

	for _, expected := range expectedSubcommands {
		found := false
		for _, subCmd := range subcommands {
			if subCmd.Name() == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected subcommand %s not found", expected)
	}
}

// TestPluginCommand_FindSubcommands tests finding specific subcommands
func TestPluginCommand_FindSubcommands(t *testing.T) {
	tests := []struct {
		subcommand string
		shouldFind bool
	}{
		{"install", true},
		{"update", true},
		{"list", true},
		{"uninstall", true},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.subcommand, func(t *testing.T) {
			cmd := pluginCmd()
			subCmd, _, err := cmd.Find([]string{tt.subcommand})

			if tt.shouldFind {
				assert.NoError(t, err)
				assert.NotNil(t, subCmd)
				assert.Equal(t, tt.subcommand, subCmd.Name())
			}
		})
	}
}

// TestPluginCommand_HelpFlag tests the help flag
func TestPluginCommand_HelpFlag(t *testing.T) {
	cmd := pluginCmd()

	helpFlag := cmd.Flags().Lookup(constants.ArgHelp)
	assert.NotNil(t, helpFlag)
	assert.Equal(t, "h", helpFlag.Shorthand)
}

// TestPluginInstallCmd_Initialization tests the install subcommand structure
func TestPluginInstallCmd_Initialization(t *testing.T) {
	cmd := pluginInstallCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "install [flags] [registry/org/]name[@version]", cmd.Use)
	assert.Equal(t, "Install one or more plugins", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.Run)
}

// TestPluginInstallCmd_Flags tests install command flags
func TestPluginInstallCmd_Flags(t *testing.T) {
	cmd := pluginInstallCmd()

	tests := []struct {
		flagName     string
		shouldExist  bool
		defaultValue string
	}{
		{constants.ArgProgress, true, "true"},
		{constants.ArgSkipConfig, true, "false"},
		{constants.ArgHelp, true, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)

			if tt.shouldExist {
				assert.NotNil(t, flag, "Flag %s should exist", tt.flagName)
				if tt.defaultValue != "" {
					assert.Equal(t, tt.defaultValue, flag.DefValue)
				}
			} else {
				assert.Nil(t, flag, "Flag %s should not exist", tt.flagName)
			}
		})
	}
}

// TestPluginInstallCmd_FlagParsing tests flag parsing for install command
func TestPluginInstallCmd_FlagParsing(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedProgress bool
		expectSkipConfig bool
		expectError      bool
	}{
		{
			name:             "default flags",
			args:             []string{},
			expectedProgress: true,
			expectSkipConfig: false,
			expectError:      false,
		},
		{
			name:             "disable progress",
			args:             []string{"--progress=false"},
			expectedProgress: false,
			expectSkipConfig: false,
			expectError:      false,
		},
		{
			name:             "skip config",
			args:             []string{"--skip-config"},
			expectedProgress: true,
			expectSkipConfig: true,
			expectError:      false,
		},
		{
			name:             "with plugin name",
			args:             []string{"aws"},
			expectedProgress: true,
			expectSkipConfig: false,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := pluginInstallCmd()
			cmd.SetArgs(tt.args)

			err := cmd.ParseFlags(tt.args)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				progress, _ := cmd.Flags().GetBool(constants.ArgProgress)
				assert.Equal(t, tt.expectedProgress, progress)

				skipConfig, _ := cmd.Flags().GetBool(constants.ArgSkipConfig)
				assert.Equal(t, tt.expectSkipConfig, skipConfig)
			}
		})
	}
}

// TestPluginUpdateCmd_Initialization tests the update subcommand structure
func TestPluginUpdateCmd_Initialization(t *testing.T) {
	cmd := pluginUpdateCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "update [flags] [registry/org/]name[@version]", cmd.Use)
	assert.Equal(t, "Update one or more plugins", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.Run)
}

// TestPluginUpdateCmd_Flags tests update command flags
func TestPluginUpdateCmd_Flags(t *testing.T) {
	cmd := pluginUpdateCmd()

	tests := []struct {
		flagName     string
		shouldExist  bool
		defaultValue string
	}{
		{constants.ArgAll, true, "false"},
		{constants.ArgProgress, true, "true"},
		{constants.ArgHelp, true, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)

			if tt.shouldExist {
				assert.NotNil(t, flag, "Flag %s should exist", tt.flagName)
				if tt.defaultValue != "" {
					assert.Equal(t, tt.defaultValue, flag.DefValue)
				}
			}
		})
	}
}

// TestPluginUpdateCmd_FlagParsing tests flag parsing for update command
func TestPluginUpdateCmd_FlagParsing(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedAll      bool
		expectedProgress bool
		expectError      bool
	}{
		{
			name:             "default flags",
			args:             []string{},
			expectedAll:      false,
			expectedProgress: true,
			expectError:      false,
		},
		{
			name:             "update all",
			args:             []string{"--all"},
			expectedAll:      true,
			expectedProgress: true,
			expectError:      false,
		},
		{
			name:             "disable progress",
			args:             []string{"--progress=false"},
			expectedAll:      false,
			expectedProgress: false,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := pluginUpdateCmd()
			cmd.SetArgs(tt.args)

			err := cmd.ParseFlags(tt.args)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				all, _ := cmd.Flags().GetBool(constants.ArgAll)
				assert.Equal(t, tt.expectedAll, all)

				progress, _ := cmd.Flags().GetBool(constants.ArgProgress)
				assert.Equal(t, tt.expectedProgress, progress)
			}
		})
	}
}

// TestPluginListCmd_Initialization tests the list subcommand structure
func TestPluginListCmd_Initialization(t *testing.T) {
	cmd := pluginListCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List currently installed plugins", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.Run)
}

// TestPluginListCmd_Flags tests list command flags
func TestPluginListCmd_Flags(t *testing.T) {
	cmd := pluginListCmd()

	tests := []struct {
		flagName     string
		shouldExist  bool
		defaultValue string
	}{
		{"outdated", true, "false"},
		{constants.ArgOutput, true, "table"},
		{constants.ArgHelp, true, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)

			if tt.shouldExist {
				assert.NotNil(t, flag, "Flag %s should exist", tt.flagName)
				if tt.defaultValue != "" {
					assert.Equal(t, tt.defaultValue, flag.DefValue)
				}
			}
		})
	}
}

// TestPluginListCmd_FlagParsing tests flag parsing for list command
func TestPluginListCmd_FlagParsing(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedOutdated bool
		expectedOutput   string
		expectError      bool
	}{
		{
			name:             "default flags",
			args:             []string{},
			expectedOutdated: false,
			expectedOutput:   "table",
			expectError:      false,
		},
		{
			name:             "outdated flag",
			args:             []string{"--outdated"},
			expectedOutdated: true,
			expectedOutput:   "table",
			expectError:      false,
		},
		{
			name:             "json output",
			args:             []string{"--output", "json"},
			expectedOutdated: false,
			expectedOutput:   "json",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := pluginListCmd()
			cmd.SetArgs(tt.args)

			err := cmd.ParseFlags(tt.args)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				outdated, _ := cmd.Flags().GetBool("outdated")
				assert.Equal(t, tt.expectedOutdated, outdated)

				output, _ := cmd.Flags().GetString(constants.ArgOutput)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

// TestPluginUninstallCmd_Initialization tests the uninstall subcommand structure
func TestPluginUninstallCmd_Initialization(t *testing.T) {
	cmd := pluginUninstallCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "uninstall [flags] [registry/org/]name", cmd.Use)
	assert.Equal(t, "Uninstall a plugin", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.Run)
}

// TestPluginUninstallCmd_Flags tests uninstall command flags
func TestPluginUninstallCmd_Flags(t *testing.T) {
	cmd := pluginUninstallCmd()

	helpFlag := cmd.Flags().Lookup(constants.ArgHelp)
	assert.NotNil(t, helpFlag)
	assert.Equal(t, "h", helpFlag.Shorthand)
	assert.Equal(t, "false", helpFlag.DefValue)
}

// TestPluginUninstallCmd_FlagParsing tests flag parsing for uninstall command
func TestPluginUninstallCmd_FlagParsing(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "no args",
			args:        []string{},
			expectError: false,
		},
		{
			name:        "with plugin name",
			args:        []string{"aws"},
			expectError: false,
		},
		// Note: help flag test omitted as it triggers help output which is an error in ParseFlags
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := pluginUninstallCmd()
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

// TestPluginCommand_LongHelp tests that all commands have proper long help text
func TestPluginCommand_LongHelp(t *testing.T) {
	commands := []struct {
		name    string
		cmdFunc func() *cobra.Command
	}{
		{"plugin", pluginCmd},
		{"install", pluginInstallCmd},
		{"update", pluginUpdateCmd},
		{"list", pluginListCmd},
		{"uninstall", pluginUninstallCmd},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			cmd := tc.cmdFunc()

			assert.NotEmpty(t, cmd.Short, "%s should have short description", tc.name)
			assert.NotEmpty(t, cmd.Long, "%s should have long description", tc.name)
			assert.NotEmpty(t, cmd.UsageString(), "%s should have usage string", tc.name)
		})
	}
}

// TestPluginInstallSteps tests that install steps are defined
func TestPluginInstallSteps(t *testing.T) {
	assert.NotEmpty(t, pluginInstallSteps)
	assert.Contains(t, pluginInstallSteps, "Downloading")
	assert.Contains(t, pluginInstallSteps, "Installing Plugin")
	assert.Contains(t, pluginInstallSteps, "Done")
}
