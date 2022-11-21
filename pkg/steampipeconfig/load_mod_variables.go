package steampipeconfig

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform/tfdiags"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/inputvars"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/parse"
	"github.com/turbot/steampipe/pkg/type_conversion"
)

func LoadVariableDefinitions(variablePath string, parseCtx *parse.ModParseContext) (*modconfig.ModVariableMap, error) {
	// only load mod and variables blocks
	parseCtx.BlockTypes = []string{modconfig.BlockTypeVariable}
	mod, err := LoadMod(variablePath, parseCtx)
	if err != nil {
		return nil, err
	}

	variableMap := modconfig.NewModVariableMap(mod, parseCtx.LoadedDependencyMods)

	return variableMap, nil
}

func GetVariableValues(ctx context.Context, parseCtx *parse.ModParseContext, variableMap *modconfig.ModVariableMap, validate bool) (*modconfig.ModVariableMap, error) {
	// now resolve all input variables
	inputVariables, err := getInputVariables(variableMap.AllVariables, validate, parseCtx)
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
		variableMap.VariableValues[name], err = type_conversion.CtyToString(inputValue.Value)
		if err != nil {
			return nil, err
		}
	}

	// add workspace mod variables to runContext
	parseCtx.AddInputVariables(variableMap)

	return variableMap, nil
}

func getInputVariables(variableMap map[string]*modconfig.Variable, validate bool, parseCtx *parse.ModParseContext) (inputvars.InputValues, error) {
	variableFileArgs := viper.GetStringSlice(constants.ArgVarFile)
	variableArgs := viper.GetStringSlice(constants.ArgVariable)

	// get mod and mod path from run context
	mod := parseCtx.CurrentMod
	path := mod.ModPath

	inputValuesUnparsed, err := inputvars.CollectVariableValues(path, variableFileArgs, variableArgs)
	if err != nil {
		return nil, err
	}

	// build map of depedency mod variable values declared in the mod 'Require' section
	depModVarValues, err := inputvars.CollectVariableValuesFromModRequire(mod, parseCtx)
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

		error_helpers.ShowError(ctx, fmt.Errorf("%s", constants.Bold(diag.Description().Summary)))
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
