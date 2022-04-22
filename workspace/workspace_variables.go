package workspace

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform/tfdiags"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/inputvars"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
)

func (w *Workspace) getAllVariables(ctx context.Context, validate bool) (*modconfig.ModVariableMap, error) {
	// load all variable definitions
	variableMap, err := w.loadVariables()
	if err != nil {
		return nil, err
	}

	// now resolve all input variables

	inputVariables, err := w.getInputVariables(variableMap.AllVariables, variableMap.VariableAliases, validate)
	if err != nil {
		return nil, err
	}

	if validate {
		if err := validateVariables(ctx, variableMap.AllVariables, inputVariables); err != nil {
			return nil, err
		}
	}

	// now update the variables map with the input values
	// TODO for now we only support setting values for variables in the workspace mod
	//  or unique variables in dependency mods
	for name, inputValue := range inputVariables {
		variable := variableMap.AllVariables[name]
		variable.SetInputValue(
			inputValue.Value,
			inputValue.SourceTypeString(),
			inputValue.SourceRange)

		// set variable value string in our workspace map
		w.VariableValues[name], err = utils.CtyToString(inputValue.Value)
		if err != nil {
			return nil, err
		}
	}

	return variableMap, nil
}

func (w *Workspace) loadVariables() (*modconfig.ModVariableMap, error) {
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

	variableMap := modconfig.NewModVariableMap(mod, runCtx.LoadedDependencyMods)

	return variableMap, nil
}

func (w *Workspace) getInputVariables(variableMap map[string]*modconfig.Variable, variableAliases map[string]string, validate bool) (inputvars.InputValues, error) {
	variableFileArgs := viper.GetStringSlice(constants.ArgVarFile)
	variableArgs := viper.GetStringSlice(constants.ArgVariable)

	inputValuesUnparsed, diags := inputvars.CollectVariableValues(w.Path, variableFileArgs, variableArgs)
	if diags.HasErrors() {
		return nil, diags.Err()
	}

	if validate {
		if err := identifyMissingVariables(inputValuesUnparsed, variableMap); err != nil {
			return nil, err
		}
	}
	parsedValues, diags := inputvars.ParseVariableValues(inputValuesUnparsed, variableMap, variableAliases, validate)

	return parsedValues, diags.Err()
}

func validateVariables(ctx context.Context, variableMap map[string]*modconfig.Variable, variables inputvars.InputValues) error {
	diags := inputvars.CheckInputVariables(variableMap, variables)
	if diags.HasErrors() {
		displayValidationErrors(ctx, diags)
		// return empty error
		return modconfig.VariableValidationFailedError{}
	}
	return nil
}

func displayValidationErrors(ctx context.Context, diags tfdiags.Diagnostics) {
	fmt.Println()
	for i, diag := range diags {

		utils.ShowError(ctx, fmt.Errorf("%s", constants.Bold(diag.Description().Summary)))
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
