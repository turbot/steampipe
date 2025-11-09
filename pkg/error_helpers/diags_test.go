package error_helpers

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/terraform-components/tfdiags"
)

func TestDiagsToError(t *testing.T) {
	tests := map[string]struct {
		prefix       string
		diags        tfdiags.Diagnostics
		expectNil    bool
		checkStrings []string
	}{
		"no diagnostics": {
			prefix:    "Operation failed",
			diags:     tfdiags.Diagnostics{},
			expectNil: true,
		},
		"only warnings": {
			prefix: "Operation failed",
			diags: tfdiags.Diagnostics{
				tfdiags.Sourceless(
					tfdiags.Warning,
					"Deprecated feature",
					"This feature is deprecated",
				),
			},
			expectNil: true,
		},
		"single error": {
			prefix: "Configuration error",
			diags: tfdiags.Diagnostics{
				tfdiags.Sourceless(
					tfdiags.Error,
					"Invalid configuration",
					"The configuration is invalid",
				),
			},
			expectNil:    false,
			checkStrings: []string{"Configuration error", "Invalid configuration", "The configuration is invalid"},
		},
		"multiple errors": {
			prefix: "Multiple issues",
			diags: tfdiags.Diagnostics{
				tfdiags.Sourceless(
					tfdiags.Error,
					"Error 1",
					"First error detail",
				),
				tfdiags.Sourceless(
					tfdiags.Error,
					"Error 2",
					"Second error detail",
				),
			},
			expectNil:    false,
			checkStrings: []string{"Multiple issues", "Error 1", "Error 2"},
		},
		"error without detail": {
			prefix: "Operation failed",
			diags: tfdiags.Diagnostics{
				tfdiags.Sourceless(
					tfdiags.Error,
					"Simple error",
					"",
				),
			},
			expectNil:    false,
			checkStrings: []string{"Operation failed", "Simple error"},
		},
		"mixed warnings and errors": {
			prefix: "Issues found",
			diags: tfdiags.Diagnostics{
				tfdiags.Sourceless(
					tfdiags.Warning,
					"Warning message",
					"Warning detail",
				),
				tfdiags.Sourceless(
					tfdiags.Error,
					"Error message",
					"Error detail",
				),
			},
			expectNil:    false,
			checkStrings: []string{"Issues found", "Error message"},
		},
		"empty prefix": {
			prefix: "",
			diags: tfdiags.Diagnostics{
				tfdiags.Sourceless(
					tfdiags.Error,
					"Error message",
					"",
				),
			},
			expectNil:    false,
			checkStrings: []string{"Error message"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := DiagsToError(tc.prefix, tc.diags)

			if tc.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				errorString := result.Error()
				for _, expected := range tc.checkStrings {
					assert.Contains(t, errorString, expected)
				}
			}
		})
	}
}

func TestDiagsToErrorWithSource(t *testing.T) {
	// Test diagnostics with source information
	tests := map[string]struct {
		prefix       string
		diags        tfdiags.Diagnostics
		checkStrings []string
	}{
		"error with source": {
			prefix: "Parse error",
			diags: tfdiags.Diagnostics{
				tfdiags.WholeContainingBody(
					tfdiags.Error,
					"Syntax error",
					"Invalid syntax in configuration",
				),
			},
			checkStrings: []string{"Parse error", "Syntax error"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := DiagsToError(tc.prefix, tc.diags)

			assert.NotNil(t, result)
			errorString := result.Error()
			for _, expected := range tc.checkStrings {
				assert.Contains(t, errorString, expected)
			}
		})
	}
}

func TestDiagsToErrorDuplicateHandling(t *testing.T) {
	// Test that duplicate error messages are deduplicated
	diags := tfdiags.Diagnostics{
		tfdiags.Sourceless(
			tfdiags.Error,
			"Duplicate error",
			"This error appears multiple times",
		),
		tfdiags.Sourceless(
			tfdiags.Error,
			"Duplicate error",
			"This error appears multiple times",
		),
	}

	result := DiagsToError("Test", diags)
	assert.NotNil(t, result)

	// The error message should only contain one instance of the duplicate
	errorString := result.Error()
	assert.Contains(t, errorString, "Duplicate error")
}

func TestDiagsToErrorWithRange(t *testing.T) {
	// Test diagnostics with range information
	diags := tfdiags.Diagnostics{
		&hclDiagnostic{
			severity: tfdiags.Error,
			summary:  "Invalid value",
			detail:   "The value is not valid",
			subject: &hcl.Range{
				Filename: "config.hcl",
				Start:    hcl.Pos{Line: 10, Column: 5, Byte: 100},
				End:      hcl.Pos{Line: 10, Column: 15, Byte: 110},
			},
		},
	}

	result := DiagsToError("Configuration error", diags)
	assert.NotNil(t, result)

	errorString := result.Error()
	assert.Contains(t, errorString, "Invalid value")
	assert.Contains(t, errorString, "config.hcl")
}

func TestDiagsToErrorFormatting(t *testing.T) {
	// Test the output formatting
	diags := tfdiags.Diagnostics{
		tfdiags.Sourceless(
			tfdiags.Error,
			"First error",
			"First detail",
		),
		tfdiags.Sourceless(
			tfdiags.Error,
			"Second error",
			"Second detail",
		),
	}

	result := DiagsToError("Prefix", diags)
	assert.NotNil(t, result)

	errorString := result.Error()
	// Should contain the prefix
	assert.Contains(t, errorString, "Prefix")
	// Should contain both errors
	assert.Contains(t, errorString, "First error")
	assert.Contains(t, errorString, "Second error")
	// Should have newlines for multiple errors
	assert.Contains(t, errorString, "\n")
}

func TestDiagsToErrorNilDiagnostics(t *testing.T) {
	// Test with nil/empty diagnostics
	result := DiagsToError("Prefix", nil)
	assert.Nil(t, result)
}

// Helper type to create custom diagnostics with range information
type hclDiagnostic struct {
	severity tfdiags.Severity
	summary  string
	detail   string
	subject  *hcl.Range
}

func (d *hclDiagnostic) Severity() tfdiags.Severity {
	return d.severity
}

func (d *hclDiagnostic) Description() tfdiags.Description {
	return tfdiags.Description{
		Summary: d.summary,
		Detail:  d.detail,
	}
}

func (d *hclDiagnostic) Source() tfdiags.Source {
	var subject *tfdiags.SourceRange
	if d.subject != nil {
		subject = &tfdiags.SourceRange{
			Filename: d.subject.Filename,
			Start:    tfdiags.SourcePos{Line: d.subject.Start.Line, Column: d.subject.Start.Column, Byte: d.subject.Start.Byte},
			End:      tfdiags.SourcePos{Line: d.subject.End.Line, Column: d.subject.End.Column, Byte: d.subject.End.Byte},
		}
	}
	return tfdiags.Source{
		Subject: subject,
	}
}

func (d *hclDiagnostic) FromExpr() *tfdiags.FromExpr {
	return nil
}

func (d *hclDiagnostic) ExtraInfo() interface{} {
	return nil
}
