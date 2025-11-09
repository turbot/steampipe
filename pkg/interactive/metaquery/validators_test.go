package metaquery

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/cmdconfig"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

func TestValidate(t *testing.T) {
	tests := map[string]struct {
		query           string
		expectError     bool
		expectShouldRun bool
		expectMessage   bool
	}{
		"valid help command": {
			query:           constants.CmdHelp,
			expectError:     false,
			expectShouldRun: true,
		},
		"valid exit command": {
			query:           constants.CmdExit,
			expectError:     false,
			expectShouldRun: true,
		},
		"valid output command": {
			query:           constants.CmdOutput + " json",
			expectError:     false,
			expectShouldRun: true,
		},
		"invalid output value": {
			query:       constants.CmdOutput + " invalid",
			expectError: true,
		},
		"unknown command": {
			query:       ".unknown",
			expectError: true,
		},
		"query with semicolon": {
			query:           constants.CmdHelp + ";",
			expectError:     false,
			expectShouldRun: true,
		},
		"headers without args": {
			query:           constants.CmdHeaders,
			expectError:     false,
			expectShouldRun: false,
			expectMessage:   true,
		},
		"timing without args": {
			query:           constants.CmdTiming,
			expectError:     false,
			expectShouldRun: true, // Timing displays status and returns true
			expectMessage:   false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := Validate(tt.query)

			if tt.expectError && result.Err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && result.Err != nil {
				t.Errorf("Unexpected error: %v", result.Err)
			}

			if result.ShouldRun != tt.expectShouldRun {
				t.Errorf("Expected ShouldRun %v, got %v", tt.expectShouldRun, result.ShouldRun)
			}

			if tt.expectMessage && result.Message == "" {
				t.Error("Expected message but got empty string")
			}
		})
	}
}

func TestAtLeastNArgs(t *testing.T) {
	tests := map[string]struct {
		n           int
		args        []string
		expectError bool
	}{
		"zero args required, zero provided": {
			n:           0,
			args:        []string{},
			expectError: false,
		},
		"one arg required, zero provided": {
			n:           1,
			args:        []string{},
			expectError: true,
		},
		"one arg required, one provided": {
			n:           1,
			args:        []string{"arg1"},
			expectError: false,
		},
		"one arg required, two provided": {
			n:           1,
			args:        []string{"arg1", "arg2"},
			expectError: false,
		},
		"two args required, one provided": {
			n:           2,
			args:        []string{"arg1"},
			expectError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			validator := atLeastNArgs(tt.n)
			result := validator(tt.args)

			if tt.expectError && result.Err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && result.Err != nil {
				t.Errorf("Unexpected error: %v", result.Err)
			}
		})
	}
}

func TestAtMostNArgs(t *testing.T) {
	tests := map[string]struct {
		n           int
		args        []string
		expectError bool
	}{
		"one arg allowed, zero provided": {
			n:           1,
			args:        []string{},
			expectError: false,
		},
		"one arg allowed, one provided": {
			n:           1,
			args:        []string{"arg1"},
			expectError: false,
		},
		"one arg allowed, two provided": {
			n:           1,
			args:        []string{"arg1", "arg2"},
			expectError: true,
		},
		"zero args allowed, one provided": {
			n:           0,
			args:        []string{"arg1"},
			expectError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			validator := atMostNArgs(tt.n)
			result := validator(tt.args)

			if tt.expectError && result.Err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && result.Err != nil {
				t.Errorf("Unexpected error: %v", result.Err)
			}
		})
	}
}

func TestExactlyNArgs(t *testing.T) {
	tests := map[string]struct {
		n           int
		args        []string
		expectError bool
	}{
		"zero args required, zero provided": {
			n:           0,
			args:        []string{},
			expectError: false,
		},
		"one arg required, one provided": {
			n:           1,
			args:        []string{"arg1"},
			expectError: false,
		},
		"one arg required, zero provided": {
			n:           1,
			args:        []string{},
			expectError: true,
		},
		"one arg required, two provided": {
			n:           1,
			args:        []string{"arg1", "arg2"},
			expectError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			validator := exactlyNArgs(tt.n)
			result := validator(tt.args)

			if tt.expectError && result.Err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && result.Err != nil {
				t.Errorf("Unexpected error: %v", result.Err)
			}
		})
	}
}

func TestNoArgs(t *testing.T) {
	tests := map[string]struct {
		args        []string
		expectError bool
	}{
		"no args provided": {
			args:        []string{},
			expectError: false,
		},
		"one arg provided": {
			args:        []string{"arg1"},
			expectError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := noArgs(tt.args)

			if tt.expectError && result.Err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && result.Err != nil {
				t.Errorf("Unexpected error: %v", result.Err)
			}
		})
	}
}

func TestAllowedArgValues(t *testing.T) {
	tests := map[string]struct {
		caseSensitive bool
		allowedValues []string
		args          []string
		expectError   bool
	}{
		"valid arg case-insensitive": {
			caseSensitive: false,
			allowedValues: []string{"json", "csv", "table"},
			args:          []string{"JSON"},
			expectError:   false,
		},
		"invalid arg case-insensitive": {
			caseSensitive: false,
			allowedValues: []string{"json", "csv", "table"},
			args:          []string{"invalid"},
			expectError:   true,
		},
		"valid arg case-sensitive": {
			caseSensitive: true,
			allowedValues: []string{"JSON", "CSV"},
			args:          []string{"JSON"},
			expectError:   false,
		},
		"invalid arg case-sensitive": {
			caseSensitive: true,
			allowedValues: []string{"JSON", "CSV"},
			args:          []string{"json"},
			expectError:   true,
		},
		"multiple valid args": {
			caseSensitive: false,
			allowedValues: []string{"on", "off", "verbose"},
			args:          []string{"on"},
			expectError:   false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			validator := allowedArgValues(tt.caseSensitive, tt.allowedValues...)
			result := validator(tt.args)

			if tt.expectError && result.Err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && result.Err != nil {
				t.Errorf("Unexpected error: %v", result.Err)
			}
		})
	}
}

func TestBooleanValidator(t *testing.T) {
	// Save and restore original state
	origHeader := viper.Get(pconstants.ArgHeader)
	defer func() {
		if origHeader != nil {
			cmdconfig.Viper().Set(pconstants.ArgHeader, origHeader)
		}
	}()

	// Set initial state
	cmdconfig.Viper().Set(pconstants.ArgHeader, false)

	tests := map[string]struct {
		args            []string
		expectError     bool
		expectShouldRun bool
		expectMessage   bool
	}{
		"no args - show status": {
			args:            []string{},
			expectError:     false,
			expectShouldRun: false,
			expectMessage:   true,
		},
		"valid arg on": {
			args:            []string{"on"},
			expectError:     false,
			expectShouldRun: true,
		},
		"valid arg off": {
			args:            []string{"off"},
			expectError:     false,
			expectShouldRun: true,
		},
		"too many args": {
			args:        []string{"on", "off"},
			expectError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			validator := booleanValidator(constants.CmdHeaders, pconstants.ArgHeader, validatorFromArgsOf(constants.CmdHeaders))
			result := validator(tt.args)

			if tt.expectError && result.Err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && result.Err != nil {
				t.Errorf("Unexpected error: %v", result.Err)
			}

			if result.ShouldRun != tt.expectShouldRun {
				t.Errorf("Expected ShouldRun %v, got %v", tt.expectShouldRun, result.ShouldRun)
			}

			if tt.expectMessage && result.Message == "" {
				t.Error("Expected message but got empty string")
			}
		})
	}
}

func TestComposeValidator(t *testing.T) {
	tests := map[string]struct {
		validators  []validator
		args        []string
		expectError bool
	}{
		"all validators pass": {
			validators: []validator{
				exactlyNArgs(1),
				allowedArgValues(false, "on", "off"),
			},
			args:        []string{"on"},
			expectError: false,
		},
		"first validator fails": {
			validators: []validator{
				exactlyNArgs(1),
				allowedArgValues(false, "on", "off"),
			},
			args:        []string{"on", "off"},
			expectError: true,
		},
		"second validator fails": {
			validators: []validator{
				exactlyNArgs(1),
				allowedArgValues(false, "on", "off"),
			},
			args:        []string{"invalid"},
			expectError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			validator := composeValidator(tt.validators...)
			result := validator(tt.args)

			if tt.expectError && result.Err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && result.Err != nil {
				t.Errorf("Unexpected error: %v", result.Err)
			}
		})
	}
}

func TestTitleSentenceCase(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"simple word": {
			input:    "headers",
			expected: "Headers",
		},
		"hyphenated word": {
			input:    "multi-line",
			expected: "Multi-line",
		},
		"already capitalized": {
			input:    "AutoComplete",
			expected: "Autocomplete",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := titleSentenceCase(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestValidatorFromArgsOf(t *testing.T) {
	tests := map[string]struct {
		cmd         string
		args        []string
		expectError bool
	}{
		"output command with valid arg": {
			cmd:         constants.CmdOutput,
			args:        []string{"json"},
			expectError: false,
		},
		"output command with invalid arg": {
			cmd:         constants.CmdOutput,
			args:        []string{"invalid"},
			expectError: true,
		},
		"timing command with valid arg": {
			cmd:         constants.CmdTiming,
			args:        []string{"on"},
			expectError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			validator := validatorFromArgsOf(tt.cmd)
			result := validator(tt.args)

			if tt.expectError && result.Err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && result.Err != nil {
				t.Errorf("Unexpected error: %v", result.Err)
			}
		})
	}
}

func TestValidationResultErrorMessage(t *testing.T) {
	// Test that error messages contain useful information
	tests := map[string]struct {
		validator       validator
		args            []string
		expectedInError string
	}{
		"too many args error contains count": {
			validator:       exactlyNArgs(1),
			args:            []string{"arg1", "arg2"},
			expectedInError: "got 2",
		},
		"invalid arg error contains value": {
			validator:       allowedArgValues(false, "on", "off"),
			args:            []string{"invalid"},
			expectedInError: "invalid",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.validator(tt.args)
			if result.Err == nil {
				t.Fatal("Expected error but got none")
			}

			if !strings.Contains(result.Err.Error(), tt.expectedInError) {
				t.Errorf("Expected error to contain '%s', got: %s", tt.expectedInError, result.Err.Error())
			}
		})
	}
}
