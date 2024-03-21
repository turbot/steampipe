package steampipeconfig

import (
	"context"
	"golang.org/x/exp/maps"
	"log"
	"sort"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/inputvars"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/parse"
	"github.com/turbot/steampipe/pkg/steampipeconfig/versionmap"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/terraform-components/tfdiags"
)

func LoadVariableDefinitions(ctx context.Context, variablePath string, parseCtx *parse.ModParseContext) (*modconfig.ModVariableMap, error) {
	// only load mod and variables blocks
	parseCtx.BlockTypes = []string{modconfig.BlockTypeVariable}
	mod, errAndWarnings := LoadMod(ctx, variablePath, parseCtx)
	if errAndWarnings.GetError() != nil {
		return nil, errAndWarnings.GetError()
	}

	variableMap := modconfig.NewModVariableMap(mod)

	return variableMap, nil
}

func GetVariableValues(parseCtx *parse.ModParseContext, variableMap *modconfig.ModVariableMap, validate bool) (*modconfig.ModVariableMap, error_helpers.ErrorAndWarnings) {
	log.Printf("[INFO] GetVariableValues")
	// now resolve all input variables
	inputValues, errorsAndWarnings := getInputVariables(parseCtx, variableMap, validate)
	if errorsAndWarnings.Error == nil {
		// now update the variables map with the input values
		inputValues.SetVariableValues(variableMap)
	}

	return variableMap, errorsAndWarnings
}

func getInputVariables(parseCtx *parse.ModParseContext, variableMap *modconfig.ModVariableMap, validate bool) (inputvars.InputValues, error_helpers.ErrorAndWarnings) {
	variableFileArgs := viper.GetStringSlice(constants.ArgVarFile)
	variableArgs := viper.GetStringSlice(constants.ArgVariable)

	// get mod and mod path from run context
	mod := parseCtx.CurrentMod
	path := mod.ModPath

	log.Printf("[INFO] getInputVariables, variableFileArgs: %s, variableArgs: %s", variableFileArgs, variableArgs)

	var inputValuesUnparsed, err = inputvars.CollectVariableValues(path, variableFileArgs, variableArgs, parseCtx.CurrentMod)
	if err != nil {
		log.Printf("[WARN] CollectVariableValues failed: %s", err.Error())

		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	log.Printf("[INFO] collected unparsed input values for vars: %s", strings.Join(maps.Keys(inputValuesUnparsed), ","))

	if validate {
		if err := identifyAllMissingVariables(parseCtx, variableMap, inputValuesUnparsed); err != nil {
			log.Printf("[INFO] identifyAllMissingVariables returned a validation error: %s", err.Error())

			return nil, error_helpers.NewErrorsAndWarning(err)
		}
	}

	// only parse values for public variables
	parsedValues, diags := inputvars.ParseVariableValues(inputValuesUnparsed, variableMap, validate)
	if diags.HasErrors() {
		log.Printf("[INFO] ParseVariableValues returned error: %s", diags.Err())
	} else {
		log.Printf("[INFO] parsed values for public variables: %s", strings.Join(maps.Keys(parsedValues), ","))
	}

	if validate {
		moreDiags := inputvars.CheckInputVariables(variableMap.PublicVariables, parsedValues)
		diags = append(diags, moreDiags...)
	}

	return parsedValues, newVariableValidationResult(diags)
}

func newVariableValidationResult(diags tfdiags.Diagnostics) error_helpers.ErrorAndWarnings {
	warnings := plugin.DiagsToWarnings(diags.ToHCL())
	var err error
	if diags.HasErrors() {
		err = newVariableValidationFailedError(diags)
	}
	return error_helpers.NewErrorsAndWarning(err, warnings...)
}

func identifyAllMissingVariables(parseCtx *parse.ModParseContext, variableMap *modconfig.ModVariableMap, variableValues map[string]inputvars.UnparsedVariableValue) error {
	// convert variableValues into a lookup
	var variableValueLookup = utils.SliceToLookup(maps.Keys(variableValues))
	missingVarsMap, err := identifyMissingVariablesForDependencies(parseCtx.WorkspaceLock, variableMap, variableValueLookup, nil)

	if err != nil {
		return err
	}
	if len(missingVarsMap) == 0 {
		// all good
		return nil
	}

	// build a MissingVariableError
	missingVarErr := NewMissingVarsError(parseCtx.CurrentMod)

	// build a lookup with the dependency path of the root mod and all top level dependencies
	rootName := variableMap.Mod.ShortName
	topLevelModLookup := map[DependencyPathKey]struct{}{DependencyPathKey(rootName): {}}
	for dep := range parseCtx.WorkspaceLock.InstallCache {
		depPathKey := newDependencyPathKey(rootName, dep)
		topLevelModLookup[depPathKey] = struct{}{}
	}
	for depPath, missingVars := range missingVarsMap {
		if _, isTopLevel := topLevelModLookup[depPath]; isTopLevel {
			missingVarErr.MissingVariables = append(missingVarErr.MissingVariables, missingVars...)
		} else {
			missingVarErr.MissingTransitiveVariables[depPath] = missingVars
		}
	}

	return missingVarErr
}

func identifyMissingVariablesForDependencies(workspaceLock *versionmap.WorkspaceLock, variableMap *modconfig.ModVariableMap, parentVariableValuesLookup map[string]struct{}, dependencyPath []string) (map[DependencyPathKey][]*modconfig.Variable, error) {
	// return a map of missing variables, keyed by dependency path
	res := make(map[DependencyPathKey][]*modconfig.Variable)

	// update the path to this dependency
	dependencyPath = append(dependencyPath, variableMap.Mod.GetInstallCacheKey())

	// clone variableValuesLookup so we can mutate it with depdency specific args overrides
	var variableValueLookup = make(map[string]struct{}, len(parentVariableValuesLookup))
	for k := range parentVariableValuesLookup {
		// convert the variable name to the short name if it is fully qualified and belongs to the current mod
		k = getVariableValueMapKey(k, variableMap)

		variableValueLookup[k] = struct{}{}
	}

	// first get any args specified in the mod requires
	// note the actual value of these may be unknown as we have not yet resolved
	depModArgs, err := inputvars.CollectVariableValuesFromModRequire(variableMap.Mod, workspaceLock)
	for varName := range depModArgs {
		// convert the variable name to the short name if it is fully qualified and belongs to the current mod
		varName = getVariableValueMapKey(varName, variableMap)

		variableValueLookup[varName] = struct{}{}
	}
	if err != nil {
		return nil, err
	}

	//  handle root variables
	missingVariables := identifyMissingVariables(variableMap.RootVariables, variableValueLookup, variableMap.Mod.ShortName)
	if len(missingVariables) > 0 {
		res[newDependencyPathKey(dependencyPath...)] = missingVariables
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

// getVariableValueMapKey checks whether the variable is fully qualified and belongs to the current mod,
// if so use the short name
func getVariableValueMapKey(k string, variableMap *modconfig.ModVariableMap) string {
	// attempt to parse the variable name.
	// Note: if the variable is not fully qualified (e.g. "var_name"),  ParseResourceName will return an error
	// in which case we add it to our map unchanged
	parsedName, err := modconfig.ParseResourceName(k)
	// if this IS a dependency variable, the parse will success
	// if the mod name is the same as the current mod (variableMap.Mod)
	// then add a map entry with the variable short name
	// this will allow us to match the variable value to a variable defined in this mod
	if err == nil && parsedName.Mod == variableMap.Mod.ShortName {
		k = parsedName.Name
	}
	return k
}

func identifyMissingVariables(variableMap map[string]*modconfig.Variable, variableValuesLookup map[string]struct{}, modName string) []*modconfig.Variable {

	var needed []*modconfig.Variable

	for shortName, v := range variableMap {
		if !v.Required() {
			continue // We only prompt for required variables
		}
		_, unparsedValExists := variableValuesLookup[shortName]

		if !unparsedValExists {
			needed = append(needed, v)
		}
	}
	sort.SliceStable(needed, func(i, j int) bool {
		return needed[i].Name() < needed[j].Name()
	})
	return needed

}
