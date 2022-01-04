package workspace

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform/tfdiags"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/inputvars"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
)

func (w *Workspace) getAllVariables() (map[string]*modconfig.Variable, error) {
	// build options used to load workspace
	runCtx, err := w.getRunContext()
	if err != nil {
		return nil, err
	}
	// only load variables blocks
	runCtx.BlockTypes = []string{modconfig.BlockTypeVariable}
	mod, err := steampipeconfig.LoadMod(w.Path, runCtx)
	if err != nil {
		return nil, err
	}

	// TACTICAL - as the tf derived code builds a map keyed by the short variable name, do the same
	variableMap := make(map[string]*modconfig.Variable)
	for k, v := range mod.Variables {
		name := strings.Split(k, ".")[1]
		variableMap[name] = v
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

	return variableMap, nil
}

func (w *Workspace) getInputVariables(variableMap map[string]*modconfig.Variable) (inputvars.InputValues, error) {
	variableFileArgs := viper.GetStringSlice(constants.ArgVarFile)
	variableArgs := viper.GetStringSlice(constants.ArgVariable)

	inputValuesUnparsed, diags := inputvars.CollectVariableValues(w.Path, variableFileArgs, variableArgs)
	if diags.HasErrors() {
		return nil, diags.Err()
	}

	if err := identifyMissingVariables(inputValuesUnparsed, variableMap); err != nil {
		return nil, err
	}
	parsedValues, diags := inputvars.ParseVariableValues(inputValuesUnparsed, variableMap)

	return parsedValues, diags.Err()
}

func validateVariables(variableMap map[string]*modconfig.Variable, variables inputvars.InputValues) error {
	diags := inputvars.CheckInputVariables(variableMap, variables)
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

func identifyMissingVariables(existing map[string]inputvars.UnparsedVariableValue, vcs map[string]*modconfig.Variable) error {
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
