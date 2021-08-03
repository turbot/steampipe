package workspace

import (
	"fmt"

	"github.com/hashicorp/terraform/tfdiags"

	"github.com/turbot/steampipe/utils"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"
	"github.com/turbot/steampipe/steampipeconfig/tf"
)

func (w *Workspace) getAllVariables() (tf.InputValues, error) {
	variableMap, err := w.loadConfigVariables()
	if err != nil {
		return nil, err
	}

	// if there is a steampipe variables file, load it
	inputVariables, err := w.getInputVariables(tf.DefaultVariableValues(variableMap))
	if err != nil {
		return nil, err
	}

	if err := validateVariables(variableMap, inputVariables); err != nil {
		return nil, err
	}
	// parse all hcl files in the workspace folder (non recursively) and either parse or create a mod
	// it is valid for 0 or 1 mod to be defined (if no mod is defined, create a default one)
	// populate mod with all hcl resources we parse
	return inputVariables, nil
}

func (w *Workspace) loadConfigVariables() (map[string]*modconfig.Variable, error) {
	opts := &parse.ParseModOptions{
		Flags: parse.CreateDefaultMod,
		ListOptions: &filehelpers.ListOptions{
			// listFlag specifies whether to load files recursively
			Flags:   w.listFlag,
			Exclude: w.exclusions,
		},
	}

	return steampipeconfig.LoadVariables(w.Path, opts)
}

func validateVariables(variableMap map[string]*modconfig.Variable, variables tf.InputValues) error {
	//var diags hcl.Diagnostics
	//for name, val := range variableMap {
	//	inputVal, isSet := variables[name]
	//	if !isSet {
	//		// Always an error, since the caller should already have included
	//		// default values from the configuration in the values map.
	//		diags = diags.Append(&hcl.Diagnostic{
	//			Severity: hcl.DiagError,
	//			Summary:  "Unassigned variable",
	//			Detail:   fmt.Sprintf("The input variable %q has not been assigned a value. This is a bug in Steampipe; please report it in a GitHub issue.", name),
	//		})
	//		continue
	//	}
	//
	//	wantType := val.Type
	//
	//	// A given value is valid if it can convert to the desired type.
	//	_, err := convert.Convert(inputVal.Value, wantType)
	//	if err != nil {
	//		switch inputVal.SourceType {
	//		case tf.ValueFromConfig, tf.ValueFromAutoFile, tf.ValueFromNamedFile:
	//			// We have source location information for these.
	//			diags = diags.Append(&hcl.Diagnostic{
	//				Severity: hcl.DiagError,
	//				Summary:  "Invalid value for input variable",
	//				Detail:   fmt.Sprintf("The given value is not valid for variable %q: %s.", name, err),
	//				//Subject:  val.SourceRange.ToHCL().Ptr(),
	//			})
	//		case tf.ValueFromEnvVar:
	//			diags = diags.Append(&hcl.Diagnostic{
	//				Severity: hcl.DiagError,
	//				Summary:  "Invalid value for input variable",
	//				Detail:   fmt.Sprintf("The environment variable TF_VAR_%s does not contain a valid value for variable %q: %s.", name, name, err),
	//			})
	//		case tf.ValueFromCLIArg:
	//			diags = diags.Append(&hcl.Diagnostic{
	//				Severity: hcl.DiagError,
	//				Summary:  "Invalid value for input variable",
	//				Detail:   fmt.Sprintf("The argument -var=\"%s=...\" does not contain a valid value for variable %q: %s.", name, name, err),
	//			})
	//		case tf.ValueFromInput:
	//			diags = diags.Append(&hcl.Diagnostic{
	//				Severity: hcl.DiagError,
	//				Summary:  "Invalid value for input variable",
	//				Detail:   fmt.Sprintf("The value entered for variable %q is not valid: %s.", name, err),
	//			})
	//		default:
	//			// The above gets us good coverage for the situations users
	//			// are likely to encounter with their own inputs. The other
	//			// cases are generally implementation bugs, so we'll just
	//			// use a generic error for these.
	//			diags = diags.Append(&hcl.Diagnostic{
	//				Severity: hcl.DiagError,
	//				Summary:  "Invalid value for input variable",
	//				Detail:   fmt.Sprintf("The value provided for variable %q is not valid: %s.", name, err),
	//			})
	//		}
	//	}
	//}
	diags := tf.CheckInputVariables(variableMap, variables)
	if diags.HasErrors() {
		displayValidationErrors(diags)
		return fmt.Errorf("%d validation errors occurred", len(diags))
	}
	return nil
}

func displayValidationErrors(diags tfdiags.Diagnostics) {
	fmt.Println()
	for _, diag := range diags {
		utils.ShowError(fmt.Errorf("%s", constants.Bold(diag.Description().Summary)))
		fmt.Println(diag.Description().Detail)
		fmt.Println()

		// TODO range if there is one
	}
}

func (w *Workspace) getInputVariables(defaultValues tf.InputValues) (tf.InputValues, error) {
	// to do add in default values
	// find all .spvars files
	res, err := w.loadFileVariables()
	if err != nil {
		return nil, err
	}
	// now add in default values
	for k, v := range defaultValues {
		if _, hasValue := res[k]; !hasValue {
			res[k] = v
		}
	}
	return res, nil

}

func (w *Workspace) loadFileVariables() (tf.InputValues, error) {
	opts := &filehelpers.ListOptions{
		// listFlag specifies whether to load files recursively
		Flags:   filehelpers.Files,
		Include: []string{fmt.Sprintf("*%s", constants.VariablesExtension)},
	}
	variablesFiles, err := filehelpers.ListFiles(w.Path, opts)

	if err != nil {
		return nil, err
	}
	fileData, diags := parse.LoadFileData(variablesFiles...)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load all variables files", diags)
	}
	fileVariables, err := parse.ParseVariables(fileData)
	if err != nil {
		return nil, err
	}

	var res = tf.InputValues{}
	for k, v := range fileVariables {
		res[k] = &tf.InputValue{
			Value: v,
			// TODO named vs auto
			SourceType: tf.ValueFromNamedFile,
			// TODO get range out of ParseVariables
			//SourceRange: nil,
		}
	}
	return res, nil
}
