package workspace

import (
	"context"
	"fmt"
	"sort"

	"github.com/turbot/steampipe/steampipeconfig/parse"

	"github.com/hashicorp/terraform/tfdiags"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/inputvars"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
)

func (w *Workspace) getAllVariables(ctx context.Context, runCtx *parse.RunContext, variableMap *modconfig.ModVariableMap, validate bool) (*modconfig.ModVariableMap, error) {

	// now resolve all input variables

	inputVariables, err := w.getInputVariables(variableMap.AllVariables, validate, runCtx)
	if err != nil {
		return nil, err
	}

	if validate {
		if err := validateVariables(ctx, variableMap.AllVariables, inputVariables); err != nil {
			return nil, err
		}
	}

	// now update the variables map with the input values
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

func (w *Workspace) loadVariableDefinitions() (*modconfig.ModVariableMap, *modconfig.Mod, error) {
	// build options used to load workspace
	runCtx, err := w.getRunContext()
	if err != nil {
		return nil, nil, err
	}
	// only load mod and variables blocks
	runCtx.BlockTypes = []string{modconfig.BlockTypeVariable}
	mod, err := steampipeconfig.LoadMod(w.Path, runCtx)
	if err != nil {
		return nil, nil, err
	}

	variableMap := modconfig.NewModVariableMap(mod, runCtx.LoadedDependencyMods)

	return variableMap, mod, nil
}

func (w *Workspace) getInputVariables(variableMap map[string]*modconfig.Variable, validate bool, runCtx *parse.RunContext) (inputvars.InputValues, error) {
	variableFileArgs := viper.GetStringSlice(constants.ArgVarFile)
	variableArgs := viper.GetStringSlice(constants.ArgVariable)

	inputValuesUnparsed, err := inputvars.CollectVariableValues(w.Path, variableFileArgs, variableArgs)
	if err != nil {
		return nil, err
	}

	// build map of depedency mod variable values declared in the mod 'Require' section
	depModVarValues, err := inputvars.CollectVariableValuesFromModRequire(w.Mod, runCtx)
	if err != nil {
		return nil, err
	}

	if validate {
		if err := identifyMissingVariables(inputValuesUnparsed, variableMap, depModVarValues); err != nil {
			return nil, err
		}
	}
	parsedValues, diags := inputvars.ParseVariableValues(inputValuesUnparsed, variableMap, depModVarValues, validate)

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
	}
}

func identifyMissingVariables(existing map[string]inputvars.UnparsedVariableValue, vcs map[string]*modconfig.Variable, depModVarValues inputvars.InputValues) error {
	var needed []*modconfig.Variable

	for name, vc := range vcs {
		if !vc.Required() {
			continue // We only prompt for required variables
		}
		_, unparsedValExists := existing[name]
		_, depModVarValueExists := depModVarValues[name]
		if !unparsedValExists && !depModVarValueExists {
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
