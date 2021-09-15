package workspace

import (
	"fmt"
	"sort"

	"github.com/hashicorp/terraform/tfdiags"
	"github.com/spf13/viper"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/input_vars"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"
	"github.com/turbot/steampipe/utils"
)

func (w *Workspace) getAllVariables() (map[string]*modconfig.Variable, error) {
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

	// now update the variables map with the input values
	for name, inputValue := range inputVariables {
		variable := variableMap[name]
		variable.SetInputValue(
			inputValue.Value,
			inputValue.SourceTypeString(),
			inputValue.SourceRange)
	}

	// parse all hcl files in the workspace folder (non recursively) and either parse or create a mod
	// it is valid for 0 or 1 mod to be defined (if no mod is defined, create a default one)
	// populate mod with all hcl resources we parse
	return variableMap, nil
}

func (w *Workspace) getInputVariables(variableMap map[string]*modconfig.Variable) (input_vars.InputValues, error) {
	variableFileArgs := viper.GetStringSlice(constants.ArgVarFile)
	variableArgs := viper.GetStringSlice(constants.ArgVariable)

	inputValuesUnparsed, diags := input_vars.CollectVariableValues(w.Path, variableFileArgs, variableArgs)
	if diags.HasErrors() {
		return nil, diags.Err()
	}

	if err := identifyMissingVariables(inputValuesUnparsed, variableMap); err != nil {
		return nil, err
	}
	parsedValues, diags := input_vars.ParseVariableValues(inputValuesUnparsed, variableMap)

	return parsedValues, diags.Err()
}

func validateVariables(variableMap map[string]*modconfig.Variable, variables input_vars.InputValues) error {
	diags := input_vars.CheckInputVariables(variableMap, variables)
	if diags.HasErrors() {
		displayValidationErrors(diags)
		// return empty error
		return modconfig.VariableValidationFailedError{}
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

func identifyMissingVariables(existing map[string]input_vars.UnparsedVariableValue, vcs map[string]*modconfig.Variable) error {
	var needed []*modconfig.Variable

	for name, vc := range vcs {
		if !vc.Required() {
			continue // We only prompt for required variables
		}
		if _, exists := existing[name]; !exists {
			needed = append(needed, vc)
		}
	}
	sort.SliceStable(needed, func(i, j int) bool {
		return needed[i].Name() < needed[j].Name()
	})
	if len(needed) > 0 {
		return modconfig.MissingVariableError{MissingVariables: needed}
	}
	return nil

}
