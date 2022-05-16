package interactive

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform/terraform"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig/inputvars"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

func PromptForMissingVariables(ctx context.Context, missingVariables []*modconfig.Variable, workspacePath string) error {
	fmt.Println()
	fmt.Println("Variables defined with no value set.")
	for _, v := range missingVariables {
		variableName := v.ShortName
		variableDisplayName := fmt.Sprintf("var.%s", v.ShortName)
		// if this variable is NOT part of the workspace mod, add the mod name to the variable name
		if v.Mod.ModPath != workspacePath {
			variableDisplayName = fmt.Sprintf("%s.var.%s", v.ModName, v.ShortName)
			variableName = fmt.Sprintf("%s.%s", v.ModName, v.ShortName)
		}
		r, err := promptForVariable(ctx, variableDisplayName, v.Description)
		if err != nil {
			return err
		}
		addInteractiveVariableToViper(variableName, r)
	}
	return nil
}

func promptForVariable(ctx context.Context, name, description string) (string, error) {
	uiInput := &inputvars.UIInput{}
	rawValue, err := uiInput.Input(ctx, &terraform.InputOpts{
		Id:          name,
		Query:       name,
		Description: description,
	})

	return rawValue, err
}

func addInteractiveVariableToViper(name string, rawValue string) {
	varMap := viper.GetStringMap(constants.ConfigInteractiveVariables)
	varMap[name] = rawValue
	viper.Set(constants.ConfigInteractiveVariables, varMap)
}
