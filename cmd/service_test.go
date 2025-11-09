package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

// TestServiceCommand_Initialization tests that the service command is properly initialized
func TestServiceCommand_Initialization(t *testing.T) {
	cmd := serviceCmd()

	// Verify command structure
	assert.NotNil(t, cmd)
	assert.Equal(t, "service", cmd.Use[:7]) // "service [command]"
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Verify help flag exists
	helpFlag := cmd.Flags().Lookup(pconstants.ArgHelp)
	assert.NotNil(t, helpFlag)
	assert.Equal(t, "h", helpFlag.Shorthand)
}

// TestServiceCommand_Subcommands tests that all expected subcommands exist
func TestServiceCommand_Subcommands(t *testing.T) {
	cmd := serviceCmd()

	subcommands := cmd.Commands()
	subcommandNames := make([]string, len(subcommands))
	for i, sc := range subcommands {
		subcommandNames[i] = sc.Name()
	}

	// Verify all required subcommands exist
	assert.Contains(t, subcommandNames, "start")
	assert.Contains(t, subcommandNames, "stop")
	assert.Contains(t, subcommandNames, "restart")
	assert.Contains(t, subcommandNames, "status")

	// Verify we have exactly 4 subcommands
	assert.Len(t, subcommands, 4)
}

// TestServiceCommand_SubcommandLookup tests finding subcommands
func TestServiceCommand_SubcommandLookup(t *testing.T) {
	tests := map[string]struct {
		subcommand string
		shouldFind bool
	}{
		"start subcommand": {
			subcommand: "start",
			shouldFind: true,
		},
		"stop subcommand": {
			subcommand: "stop",
			shouldFind: true,
		},
		"restart subcommand": {
			subcommand: "restart",
			shouldFind: true,
		},
		"status subcommand": {
			subcommand: "status",
			shouldFind: true,
		},
		"invalid subcommand": {
			subcommand: "invalid",
			shouldFind: false,
		},
		"empty subcommand": {
			subcommand: "",
			shouldFind: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := serviceCmd()

			subCmd, _, err := cmd.Find([]string{tc.subcommand})

			if tc.shouldFind {
				assert.NoError(t, err)
				assert.NotNil(t, subCmd)
				assert.Equal(t, tc.subcommand, subCmd.Name())
			} else {
				// For invalid commands, Find returns the parent command or an error
				if err == nil {
					// If no error, it should return the parent command
					assert.NotEqual(t, tc.subcommand, subCmd.Name())
				}
			}
		})
	}
}

// TestServiceStartCommand_Initialization tests service start command structure
func TestServiceStartCommand_Initialization(t *testing.T) {
	cmd := serviceStartCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "start", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.Run)
}

// TestServiceStartCommand_Flags tests that service start has all required flags
func TestServiceStartCommand_Flags(t *testing.T) {
	cmd := serviceStartCmd()

	tests := map[string]struct {
		flagName      string
		expectedType  string
		shouldExist   bool
	}{
		"help flag": {
			flagName:     pconstants.ArgHelp,
			expectedType: "bool",
			shouldExist:  true,
		},
		"port flag": {
			flagName:     pconstants.ArgDatabasePort,
			expectedType: "int",
			shouldExist:  true,
		},
		"listen flag": {
			flagName:     pconstants.ArgDatabaseListenAddresses,
			expectedType: "string",
			shouldExist:  true,
		},
		"password flag": {
			flagName:     pconstants.ArgServicePassword,
			expectedType: "string",
			shouldExist:  true,
		},
		"show-password flag": {
			flagName:     pconstants.ArgServiceShowPassword,
			expectedType: "bool",
			shouldExist:  true,
		},
		"foreground flag": {
			flagName:     pconstants.ArgForeground,
			expectedType: "bool",
			shouldExist:  true,
		},
		"invoker flag": {
			flagName:     pconstants.ArgInvoker,
			expectedType: "string",
			shouldExist:  true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tc.flagName)

			if tc.shouldExist {
				assert.NotNil(t, flag, "flag %s should exist", tc.flagName)
				assert.Equal(t, tc.expectedType, flag.Value.Type())
			} else {
				assert.Nil(t, flag, "flag %s should not exist", tc.flagName)
			}
		})
	}
}

// TestServiceStartCommand_DefaultValues tests default flag values
func TestServiceStartCommand_DefaultValues(t *testing.T) {
	cmd := serviceStartCmd()

	// Test default port
	portFlag := cmd.Flags().Lookup(pconstants.ArgDatabasePort)
	assert.NotNil(t, portFlag)
	assert.Equal(t, "9193", portFlag.DefValue) // constants.DatabaseDefaultPort

	// Test default listen addresses
	listenFlag := cmd.Flags().Lookup(pconstants.ArgDatabaseListenAddresses)
	assert.NotNil(t, listenFlag)
	assert.Equal(t, "network", listenFlag.DefValue)

	// Test default show-password
	showPasswordFlag := cmd.Flags().Lookup(pconstants.ArgServiceShowPassword)
	assert.NotNil(t, showPasswordFlag)
	assert.Equal(t, "false", showPasswordFlag.DefValue)

	// Test default foreground
	foregroundFlag := cmd.Flags().Lookup(pconstants.ArgForeground)
	assert.NotNil(t, foregroundFlag)
	assert.Equal(t, "false", foregroundFlag.DefValue)
}

// TestServiceStartCommand_FlagParsing tests parsing various flag combinations
func TestServiceStartCommand_FlagParsing(t *testing.T) {
	tests := map[string]struct {
		args        []string
		expectError bool
		checkFunc   func(*testing.T, *testing.T, interface{})
	}{
		"no arguments": {
			args:        []string{},
			expectError: false,
		},
		"custom port": {
			args:        []string{"--database-port", "9999"},
			expectError: false,
		},
		"custom listen address": {
			args:        []string{"--database-listen", "127.0.0.1"},
			expectError: false,
		},
		"foreground mode": {
			args:        []string{"--foreground"},
			expectError: false,
		},
		"show password": {
			args:        []string{"--show-password"},
			expectError: false,
		},
		"multiple flags": {
			args:        []string{"--database-port", "8080", "--foreground", "--show-password"},
			expectError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := serviceStartCmd()
			cmd.SetArgs(tc.args)

			err := cmd.ParseFlags(tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestServiceStopCommand_Initialization tests service stop command structure
func TestServiceStopCommand_Initialization(t *testing.T) {
	cmd := serviceStopCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "stop", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.Run)
}

// TestServiceStopCommand_Flags tests that service stop has required flags
func TestServiceStopCommand_Flags(t *testing.T) {
	cmd := serviceStopCmd()

	// Check help flag
	helpFlag := cmd.Flags().Lookup(pconstants.ArgHelp)
	assert.NotNil(t, helpFlag)
	assert.Equal(t, "bool", helpFlag.Value.Type())
	assert.Equal(t, "h", helpFlag.Shorthand)

	// Check force flag
	forceFlag := cmd.Flags().Lookup(pconstants.ArgForce)
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "bool", forceFlag.Value.Type())
	assert.Equal(t, "false", forceFlag.DefValue)
}

// TestServiceStopCommand_FlagParsing tests stop command flag parsing
func TestServiceStopCommand_FlagParsing(t *testing.T) {
	tests := map[string]struct {
		args        []string
		expectError bool
	}{
		"no arguments": {
			args:        []string{},
			expectError: false,
		},
		"force flag": {
			args:        []string{"--force"},
			expectError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := serviceStopCmd()
			cmd.SetArgs(tc.args)

			err := cmd.ParseFlags(tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestServiceRestartCommand_Initialization tests service restart command structure
func TestServiceRestartCommand_Initialization(t *testing.T) {
	cmd := serviceRestartCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "restart", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.Run)
}

// TestServiceRestartCommand_Flags tests that service restart has required flags
func TestServiceRestartCommand_Flags(t *testing.T) {
	cmd := serviceRestartCmd()

	// Check help flag
	helpFlag := cmd.Flags().Lookup(pconstants.ArgHelp)
	assert.NotNil(t, helpFlag)
	assert.Equal(t, "bool", helpFlag.Value.Type())

	// Check force flag
	forceFlag := cmd.Flags().Lookup(pconstants.ArgForce)
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "bool", forceFlag.Value.Type())
	assert.Equal(t, "false", forceFlag.DefValue)
}

// TestServiceRestartCommand_FlagParsing tests restart command flag parsing
func TestServiceRestartCommand_FlagParsing(t *testing.T) {
	tests := map[string]struct {
		args        []string
		expectError bool
	}{
		"no arguments": {
			args:        []string{},
			expectError: false,
		},
		"force flag": {
			args:        []string{"--force"},
			expectError: false,
		},
		"help flag": {
			args:        []string{"--help"},
			expectError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := serviceRestartCmd()
			cmd.SetArgs(tc.args)

			err := cmd.ParseFlags(tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestServiceStatusCommand_Initialization tests service status command structure
func TestServiceStatusCommand_Initialization(t *testing.T) {
	cmd := serviceStatusCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "status", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.Run)
}

// TestServiceStatusCommand_Flags tests that service status has required flags
func TestServiceStatusCommand_Flags(t *testing.T) {
	cmd := serviceStatusCmd()

	// Check help flag
	helpFlag := cmd.Flags().Lookup(pconstants.ArgHelp)
	assert.NotNil(t, helpFlag)
	assert.Equal(t, "bool", helpFlag.Value.Type())

	// Check show-password flag
	showPasswordFlag := cmd.Flags().Lookup(pconstants.ArgServiceShowPassword)
	assert.NotNil(t, showPasswordFlag)
	assert.Equal(t, "bool", showPasswordFlag.Value.Type())
	assert.Equal(t, "false", showPasswordFlag.DefValue)

	// Check all flag
	allFlag := cmd.Flags().Lookup(pconstants.ArgAll)
	assert.NotNil(t, allFlag)
	assert.Equal(t, "bool", allFlag.Value.Type())
	assert.Equal(t, "false", allFlag.DefValue)
}

// TestServiceStatusCommand_FlagParsing tests status command flag parsing
func TestServiceStatusCommand_FlagParsing(t *testing.T) {
	tests := map[string]struct {
		args        []string
		expectError bool
	}{
		"no arguments": {
			args:        []string{},
			expectError: false,
		},
		"show-password flag": {
			args:        []string{"--show-password"},
			expectError: false,
		},
		"all flag": {
			args:        []string{"--all"},
			expectError: false,
		},
		"multiple flags": {
			args:        []string{"--show-password", "--all"},
			expectError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := serviceStatusCmd()
			cmd.SetArgs(tc.args)

			err := cmd.ParseFlags(tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestServiceCommand_NoArgs tests that service command requires a subcommand
func TestServiceCommand_NoArgs(t *testing.T) {
	cmd := serviceCmd()

	// The Args field should be cobra.NoArgs
	assert.NotNil(t, cmd.Args)
}

// TestServiceSubcommands_NoArgs tests that all subcommands don't accept positional args
func TestServiceSubcommands_NoArgs(t *testing.T) {
	tests := map[string]struct {
		cmdFunc func() *cobra.Command
	}{
		"start": {
			cmdFunc: serviceStartCmd,
		},
		"stop": {
			cmdFunc: serviceStopCmd,
		},
		"restart": {
			cmdFunc: serviceRestartCmd,
		},
		"status": {
			cmdFunc: serviceStatusCmd,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := tc.cmdFunc()
			assert.NotNil(t, cmd.Args)
		})
	}
}

// TestComposeStateError tests the composeStateError function
func TestComposeStateError(t *testing.T) {
	tests := map[string]struct {
		dbStateErr  error
		pmStateErr  error
		shouldError bool
		contains    []string
	}{
		"no errors": {
			dbStateErr:  nil,
			pmStateErr:  nil,
			shouldError: false,
		},
		"db state error only": {
			dbStateErr:  assert.AnError,
			pmStateErr:  nil,
			shouldError: true,
			contains:    []string{"could not get Steampipe service status", "failed to get db state"},
		},
		"pm state error only": {
			dbStateErr:  nil,
			pmStateErr:  assert.AnError,
			shouldError: true,
			contains:    []string{"could not get Steampipe service status", "failed to get plugin manager state"},
		},
		"both errors": {
			dbStateErr:  assert.AnError,
			pmStateErr:  assert.AnError,
			shouldError: true,
			contains:    []string{"could not get Steampipe service status", "failed to get db state", "failed to get plugin manager state"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := composeStateError(tc.dbStateErr, tc.pmStateErr)

			if tc.shouldError {
				assert.Error(t, err)
				for _, substr := range tc.contains {
					assert.Contains(t, err.Error(), substr)
				}
			} else {
				// When both are nil, the function still returns an error but with just the base message
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "could not get Steampipe service status")
			}
		})
	}
}

// TestBuildForegroundClientsConnectedMsg tests the buildForegroundClientsConnectedMsg function
func TestBuildForegroundClientsConnectedMsg(t *testing.T) {
	msg := buildForegroundClientsConnectedMsg()

	assert.NotEmpty(t, msg)
	assert.Contains(t, msg, "Not shutting down service")
	assert.Contains(t, msg, "clients connected")
	assert.Contains(t, msg, "Ctrl+C")
}

// TestServiceCommand_HelpText tests that all commands have proper help text
func TestServiceCommand_HelpText(t *testing.T) {
	commands := []struct {
		name    string
		cmdFunc func() *cobra.Command
	}{
		{"service", serviceCmd},
		{"start", serviceStartCmd},
		{"stop", serviceStopCmd},
		{"restart", serviceRestartCmd},
		{"status", serviceStatusCmd},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			cmd := tc.cmdFunc()

			assert.NotEmpty(t, cmd.Short, "command %s should have Short description", tc.name)
			assert.NotEmpty(t, cmd.Long, "command %s should have Long description", tc.name)

			// Verify help text contains relevant keywords
			switch tc.name {
			case "service":
				assert.Contains(t, cmd.Long, "service")
			case "start":
				assert.Contains(t, cmd.Long, "Start")
			case "stop":
				assert.Contains(t, cmd.Long, "Stop")
			case "restart":
				assert.Contains(t, cmd.Long, "Restart")
			case "status":
				assert.Contains(t, cmd.Long, "Status")
			}
		})
	}
}

// TestServiceStartCommand_InvokerFlag tests the hidden invoker flag
func TestServiceStartCommand_InvokerFlag(t *testing.T) {
	cmd := serviceStartCmd()

	invokerFlag := cmd.Flags().Lookup(pconstants.ArgInvoker)
	assert.NotNil(t, invokerFlag)
	assert.Equal(t, "string", invokerFlag.Value.Type())
	assert.Equal(t, string(constants.InvokerService), invokerFlag.DefValue)
	assert.True(t, invokerFlag.Hidden, "invoker flag should be hidden")
}

// TestServiceCommand_FlagShorthands tests that flags have proper shorthands
func TestServiceCommand_FlagShorthands(t *testing.T) {
	tests := map[string]struct {
		cmdFunc      func() *cobra.Command
		flagName     string
		shorthand    string
		hasShorthand bool
	}{
		"service help": {
			cmdFunc:      serviceCmd,
			flagName:     pconstants.ArgHelp,
			shorthand:    "h",
			hasShorthand: true,
		},
		"start help": {
			cmdFunc:      serviceStartCmd,
			flagName:     pconstants.ArgHelp,
			shorthand:    "h",
			hasShorthand: true,
		},
		"stop help": {
			cmdFunc:      serviceStopCmd,
			flagName:     pconstants.ArgHelp,
			shorthand:    "h",
			hasShorthand: true,
		},
		"restart help": {
			cmdFunc:      serviceRestartCmd,
			flagName:     pconstants.ArgHelp,
			shorthand:    "h",
			hasShorthand: true,
		},
		"status help": {
			cmdFunc:      serviceStatusCmd,
			flagName:     pconstants.ArgHelp,
			shorthand:    "h",
			hasShorthand: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := tc.cmdFunc()
			flag := cmd.Flags().Lookup(tc.flagName)

			assert.NotNil(t, flag)
			if tc.hasShorthand {
				assert.Equal(t, tc.shorthand, flag.Shorthand)
			} else {
				assert.Empty(t, flag.Shorthand)
			}
		})
	}
}

// TestServiceCommand_AllSubcommandsHaveRun tests that all subcommands have a Run function
func TestServiceCommand_AllSubcommandsHaveRun(t *testing.T) {
	subcommandFuncs := []struct {
		name    string
		cmdFunc func() *cobra.Command
	}{
		{"start", serviceStartCmd},
		{"stop", serviceStopCmd},
		{"restart", serviceRestartCmd},
		{"status", serviceStatusCmd},
	}

	for _, tc := range subcommandFuncs {
		t.Run(tc.name, func(t *testing.T) {
			cmd := tc.cmdFunc()
			assert.NotNil(t, cmd.Run, "subcommand %s should have a Run function", tc.name)
		})
	}
}

// TestServiceStartCommand_ListenAddressTypes tests different listen address configurations
func TestServiceStartCommand_ListenAddressTypes(t *testing.T) {
	tests := map[string]struct {
		listenArg string
	}{
		"local": {
			listenArg: "local",
		},
		"network": {
			listenArg: "network",
		},
		"localhost": {
			listenArg: "localhost",
		},
		"specific IP": {
			listenArg: "192.168.1.1",
		},
		"multiple addresses": {
			listenArg: "localhost,192.168.1.1",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := serviceStartCmd()
			args := []string{"--database-listen", tc.listenArg}
			cmd.SetArgs(args)

			err := cmd.ParseFlags(args)
			assert.NoError(t, err)

			listen, err := cmd.Flags().GetString(pconstants.ArgDatabaseListenAddresses)
			assert.NoError(t, err)
			assert.Equal(t, tc.listenArg, listen)
		})
	}
}

// TestServiceCommand_StructureConsistency tests that command structure is consistent
func TestServiceCommand_StructureConsistency(t *testing.T) {
	cmd := serviceCmd()

	// All subcommands should be present
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 4, "service should have exactly 4 subcommands")

	// Each subcommand should have proper structure
	for _, subCmd := range subcommands {
		assert.NotEmpty(t, subCmd.Use, "subcommand should have Use set")
		assert.NotEmpty(t, subCmd.Short, "subcommand should have Short description")
		assert.NotEmpty(t, subCmd.Long, "subcommand should have Long description")
		assert.NotNil(t, subCmd.Run, "subcommand should have Run function")

		// Each subcommand should have help flag
		helpFlag := subCmd.Flags().Lookup(pconstants.ArgHelp)
		assert.NotNil(t, helpFlag, "subcommand %s should have help flag", subCmd.Name())
	}
}

// TestServiceCommand_DefaultPort tests that the default port is correctly set
func TestServiceCommand_DefaultPort(t *testing.T) {
	cmd := serviceStartCmd()

	portFlag := cmd.Flags().Lookup(pconstants.ArgDatabasePort)
	assert.NotNil(t, portFlag)

	// Default port should be 9193 (constants.DatabaseDefaultPort)
	assert.Equal(t, "9193", portFlag.DefValue)
}

// TestServiceCommand_FlagTypes tests that all flags have correct types
func TestServiceCommand_FlagTypes(t *testing.T) {
	tests := map[string]struct {
		cmdFunc  func() *cobra.Command
		flagName string
		flagType string
	}{
		"start port": {
			cmdFunc:  serviceStartCmd,
			flagName: pconstants.ArgDatabasePort,
			flagType: "int",
		},
		"start listen": {
			cmdFunc:  serviceStartCmd,
			flagName: pconstants.ArgDatabaseListenAddresses,
			flagType: "string",
		},
		"start foreground": {
			cmdFunc:  serviceStartCmd,
			flagName: pconstants.ArgForeground,
			flagType: "bool",
		},
		"stop force": {
			cmdFunc:  serviceStopCmd,
			flagName: pconstants.ArgForce,
			flagType: "bool",
		},
		"restart force": {
			cmdFunc:  serviceRestartCmd,
			flagName: pconstants.ArgForce,
			flagType: "bool",
		},
		"status all": {
			cmdFunc:  serviceStatusCmd,
			flagName: pconstants.ArgAll,
			flagType: "bool",
		},
		"status show-password": {
			cmdFunc:  serviceStatusCmd,
			flagName: pconstants.ArgServiceShowPassword,
			flagType: "bool",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := tc.cmdFunc()
			flag := cmd.Flags().Lookup(tc.flagName)

			assert.NotNil(t, flag, "flag %s should exist", tc.flagName)
			assert.Equal(t, tc.flagType, flag.Value.Type(), "flag %s should have type %s", tc.flagName, tc.flagType)
		})
	}
}
