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
	for _, diag := range diags {
		utils.ShowError(fmt.Errorf("%s", constants.Bold(diag.Description().Summary)))
		fmt.Println(diag.Description().Detail)
		fmt.Println()

		// TODO range if there is one
	}
}
