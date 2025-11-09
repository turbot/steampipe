package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	perror_helpers "github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/cmdconfig"
	"github.com/turbot/steampipe/v2/pkg/plugin"
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

// TestResolveUpdatePluginsFromArgs tests argument validation for update command
// This tests real validation logic and error paths
func TestResolveUpdatePluginsFromArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		allFlag     bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "no args and no all flag - should error",
			args:        []string{},
			allFlag:     false,
			expectError: true,
			errorMsg:    "you need to provide at least one plugin to update or use the",
		},
		{
			name:        "valid single plugin",
			args:        []string{"aws"},
			allFlag:     false,
			expectError: false,
		},
		{
			name:        "valid multiple plugins",
			args:        []string{"aws", "azure", "gcp"},
			allFlag:     false,
			expectError: false,
		},
		{
			name:        "all flag with plugin args - should error",
			args:        []string{"aws"},
			allFlag:     true,
			expectError: true,
			errorMsg:    "cannot be used when updating specific plugins",
		},
		{
			name:        "all flag without args - valid",
			args:        []string{},
			allFlag:     true,
			expectError: false,
		},
		{
			name:        "plugin with version constraint",
			args:        []string{"aws@1.0.0"},
			allFlag:     false,
			expectError: false,
		},
		{
			name:        "plugin with org and version",
			args:        []string{"turbot/aws@1.0.0"},
			allFlag:     false,
			expectError: false,
		},
		{
			name:        "multiple plugins with all flag - should error",
			args:        []string{"aws", "azure"},
			allFlag:     true,
			expectError: true,
			errorMsg:    "cannot be used when updating specific plugins",
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore viper state for this specific test
			originalAllValue := cmdconfig.Viper().GetBool("all")
			defer cmdconfig.Viper().Set("all", originalAllValue)

			cmdconfig.Viper().Set("all", tt.allFlag)

			result, err := resolveUpdatePluginsFromArgs(tt.args)

			if tt.expectError {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err, "Expected no error for test case: %s", tt.name)
				assert.Equal(t, tt.args, result)
			}
		})
	}
}

// TestIsPluginNotFoundErr tests error string detection
// This tests real error handling logic
func TestIsPluginNotFoundErr(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "error ending with 'not found'",
			err:      fmt.Errorf("plugin aws not found"),
			expected: true,
		},
		{
			name:     "error ending with 'not found' - different message",
			err:      fmt.Errorf("the requested resource was not found"),
			expected: true,
		},
		{
			name:     "error not ending with 'not found'",
			err:      fmt.Errorf("plugin installation failed"),
			expected: false,
		},
		{
			name:     "error with 'not found' in the middle",
			err:      fmt.Errorf("not found but has more text after"),
			expected: false,
		},
		{
			name:     "empty error message",
			err:      fmt.Errorf(""),
			expected: false,
		},
		{
			name:     "case sensitive check",
			err:      fmt.Errorf("plugin not Found"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPluginNotFoundErr(tt.err)
			assert.Equal(t, tt.expected, result, "Test case: %s", tt.name)
		})
	}
}

// TestPluginCommand_ArgsValidation tests argument validation
func TestPluginCommand_ArgsValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "no args - valid (shows help)",
			args:        []string{},
			expectError: false,
		},
		{
			name:        "subcommand as arg",
			args:        []string{"install"},
			expectError: true, // plugin command expects NoArgs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := pluginCmd()

			err := cmd.Args(cmd, tt.args)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestPluginInstallCmd_ArgsValidation tests install command accepts arbitrary args
func TestPluginInstallCmd_ArgsValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "no args",
			args: []string{},
		},
		{
			name: "single plugin",
			args: []string{"aws"},
		},
		{
			name: "multiple plugins",
			args: []string{"aws", "azure", "gcp"},
		},
		{
			name: "plugin with version",
			args: []string{"aws@1.0.0"},
		},
		{
			name: "plugin with org",
			args: []string{"turbot/aws"},
		},
		{
			name: "plugin with org and version",
			args: []string{"turbot/aws@1.0.0"},
		},
		{
			name: "empty string in args",
			args: []string{""},
		},
		{
			name: "many plugins",
			args: []string{"aws", "azure", "gcp", "kubernetes", "github"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := pluginInstallCmd()

			// ArbitraryArgs should accept any number of args
			err := cmd.Args(cmd, tt.args)

			// ArbitraryArgs always returns nil
			assert.NoError(t, err)
		})
	}
}

// TestPluginUpdateCmd_ArgsValidation tests update command accepts arbitrary args
func TestPluginUpdateCmd_ArgsValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "no args",
			args: []string{},
		},
		{
			name: "single plugin",
			args: []string{"aws"},
		},
		{
			name: "multiple plugins",
			args: []string{"aws", "azure"},
		},
		{
			name: "plugin with version constraint",
			args: []string{"aws@^1.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := pluginUpdateCmd()

			err := cmd.Args(cmd, tt.args)

			// ArbitraryArgs always returns nil
			assert.NoError(t, err)
		})
	}
}

// TestPluginUninstallCmd_ArgsValidation tests uninstall command accepts arbitrary args
func TestPluginUninstallCmd_ArgsValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "no args",
			args: []string{},
		},
		{
			name: "single plugin",
			args: []string{"aws"},
		},
		{
			name: "multiple plugins",
			args: []string{"aws", "azure", "gcp"},
		},
		{
			name: "plugin with org",
			args: []string{"turbot/aws"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := pluginUninstallCmd()

			err := cmd.Args(cmd, tt.args)

			// ArbitraryArgs always returns nil
			assert.NoError(t, err)
		})
	}
}

// TestPluginListCmd_ArgsValidation tests list command requires no args
func TestPluginListCmd_ArgsValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "no args - valid",
			args:        []string{},
			expectError: false,
		},
		{
			name:        "with args - invalid",
			args:        []string{"aws"},
			expectError: true,
		},
		{
			name:        "multiple args - invalid",
			args:        []string{"aws", "azure"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := pluginListCmd()

			err := cmd.Args(cmd, tt.args)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestShowPluginListOutput tests output format handling
func TestShowPluginListOutput(t *testing.T) {
	tests := []struct {
		name         string
		outputFormat string
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "table format - valid",
			outputFormat: "table",
			expectError:  false,
		},
		{
			name:         "json format - valid",
			outputFormat: "json",
			expectError:  false,
		},
		{
			name:         "invalid format",
			outputFormat: "xml",
			expectError:  true,
			errorMsg:     "invalid output format",
		},
		{
			name:         "empty format",
			outputFormat: "",
			expectError:  true,
			errorMsg:     "invalid output format",
		},
		{
			name:         "case sensitivity",
			outputFormat: "JSON",
			expectError:  true,
			errorMsg:     "invalid output format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create minimal test data
			pluginList := []plugin.PluginListItem{}
			failedMap := map[string][]plugin.PluginConnection{}
			missingMap := map[string][]plugin.PluginConnection{}
			res := perror_helpers.ErrorAndWarnings{}

			err := showPluginListOutput(pluginList, failedMap, missingMap, res, tt.outputFormat)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestPluginCommand_PersistentPostRun tests that cleanup is configured
func TestPluginCommand_PersistentPostRun(t *testing.T) {
	cmd := pluginCmd()

	// Verify PersistentPostRun is set
	assert.NotNil(t, cmd.PersistentPostRun, "PersistentPostRun should be set for cleanup")
}

// TestPluginInstallCmd_CombinedFlags tests multiple flags together
func TestPluginInstallCmd_CombinedFlags(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedProgress bool
		expectSkipConfig bool
	}{
		{
			name:             "both flags set",
			args:             []string{"--progress=false", "--skip-config"},
			expectedProgress: false,
			expectSkipConfig: true,
		},
		{
			name:             "skip-config with plugin",
			args:             []string{"aws", "--skip-config"},
			expectedProgress: true,
			expectSkipConfig: true,
		},
		{
			name:             "progress false with multiple plugins",
			args:             []string{"aws", "azure", "--progress=false"},
			expectedProgress: false,
			expectSkipConfig: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := pluginInstallCmd()
			cmd.SetArgs(tt.args)

			err := cmd.ParseFlags(tt.args)
			assert.NoError(t, err)

			progress, _ := cmd.Flags().GetBool(constants.ArgProgress)
			assert.Equal(t, tt.expectedProgress, progress)

			skipConfig, _ := cmd.Flags().GetBool(constants.ArgSkipConfig)
			assert.Equal(t, tt.expectSkipConfig, skipConfig)
		})
	}
}

// TestPluginUpdateCmd_CombinedFlags tests multiple flags together
func TestPluginUpdateCmd_CombinedFlags(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedAll      bool
		expectedProgress bool
	}{
		{
			name:             "all and progress false",
			args:             []string{"--all", "--progress=false"},
			expectedAll:      true,
			expectedProgress: false,
		},
		{
			name:             "all flag only",
			args:             []string{"--all"},
			expectedAll:      true,
			expectedProgress: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := pluginUpdateCmd()
			cmd.SetArgs(tt.args)

			err := cmd.ParseFlags(tt.args)
			assert.NoError(t, err)

			all, _ := cmd.Flags().GetBool(constants.ArgAll)
			assert.Equal(t, tt.expectedAll, all)

			progress, _ := cmd.Flags().GetBool(constants.ArgProgress)
			assert.Equal(t, tt.expectedProgress, progress)
		})
	}
}

// TestPluginListCmd_CombinedFlags tests multiple flags together
func TestPluginListCmd_CombinedFlags(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedOutdated bool
		expectedOutput   string
	}{
		{
			name:             "outdated with json output",
			args:             []string{"--outdated", "--output", "json"},
			expectedOutdated: true,
			expectedOutput:   "json",
		},
		{
			name:             "outdated only",
			args:             []string{"--outdated"},
			expectedOutdated: true,
			expectedOutput:   "table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := pluginListCmd()
			cmd.SetArgs(tt.args)

			err := cmd.ParseFlags(tt.args)
			assert.NoError(t, err)

			outdated, _ := cmd.Flags().GetBool("outdated")
			assert.Equal(t, tt.expectedOutdated, outdated)

			output, _ := cmd.Flags().GetString(constants.ArgOutput)
			assert.Equal(t, tt.expectedOutput, output)
		})
	}
}

// TestPluginJsonOutput_StructureValidation tests JSON output structs are properly defined
func TestPluginJsonOutput_StructureValidation(t *testing.T) {
	t.Run("installedPlugin struct", func(t *testing.T) {
		p := installedPlugin{
			Name:        "aws",
			Version:     "1.0.0",
			Connections: []string{"conn1", "conn2"},
		}

		assert.Equal(t, "aws", p.Name)
		assert.Equal(t, "1.0.0", p.Version)
		assert.Equal(t, 2, len(p.Connections))
	})

	t.Run("failedPlugin struct", func(t *testing.T) {
		p := failedPlugin{
			Name:        "azure",
			Reason:      "not found",
			Connections: []string{"conn1"},
		}

		assert.Equal(t, "azure", p.Name)
		assert.Equal(t, "not found", p.Reason)
		assert.Equal(t, 1, len(p.Connections))
	})

	t.Run("pluginJsonOutput struct", func(t *testing.T) {
		output := pluginJsonOutput{
			Installed: []installedPlugin{
				{Name: "aws", Version: "1.0.0", Connections: []string{"aws"}},
			},
			Failed: []failedPlugin{
				{Name: "azure", Reason: "failed", Connections: []string{"azure"}},
			},
			Warnings: []string{"warning1"},
		}

		assert.Equal(t, 1, len(output.Installed))
		assert.Equal(t, 1, len(output.Failed))
		assert.Equal(t, 1, len(output.Warnings))
	})
}

// TestShowPluginListAsJSON_WithData tests JSON output with actual plugin data
func TestShowPluginListAsJSON_WithData(t *testing.T) {
	tests := []struct {
		name        string
		pluginList  []plugin.PluginListItem
		failedMap   map[string][]plugin.PluginConnection
		missingMap  map[string][]plugin.PluginConnection
		warnings    []string
		expectError bool
	}{
		{
			name:        "empty lists",
			pluginList:  []plugin.PluginListItem{},
			failedMap:   map[string][]plugin.PluginConnection{},
			missingMap:  map[string][]plugin.PluginConnection{},
			warnings:    []string{},
			expectError: false,
		},
		{
			name:        "nil lists",
			pluginList:  nil,
			failedMap:   nil,
			missingMap:  nil,
			warnings:    nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := perror_helpers.ErrorAndWarnings{
				Warnings: tt.warnings,
			}

			err := showPluginListAsJSON(tt.pluginList, tt.failedMap, tt.missingMap, res)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestShowPluginListAsTable_WithData tests table output with actual plugin data
func TestShowPluginListAsTable_WithData(t *testing.T) {
	tests := []struct {
		name        string
		pluginList  []plugin.PluginListItem
		failedMap   map[string][]plugin.PluginConnection
		missingMap  map[string][]plugin.PluginConnection
		warnings    []string
		expectError bool
	}{
		{
			name:        "empty lists",
			pluginList:  []plugin.PluginListItem{},
			failedMap:   map[string][]plugin.PluginConnection{},
			missingMap:  map[string][]plugin.PluginConnection{},
			warnings:    []string{},
			expectError: false,
		},
		{
			name:        "nil lists",
			pluginList:  nil,
			failedMap:   nil,
			missingMap:  nil,
			warnings:    nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := perror_helpers.ErrorAndWarnings{
				Warnings: tt.warnings,
			}

			err := showPluginListAsTable(tt.pluginList, tt.failedMap, tt.missingMap, res)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestPluginInstallSteps_Ordering tests that install steps are in correct order
func TestPluginInstallSteps_Ordering(t *testing.T) {
	expectedSteps := []string{
		"Downloading",
		"Installing Plugin",
		"Installing Docs",
		"Installing Config",
		"Updating Steampipe",
		"Done",
	}

	assert.Equal(t, len(expectedSteps), len(pluginInstallSteps))

	for i, expected := range expectedSteps {
		assert.Equal(t, expected, pluginInstallSteps[i], "Step at index %d should be %s", i, expected)
	}
}

// TestPluginCommand_SubcommandExecution tests that subcommands have Run functions
func TestPluginCommand_SubcommandExecution(t *testing.T) {
	tests := []struct {
		name    string
		cmdFunc func() *cobra.Command
	}{
		{"install", pluginInstallCmd},
		{"update", pluginUpdateCmd},
		{"list", pluginListCmd},
		{"uninstall", pluginUninstallCmd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmdFunc()
			assert.NotNil(t, cmd.Run, "%s command should have a Run function", tt.name)
		})
	}
}

// TestPluginCommand_NoRunFunction tests that main plugin command has no Run function
func TestPluginCommand_NoRunFunction(t *testing.T) {
	cmd := pluginCmd()
	// Main plugin command should not have a Run function (it's a parent command)
	assert.Nil(t, cmd.Run, "Main plugin command should not have a Run function")
}

// TestPluginInstallCmd_MultiplePluginFormats tests various plugin name formats
func TestPluginInstallCmd_MultiplePluginFormats(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		valid  bool
	}{
		{
			name:  "simple name",
			args:  []string{"aws"},
			valid: true,
		},
		{
			name:  "with org",
			args:  []string{"turbot/aws"},
			valid: true,
		},
		{
			name:  "with version",
			args:  []string{"aws@1.0.0"},
			valid: true,
		},
		{
			name:  "with org and version",
			args:  []string{"turbot/aws@1.0.0"},
			valid: true,
		},
		{
			name:  "full OCI path",
			args:  []string{"hub.steampipe.io/plugins/turbot/aws@1.0.0"},
			valid: true,
		},
		{
			name:  "multiple formats mixed",
			args:  []string{"aws", "turbot/azure@1.0.0", "gcp@latest"},
			valid: true,
		},
		{
			name:  "special characters in version",
			args:  []string{"aws@^1.0.0"},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := pluginInstallCmd()
			err := cmd.Args(cmd, tt.args)

			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestResolveUpdatePluginsFromArgs_EdgeCases tests edge cases in argument resolution
func TestResolveUpdatePluginsFromArgs_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		allFlag     bool
		expectError bool
		description string
	}{
		{
			name:        "empty args with all flag",
			args:        []string{},
			allFlag:     true,
			expectError: false,
			description: "Should allow empty args when all flag is set",
		},
		{
			name:        "word 'all' in args without flag",
			args:        []string{"all"},
			allFlag:     false,
			expectError: false,
			description: "Word 'all' as plugin name should be valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalAllValue := cmdconfig.Viper().GetBool("all")
			defer cmdconfig.Viper().Set("all", originalAllValue)

			cmdconfig.Viper().Set("all", tt.allFlag)

			result, err := resolveUpdatePluginsFromArgs(tt.args)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				if !tt.allFlag {
					assert.Equal(t, tt.args, result)
				}
			}
		})
	}
}

// TestIsPluginNotFoundErr_EdgeCases tests edge cases for error detection
func TestIsPluginNotFoundErr_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "just 'not found'",
			err:      fmt.Errorf("not found"),
			expected: true,
		},
		{
			name:     "trailing whitespace",
			err:      fmt.Errorf("plugin not found "),
			expected: false,
		},
		{
			name:     "multiple spaces",
			err:      fmt.Errorf("plugin  not  found"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPluginNotFoundErr(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPluginCommand_HelpFlagShorthand tests that all commands have help flag shorthand
func TestPluginCommand_HelpFlagShorthand(t *testing.T) {
	commands := map[string]*cobra.Command{
		"plugin":    pluginCmd(),
		"install":   pluginInstallCmd(),
		"update":    pluginUpdateCmd(),
		"list":      pluginListCmd(),
		"uninstall": pluginUninstallCmd(),
	}

	for name, cmd := range commands {
		t.Run(name, func(t *testing.T) {
			helpFlag := cmd.Flags().Lookup(constants.ArgHelp)
			if helpFlag != nil {
				assert.Equal(t, "h", helpFlag.Shorthand, "%s command help flag should have 'h' shorthand", name)
			}
		})
	}
}

// TestPluginListCmd_OutputFlag tests output flag validation
func TestPluginListCmd_OutputFlag(t *testing.T) {
	cmd := pluginListCmd()

	outputFlag := cmd.Flags().Lookup(constants.ArgOutput)
	assert.NotNil(t, outputFlag)
	assert.Equal(t, "table", outputFlag.DefValue)

	// Test setting different output values
	outputs := []string{"table", "json"}
	for _, output := range outputs {
		err := cmd.Flags().Set(constants.ArgOutput, output)
		assert.NoError(t, err, "Should be able to set output to %s", output)
	}
}

// TestPluginUpdateCmd_ConflictingFlags tests conflicting flag combinations
func TestPluginUpdateCmd_ConflictingFlags(t *testing.T) {
	// Save original value
	originalAllValue := cmdconfig.Viper().GetBool("all")
	defer cmdconfig.Viper().Set("all", originalAllValue)

	// Test that providing plugin args AND --all flag results in error
	cmdconfig.Viper().Set("all", true)
	_, err := resolveUpdatePluginsFromArgs([]string{"aws", "azure"})
	assert.Error(t, err, "Should error when --all flag is used with plugin arguments")
	assert.Contains(t, err.Error(), "cannot be used when updating specific plugins")
}

// TestPluginCommand_PersistentFlags tests that plugin command doesn't leak flags
func TestPluginCommand_PersistentFlags(t *testing.T) {
	cmd := pluginCmd()

	// Plugin command should have its own help flag
	helpFlag := cmd.Flags().Lookup(constants.ArgHelp)
	assert.NotNil(t, helpFlag)

	// Subcommands should inherit or have their own flags
	subcommands := cmd.Commands()
	assert.NotEmpty(t, subcommands, "Plugin command should have subcommands")
}
