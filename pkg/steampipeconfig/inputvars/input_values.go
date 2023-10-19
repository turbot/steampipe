package inputvars

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/terraform-components/terraform"
	"github.com/turbot/terraform-components/tfdiags"
	"github.com/zclconf/go-cty/cty/convert"

	"github.com/turbot/pipe-fittings/modconfig"
)

// SetVariableValues determines whether the given variable is a public variable and if so sets its value
func SetVariableValues(vv terraform.InputValues, m *modconfig.ModVariableMap) {
	for name, inputValue := range vv {
		variable, ok := m.PublicVariables[name]
		// if this variable does not exist in public variables, skip
		if !ok {
			// we should have already caught this
			continue
		}
		variable.SetInputValue(
			inputValue.Value,
			SourceTypeString(inputValue),
			inputValue.SourceRange)
	}
}

// CheckInputVariables ensures that variable values supplied at the UI conform
// to their corresponding declarations in configuration.
//
// The set of values is considered valid only if the returned diagnostics
// does not contain errors. A valid set of values may still produce warnings,
// which should be returned to the user.
func CheckInputVariables(vcs map[string]*modconfig.Variable, vs terraform.InputValues) tfdiags.Diagnostics {
	var diags tfdiags.Diagnostics

	for name, vc := range vcs {
		val, isSet := vs[name]
		if !isSet {
			// Always an error, since the caller should already have included
			// default values from the configuration in the values map.
			diags = diags.Append(tfdiags.Sourceless(
				tfdiags.Error,
				"Unassigned variable",
				fmt.Sprintf("The input variable %q has not been assigned a value. This is a bug in Steampipe; please report it in a GitHub issue.", name),
			))
			continue
		}

		wantType := vc.Type

		// A given value is valid if it can convert to the desired type.
		_, err := convert.Convert(val.Value, wantType)
		if err != nil {
			switch val.SourceType {
			case terraform.ValueFromConfig, terraform.ValueFromAutoFile, terraform.ValueFromNamedFile:
				// We have source location information for these.
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid value for input variable",
					Detail:   fmt.Sprintf("The given value is not valid for variable %q: %s.", name, err),
					Subject:  val.SourceRange.ToHCL().Ptr(),
				})
			case terraform.ValueFromEnvVar:
				diags = diags.Append(tfdiags.Sourceless(
					tfdiags.Error,
					"Invalid value for input variable",
					fmt.Sprintf("The environment variable SP_VAR_%s does not contain a valid value for variable %q: %s.", name, name, err),
				))
			case terraform.ValueFromCLIArg:
				diags = diags.Append(tfdiags.Sourceless(
					tfdiags.Error,
					"Invalid value for input variable",
					fmt.Sprintf("The argument --var=\"%s=...\" does not contain a valid value for variable %q: %s.", name, name, err),
				))
			case terraform.ValueFromInput:
				diags = diags.Append(tfdiags.Sourceless(
					tfdiags.Error,
					"Invalid value for input variable",
					fmt.Sprintf("The value entered for variable %q is not valid: %s.", name, err),
				))
			default:
				// The above gets us good coverage for the situations users
				// are likely to encounter with their own inputs. The other
				// cases are generally implementation bugs, so we'll just
				// use a generic error for these.
				diags = diags.Append(tfdiags.Sourceless(
					tfdiags.Error,
					"Invalid value for input variable",
					fmt.Sprintf("The value provided for variable %q is not valid: %s.", name, err),
				))
			}
		}
	}

	// Check for any variables that are assigned without being configured.
	// This is always an implementation error in the caller, because we
	// expect undefined variables to be caught during context construction
	// where there is better context to report it well.
	for name := range vs {
		if _, defined := vcs[name]; !defined {
			diags = diags.Append(tfdiags.Sourceless(
				tfdiags.Error,
				"Value assigned to undeclared variable",
				fmt.Sprintf("A value was assigned to an undeclared input variable %q.", name),
			))
		}
	}

	return diags
}

func SourceTypeString(v *terraform.InputValue) string {
	switch v.SourceType {
	case terraform.ValueFromConfig:
		return "config"
	case terraform.ValueFromAutoFile:
		return "auto file"
	case terraform.ValueFromNamedFile:
		return "name file"
	case terraform.ValueFromCLIArg:
		return "CLI arg"
	case terraform.ValueFromEnvVar:
		return "env var"
	case terraform.ValueFromInput:
		return "user input"
	default:
		return "unknown"
	}
}
