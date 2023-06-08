package steampipeconfig

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform/tfdiags"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/inputvars"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/parse"
	"github.com/turbot/steampipe/pkg/steampipeconfig/versionmap"
)

func LoadVariableDefinitions(variablePath string, parseCtx *parse.ModParseContext) (*modconfig.ModVariableMap, error) {
	// only load mod and variables blocks
	parseCtx.BlockTypes = []string{modconfig.BlockTypeVariable}
	mod, errAndWarnings := LoadMod(variablePath, parseCtx)
	if errAndWarnings.GetError() != nil {
		return nil, errAndWarnings.GetError()
	}

	variableMap := modconfig.NewModVariableMap(mod)

	return variableMap, nil
}

func GetVariableValues(ctx context.Context, parseCtx *parse.ModParseContext, variableMap *modconfig.ModVariableMap, validate bool) (*modconfig.ModVariableMap, error) {
	// now resolve all input variables
	inputValues, err := getInputVariables(ctx, parseCtx, variableMap, validate)
	if err != nil {
		return nil, err
	}

	if validate {
		if err := validateInputVariables(ctx, parseCtx, variableMap, inputValues); err != nil {
			return nil, err
		}
	}

	// now update the variables map with the input values
	inputValues.SetVariableValues(variableMap)

	return variableMap, nil
}

func getInputVariables(ctx context.Context, parseCtx *parse.ModParseContext, variableMap *modconfig.ModVariableMap, validate bool) (inputvars.InputValues, error) {
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
		if err := identifyAllMissingVariables(ctx, parseCtx, variableMap, inputValuesUnparsed); err != nil {
			return nil, err
		}
	}
	// only parse values for public variables
	parsedValues, diags := inputvars.ParseVariableValues(inputValuesUnparsed, variableMap.PublicVariables, validate)

	return parsedValues, diags.Err()
}

func validateInputVariables(ctx context.Context, parseCtx *parse.ModParseContext, variableMap *modconfig.ModVariableMap, variables inputvars.InputValues) error {
	diags := inputvars.CheckInputVariables(variableMap.PublicVariables, variables)
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

func identifyAllMissingVariables(ctx context.Context, parseCtx *parse.ModParseContext, variableMap *modconfig.ModVariableMap, variableValues map[string]inputvars.UnparsedVariableValue) error {
	// convert variableValues into a lookup
	var variableValueLookup = make(map[string]struct{}, len(variableValues))
	missingVarsMap, err := identifyMissingVariablesForDependencies(parseCtx.WorkspaceLock, variableMap, variableValueLookup, nil)

	if err != nil {
		return err
	}
	if len(missingVarsMap) == 0 {
		// all good
		return nil
	}

	// build a MissingVariableError
	missingVarErr := &modconfig.MissingVariableError{}

	// build a lookup with the dependency path of the root mod and all top level dependencies
	rootName := variableMap.Mod.ShortName
	topLevelModLookup := map[string]struct{}{rootName: {}}
	for dep := range parseCtx.WorkspaceLock.InstallCache {
		varDepPath := buildDependencyPathKey(rootName, dep)
		topLevelModLookup[varDepPath] = struct{}{}
	}
	for depPath, missingVars := range missingVarsMap {
		_, isTopLevel := topLevelModLookup[depPath]

		// convert missing vars into MissingVar structs
		missingWithPath := make([]*modconfig.MissingVariable, len(missingVars))
		for i, v := range missingVars {
			varDepPath := buildDependencyPathKey(depPath, v.ShortName)
			missingWithPath[i] = &modconfig.MissingVariable{Variable: v, Path: varDepPath}
		}
		missingVarErr.Add(missingWithPath, isTopLevel)
	}

	return missingVarErr
}

func identifyMissingVariablesForDependencies(workspaceLock *versionmap.WorkspaceLock, variableMap *modconfig.ModVariableMap, parentVariableValuesLookup map[string]struct{}, dependencyPath []string) (map[string][]*modconfig.Variable, error) {
	// return a map of missing variables, keyed by dependency path
	res := make(map[string][]*modconfig.Variable)

	// update the path to this dependency
	dependencyPath = append(dependencyPath, variableMap.Mod.GetInstallCacheKey())

	// clone variableValuesLookup so we can mutate it with depdency specific args overrides
	var variableValueLookup = make(map[string]struct{}, len(parentVariableValuesLookup))
	for k := range parentVariableValuesLookup {
		variableValueLookup[k] = struct{}{}
	}

	// first get any args specified in the mod requires
	// note the actual value of these may be unknown as we have not yet resolved
	depModArgs, err := inputvars.CollectVariableValuesFromModRequire(variableMap.Mod, workspaceLock)
	for varName := range depModArgs {
		variableValueLookup[varName] = struct{}{}
	}
	if err != nil {
		return nil, err
	}

	//  handle root variables
	missingVariables := identifyMissingVariables(variableMap.RootVariables, variableValueLookup)
	if len(missingVariables) > 0 {
		res[buildDependencyPathKey(dependencyPath...)] = missingVariables
	}

	// now iterate through all the dependency variable maps
	for _, dependencyVariableMap := range variableMap.DependencyVariables {
		childMissingMap, err := identifyMissingVariablesForDependencies(workspaceLock, dependencyVariableMap, variableValueLookup, dependencyPath)
		if err != nil {
			return nil, err
		}
		// add results into map
		for k, v := range childMissingMap {
			res[k] = v
		}
	}
	return res, nil
}

func buildDependencyPathKey(dependencyPath ...string) string {
	return strings.Join(dependencyPath, " -> ")
}

func identifyMissingVariables(variableMap map[string]*modconfig.Variable, variableValuesLookup map[string]struct{}) []*modconfig.Variable {

	var needed []*modconfig.Variable

	for name, v := range variableMap {
		if !v.Required() {
			continue // We only prompt for required variables
		}
		_, unparsedValExists := variableValuesLookup[name]

		if !unparsedValExists {
			needed = append(needed, v)
		}
	}
	sort.SliceStable(needed, func(i, j int) bool {
		return needed[i].Name() < needed[j].Name()
	})
	return needed

}
