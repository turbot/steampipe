package workspace

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/hashicorp/terraform/tfdiags"

	"github.com/turbot/steampipe/utils"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"
	"github.com/turbot/steampipe/steampipeconfig/tf"
)

func (w *Workspace) getAllVariables() (tf.InputValues, error) {
	opts := &parse.ParseModOptions{
		Flags: parse.CreateDefaultMod,
		ListOptions: &filehelpers.ListOptions{
			// listFlag specifies whether to load files recursively
			Flags:   w.listFlag,
			Exclude: w.exclusions,
		},
	}
	variableMap, err := steampipeconfig.LoadVariables(w.Path, opts)
	if err != nil {
		return nil, err
	}

	// if there is a steampipe variables file, load it
	inputVariables, err := w.getInputVariables(variableMap)
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

func (w *Workspace) getInputVariables(variableMap map[string]*modconfig.Variable) (tf.InputValues, error) {
	variableFileArgs := viper.GetStringSlice(constants.ArgVarFile)
	variableArgs := viper.GetStringSlice(constants.ArgVariable)

	inputValuesUnparsed, diags := tf.CollectVariableValues(w.Path, variableFileArgs, variableArgs)
	if diags.HasErrors() {
		return nil, diags.Err()
	}

	parsedValues, diags := tf.ParseVariableValues(inputValuesUnparsed, variableMap)

	return parsedValues, diags.Err()
}

func validateVariables(variableMap map[string]*modconfig.Variable, variables tf.InputValues) error {
	diags := tf.CheckInputVariables(variableMap, variables)
	if diags.HasErrors() {
		displayValidationErrors(diags)
		return fmt.Errorf("%d validation errors occurred", len(diags))
	}
	return nil
}

func displayValidationErrors(diags tfdiags.Diagnostics) {
	fmt.Println()
	for i, diag := range diags {

		utils.ShowError(fmt.Errorf("%s", constants.Bold(diag.Description().Summary)))
		fmt.Println(diag.Description().Detail)
		if i < len(diags)-1 {
			fmt.Println()
		}
		// TODO range if there is one
	}
}

//
//func interactiveCollectVariables(existing map[string]backend.UnparsedVariableValue, vcs map[string]*configs.Variable, uiInput terraform.UIInput) map[string]backend.UnparsedVariableValue {
//	var needed []string
//	if b.OpInput && uiInput != nil {
//		for name, vc := range vcs {
//			if !vc.Required() {
//				continue // We only prompt for required variables
//			}
//			if _, exists := existing[name]; !exists {
//				needed = append(needed, name)
//			}
//		}
//	} else {
//		log.Print("[DEBUG] backend/local: Skipping interactive prompts for variables because input is disabled")
//	}
//	if len(needed) == 0 {
//		return existing
//	}
//
//	log.Printf("[DEBUG] backend/local: will prompt for input of unset required variables %s", needed)
//
//	// If we get here then we're planning to prompt for at least one additional
//	// variable's value.
//	sort.Strings(needed) // prompt in lexical order
//	ret := make(map[string]backend.UnparsedVariableValue, len(vcs))
//	for k, v := range existing {
//		ret[k] = v
//	}
//	for _, name := range needed {
//		vc := vcs[name]
//		rawValue, err := uiInput.Input(ctx, &terraform.InputOpts{
//			Id:          fmt.Sprintf("var.%s", name),
//			Query:       fmt.Sprintf("var.%s", name),
//			Description: vc.Description,
//		})
//		if err != nil {
//			// Since interactive prompts are best-effort, we'll just continue
//			// here and let subsequent validation report this as a variable
//			// not specified.
//			log.Printf("[WARN] backend/local: Failed to request user input for variable %q: %s", name, err)
//			continue
//		}
//		ret[name] = unparsedInteractiveVariableValue{Name: name, RawValue: rawValue}
//	}
//	return ret
//}
