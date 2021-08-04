package workspace

import (
	"fmt"

	"github.com/hashicorp/terraform/tfdiags"

	"github.com/turbot/steampipe/utils"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"
	"github.com/turbot/steampipe/steampipeconfig/tf"
)

func (w *Workspace) getAllVariables() (tf.InputValues, error) {
	variableMap, err := w.loadConfigVariables()
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

func (w *Workspace) loadConfigVariables() (map[string]*modconfig.Variable, error) {
	opts := &parse.ParseModOptions{
		Flags: parse.CreateDefaultMod,
		ListOptions: &filehelpers.ListOptions{
			// listFlag specifies whether to load files recursively
			Flags:   w.listFlag,
			Exclude: w.exclusions,
		},
	}

	return steampipeconfig.LoadVariables(w.Path, opts)
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

func (w *Workspace) getInputVariables(variableMap map[string]*modconfig.Variable) (tf.InputValues, error) {
	meta := &tf.Meta{}

	inputValuesUnparsed, diags := meta.CollectVariableValues(w.Path)
	if diags.HasErrors() {
		return nil, diags.Err()
	}
	parsedValues, diags := tf.ParseVariableValues(inputValuesUnparsed, variableMap)
	return parsedValues, diags.Err()

}

func (w *Workspace) loadFileVariables() (tf.InputValues, error) {
	opts := &filehelpers.ListOptions{
		// listFlag specifies whether to load files recursively
		Flags:   filehelpers.Files,
		Include: []string{fmt.Sprintf("*%s", constants.VariablesExtension)},
	}
	variablesFiles, err := filehelpers.ListFiles(w.Path, opts)

	if err != nil {
		return nil, err
	}
	fileData, diags := parse.LoadFileData(variablesFiles...)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load all variables files", diags)
	}
	fileVariables, err := parse.ParseVariables(fileData)
	if err != nil {
		return nil, err
	}

	var res = tf.InputValues{}
	for k, v := range fileVariables {
		res[k] = &tf.InputValue{
			Value: v,
			// TODO named vs auto
			SourceType: tf.ValueFromNamedFile,
			// TODO get range out of ParseVariables
			//SourceRange: nil,
		}
	}
	return res, nil
}
