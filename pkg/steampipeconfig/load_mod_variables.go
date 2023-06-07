package steampipeconfig

import (
	"context"
	"fmt"
	"github.com/turbot/steampipe/pkg/type_conversion"
	"sort"

	"github.com/hashicorp/terraform/tfdiags"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/inputvars"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/parse"
)

func LoadVariableDefinitions(variablePath string, parseCtx *parse.ModParseContext) (*modconfig.ModVariableValueMap, error) {
	// only load mod and variables blocks
	parseCtx.BlockTypes = []string{modconfig.BlockTypeVariable}
	mod, errAndWarnings := LoadMod(variablePath, parseCtx)
	if errAndWarnings.GetError() != nil {
		return nil, errAndWarnings.GetError()
	}

	variableMap := modconfig.NewModVariableValueMap(mod)

	return variableMap, nil
}

func GetVariableValues(ctx context.Context, parseCtx *parse.ModParseContext, variableMap *modconfig.ModVariableValueMap, validate bool) (*modconfig.ModVariableValueMap, error) {
	// now resolve all input variables
	inputValues, err := getInputVariables(variableMap.PublicVariables, validate, parseCtx)
	if err != nil {
		return nil, err
	}

	if validate {
		if err := validateVariables(ctx, variableMap.PublicVariables, inputValues); err != nil {
			return nil, err
		}
	}

	// now update the variables map with the input values
	for name, inputValue := range inputValues {
		variable := variableMap.PublicVariables[name]
		variable.SetInputValue(
			inputValue.Value,
			inputValue.SourceTypeString(),
			inputValue.SourceRange)

		// set variable value string in our workspace map
		variableMap.PublicVariableValues[name], err = type_conversion.CtyToString(inputValue.Value)
		if err != nil {
			return nil, err
		}
	}

	return variableMap, nil
}

func getInputVariables(publicVariableMap map[string]*modconfig.Variable, validate bool, parseCtx *parse.ModParseContext) (inputvars.InputValues, error) {
	variableFileArgs := viper.GetStringSlice(constants.ArgVarFile)
	variableArgs := viper.GetStringSlice(constants.ArgVariable)

	// get mod and mod path from run context
	mod := parseCtx.CurrentMod
	path := mod.ModPath

	var inputValuesUnparsed, err = inputvars.CollectVariableValues(path, variableFileArgs, variableArgs, parseCtx.CurrentMod.ShortName)
	if err != nil {
		return nil, err
	}

	if validate {
		if err := identifyMissingVariables(inputValuesUnparsed, publicVariableMap); err != nil {
			return nil, err
		}
	}
	parsedValues, diags := inputvars.ParseVariableValues(inputValuesUnparsed, publicVariableMap, validate)

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

func identifyMissingVariables(existing map[string]inputvars.UnparsedVariableValue, vcs map[string]*modconfig.Variable) error {
	// TODO KAI this does not take into account require args
	var needed []*modconfig.Variable

	for name, vc := range vcs {
		if !vc.Required() {
			continue // We only prompt for required variables
		}
		_, unparsedValExists := existing[name]

		if !unparsedValExists {
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
