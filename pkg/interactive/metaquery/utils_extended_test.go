package metaquery

import (
	"testing"

	"github.com/turbot/steampipe/v2/pkg/constants"
)

func TestIsMetaQuery(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"help command": {
			input:    constants.CmdHelp,
			expected: true,
		},
		"exit command": {
			input:    constants.CmdExit,
			expected: true,
		},
		"timing command": {
			input:    constants.CmdTiming,
			expected: true,
		},
		"output command": {
			input:    constants.CmdOutput + " json",
			expected: true,
		},
		"SQL query": {
			input:    "SELECT * FROM aws_ec2_instance",
			expected: false,
		},
		"empty string": {
			input:    "",
			expected: false,
		},
		"dot only": {
			input:    ".",
			expected: false,
		},
		"invalid metaquery": {
			input:    ".invalid",
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsMetaQuery(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for input '%s'", tt.expected, result, tt.input)
			}
		})
	}
}

func TestBuildTable(t *testing.T) {
	tests := map[string]struct {
		rows      [][]string
		autoMerge bool
	}{
		"simple table": {
			rows: [][]string{
				{"Name", "Type"},
				{"table1", "BASE TABLE"},
				{"table2", "VIEW"},
			},
			autoMerge: false,
		},
		"single column": {
			rows: [][]string{
				{"Name"},
				{"value1"},
				{"value2"},
			},
			autoMerge: false,
		},
		"empty table": {
			rows:      [][]string{},
			autoMerge: false,
		},
		"auto merge enabled": {
			rows: [][]string{
				{"Col1", "Col2"},
				{"val1", "val2"},
			},
			autoMerge: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// buildTable returns a string
			// Just verify it doesn't panic and returns something
			result := buildTable(tt.rows, tt.autoMerge)
			// For non-empty rows, we should get a non-empty string
			if len(tt.rows) > 0 && result == "" {
				t.Error("Expected non-empty result for non-empty rows")
			}
		})
	}
}

func TestGetArguments(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected []string
	}{
		"no arguments": {
			input:    constants.CmdHelp,
			expected: []string{},
		},
		"single argument": {
			input:    constants.CmdOutput + " json",
			expected: []string{"json"},
		},
		"multiple arguments": {
			input:    constants.CmdInspect + " aws_ec2_instance",
			expected: []string{"aws_ec2_instance"},
		},
		"quoted argument": {
			input:    `.inspect "my table"`,
			expected: []string{"my table"},
		},
		"mixed arguments": {
			input:    `.cmd arg1 "arg 2" arg3`,
			expected: []string{"arg1", "arg 2", "arg3"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := getArguments(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d arguments, got %d", len(tt.expected), len(result))
			}

			for i, arg := range result {
				if i >= len(tt.expected) {
					break
				}
				if arg != tt.expected[i] {
					t.Errorf("Expected arg[%d] to be '%s', got '%s'", i, tt.expected[i], arg)
				}
			}
		})
	}
}
