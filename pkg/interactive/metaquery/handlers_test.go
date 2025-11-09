package metaquery

import (
	"context"
	"testing"

	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/cmdconfig"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

func TestHandle(t *testing.T) {
	tests := map[string]struct {
		query       string
		expectError bool
	}{
		"help command": {
			query:       constants.CmdHelp,
			expectError: false,
		},
		"unknown command": {
			query:       ".unknown",
			expectError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			input := &HandlerInput{
				Query: tt.query,
				// Provide mock ClosePrompt for handlers that need it
				ClosePrompt: func() {},
			}

			err := Handle(context.Background(), input)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Test doExit separately with proper mocking
func TestDoExit(t *testing.T) {
	called := false
	input := &HandlerInput{
		Query: constants.CmdExit,
		ClosePrompt: func() {
			called = true
		},
	}

	err := doExit(context.Background(), input)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !called {
		t.Error("Expected ClosePrompt to be called")
	}
}

func TestSetHeader(t *testing.T) {
	// Save original state
	origValue := viper.Get(pconstants.ArgHeader)
	defer func() {
		if origValue != nil {
			cmdconfig.Viper().Set(pconstants.ArgHeader, origValue)
		}
	}()

	tests := map[string]struct {
		arg      string
		expected bool
	}{
		"true value": {
			arg:      "true",
			expected: true,
		},
		"false value": {
			arg:      "false",
			expected: false,
		},
		"on value": {
			arg:      "on",
			expected: true,
		},
		"off value": {
			arg:      "off",
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			input := &HandlerInput{
				Query: constants.CmdHeaders + " " + tt.arg,
			}

			err := setHeader(context.Background(), input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			result := cmdconfig.Viper().GetBool(pconstants.ArgHeader)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestSetMultiLine(t *testing.T) {
	// Save original state
	origValue := viper.Get(pconstants.ArgMultiLine)
	defer func() {
		if origValue != nil {
			cmdconfig.Viper().Set(pconstants.ArgMultiLine, origValue)
		}
	}()

	tests := map[string]struct {
		arg      string
		expected bool
	}{
		"true value": {
			arg:      "true",
			expected: true,
		},
		"false value": {
			arg:      "false",
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			input := &HandlerInput{
				Query: constants.CmdMulti + " " + tt.arg,
			}

			err := setMultiLine(context.Background(), input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			result := cmdconfig.Viper().GetBool(pconstants.ArgMultiLine)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestSetTiming(t *testing.T) {
	// Save original state
	origValue := viper.Get(pconstants.ArgTiming)
	defer func() {
		if origValue != nil {
			cmdconfig.Viper().Set(pconstants.ArgTiming, origValue)
		}
	}()

	tests := map[string]struct {
		args        []string
		expectValue string
	}{
		"on value": {
			args:        []string{"on"},
			expectValue: "on",
		},
		"off value": {
			args:        []string{"off"},
			expectValue: "off",
		},
		"verbose value": {
			args:        []string{"verbose"},
			expectValue: "verbose",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			input := &HandlerInput{
				Query: constants.CmdTiming + " " + tt.args[0],
			}

			err := setTiming(context.Background(), input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			result := cmdconfig.Viper().GetString(pconstants.ArgTiming)
			if result != tt.expectValue {
				t.Errorf("Expected %v, got %v", tt.expectValue, result)
			}
		})
	}
}

func TestGetCmdAndArgs_Extended(t *testing.T) {
	tests := map[string]struct {
		input       string
		expectedCmd string
		expectedArgs []string
	}{
		"simple command": {
			input:       ".help",
			expectedCmd: ".help",
			expectedArgs: []string{},
		},
		"command with single arg": {
			input:       ".output json",
			expectedCmd: ".output",
			expectedArgs: []string{"json"},
		},
		"command with multiple args": {
			input:       ".inspect aws_ec2_instance",
			expectedCmd: ".inspect",
			expectedArgs: []string{"aws_ec2_instance"},
		},
		"command with quoted args": {
			input:       `.inspect "my table"`,
			expectedCmd: ".inspect",
			expectedArgs: []string{"my table"},
		},
		"command with mixed args": {
			input:       `.cmd arg1 "arg 2" arg3`,
			expectedCmd: ".cmd",
			expectedArgs: []string{"arg1", "arg 2", "arg3"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmd, args := getCmdAndArgs(tt.input)

			if cmd != tt.expectedCmd {
				t.Errorf("Expected cmd %s, got %s", tt.expectedCmd, cmd)
			}

			if len(args) != len(tt.expectedArgs) {
				t.Errorf("Expected %d args, got %d", len(tt.expectedArgs), len(args))
			}

			for i, arg := range args {
				if i >= len(tt.expectedArgs) {
					break
				}
				if arg != tt.expectedArgs[i] {
					t.Errorf("Expected arg[%d] to be %s, got %s", i, tt.expectedArgs[i], arg)
				}
			}
		})
	}
}

func TestSetViperConfigFromArg(t *testing.T) {
	// Save original state
	origOutput := viper.Get(pconstants.ArgOutput)
	defer func() {
		if origOutput != nil {
			cmdconfig.Viper().Set(pconstants.ArgOutput, origOutput)
		}
	}()

	handler := setViperConfigFromArg(pconstants.ArgOutput)
	input := &HandlerInput{
		Query: constants.CmdOutput + " json",
	}

	err := handler(context.Background(), input)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	result := cmdconfig.Viper().GetString(pconstants.ArgOutput)
	if result != "json" {
		t.Errorf("Expected 'json', got '%s'", result)
	}
}

func TestSetAutoComplete(t *testing.T) {
	// Save original state
	origValue := viper.Get(pconstants.ArgAutoComplete)
	defer func() {
		if origValue != nil {
			cmdconfig.Viper().Set(pconstants.ArgAutoComplete, origValue)
		}
	}()

	tests := map[string]struct {
		arg      string
		expected bool
	}{
		"true value": {
			arg:      "true",
			expected: true,
		},
		"false value": {
			arg:      "false",
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			input := &HandlerInput{
				Query: constants.CmdAutoComplete + " " + tt.arg,
			}

			err := setAutoComplete(context.Background(), input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			result := cmdconfig.Viper().GetBool(pconstants.ArgAutoComplete)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMetaQueryDefinitionsExists(t *testing.T) {
	// Test that all expected metaquery definitions exist
	expectedCommands := []string{
		constants.CmdHelp,
		constants.CmdExit,
		constants.CmdQuit,
		constants.CmdTableList,
		constants.CmdSeparator,
		constants.CmdHeaders,
		constants.CmdMulti,
		constants.CmdTiming,
		constants.CmdOutput,
		constants.CmdCache,
		constants.CmdCacheTtl,
		constants.CmdInspect,
		constants.CmdConnections,
		constants.CmdClear,
		constants.CmdSearchPath,
		constants.CmdSearchPathPrefix,
		constants.CmdAutoComplete,
	}

	for _, cmd := range expectedCommands {
		t.Run(cmd, func(t *testing.T) {
			def, exists := metaQueryDefinitions[cmd]
			if !exists {
				t.Errorf("Expected metaquery definition for %s to exist", cmd)
			}
			if def.handler == nil {
				t.Errorf("Expected handler for %s to be non-nil", cmd)
			}
			if def.validator == nil {
				t.Errorf("Expected validator for %s to be non-nil", cmd)
			}
		})
	}
}

func TestShowTimingFlag(t *testing.T) {
	// Save original state
	origValue := viper.Get(pconstants.ArgTiming)
	defer func() {
		if origValue != nil {
			cmdconfig.Viper().Set(pconstants.ArgTiming, origValue)
		}
	}()

	// Set a known timing value
	cmdconfig.Viper().Set(pconstants.ArgTiming, "on")

	// Just verify it doesn't panic
	showTimingFlag()
}
