package steampipeconfig

import (
	"strings"
	"testing"
)

func TestValidationFailureString(t *testing.T) {
	testCases := []struct {
		name     string
		failure  ValidationFailure
		expected []string
	}{
		{
			name: "basic validation failure",
			failure: ValidationFailure{
				Plugin:             "hub.steampipe.io/plugins/turbot/aws@latest",
				ConnectionName:     "aws_prod",
				Message:            "invalid configuration",
				ShouldDropIfExists: false,
			},
			expected: []string{
				"Connection: aws_prod",
				"Plugin:     hub.steampipe.io/plugins/turbot/aws@latest",
				"Error:      invalid configuration",
			},
		},
		{
			name: "validation failure with drop flag",
			failure: ValidationFailure{
				Plugin:             "hub.steampipe.io/plugins/turbot/gcp@latest",
				ConnectionName:     "gcp_dev",
				Message:            "missing required field",
				ShouldDropIfExists: true,
			},
			expected: []string{
				"Connection: gcp_dev",
				"Plugin:     hub.steampipe.io/plugins/turbot/gcp@latest",
				"Error:      missing required field",
			},
		},
		{
			name: "validation failure with empty message",
			failure: ValidationFailure{
				Plugin:             "test_plugin",
				ConnectionName:     "test_conn",
				Message:            "",
				ShouldDropIfExists: false,
			},
			expected: []string{
				"Connection: test_conn",
				"Plugin:     test_plugin",
				"Error:      ",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.failure.String()

			for _, expected := range testCase.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s', got: %s", expected, result)
				}
			}
		})
	}
}

func TestValidationFailureStringFormat(t *testing.T) {
	failure := ValidationFailure{
		Plugin:             "test_plugin",
		ConnectionName:     "test_connection",
		Message:            "test error",
		ShouldDropIfExists: false,
	}

	result := failure.String()

	// Verify the format includes the expected labels
	if !strings.Contains(result, "Connection:") {
		t.Error("Expected result to contain 'Connection:' label")
	}

	if !strings.Contains(result, "Plugin:") {
		t.Error("Expected result to contain 'Plugin:' label")
	}

	if !strings.Contains(result, "Error:") {
		t.Error("Expected result to contain 'Error:' label")
	}

	// Verify the values are present
	if !strings.Contains(result, "test_connection") {
		t.Error("Expected result to contain connection name")
	}

	if !strings.Contains(result, "test_plugin") {
		t.Error("Expected result to contain plugin name")
	}

	if !strings.Contains(result, "test error") {
		t.Error("Expected result to contain error message")
	}
}
