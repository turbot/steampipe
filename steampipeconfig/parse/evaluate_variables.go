package parse

import (
	"fmt"
	"sort"
	"strings"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/hashicorp/terraform/tfdiags"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig/input_vars"
	"github.com/turbot/steampipe/utils"
	"github.com/zclconf/go-cty/cty"
)

func EvaluateVariables(m *modconfig.Mod) error {
	// TACTICAL - as the tf derived code builds a map keyed by the short variable name, do the same
	variableMap := make(map[string]*modconfig.Variable)
	for k, v := range m.Variables {
		name := strings.Split(k, ".")[1]
		variableMap[name] = v
	}

	// if there is a steampipe variables file, load it
	inputVariables, err := getInputVariables(variableMap, m.ModPath)
	if err != nil {
		return err
	}

	if err := validateVariables(variableMap, inputVariables); err != nil {
		return err
	}

	// now update the variables map with the input values
	for name, inputValue := range inputVariables {
		variable := variableMap[name]
		variable.SetInputValue(
			inputValue.Value,
			inputValue.SourceTypeString(),
			inputValue.SourceRange)
	}

	// as the variables are stored by pointer, the mod variables map has been updated too
	return nil
}

func getInputVariables(variableMap map[string]*modconfig.Variable, modPath string) (input_vars.InputValues, error) {
	variableFileArgs := viper.GetStringSlice(constants.ArgVarFile)
	variableArgs := viper.GetStringSlice(constants.ArgVariable)

	inputValuesUnparsed, diags := input_vars.CollectVariableValues(modPath, variableFileArgs, variableArgs)
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

// VariableValueMap converts Variables map into cty value map
func VariableValueMap(variableMap map[string]*modconfig.Variable) map[string]cty.Value {
	ret := make(map[string]cty.Value, len(variableMap))
	for k, v := range variableMap {
		ret[k] = v.Value
	}
	return ret
}
