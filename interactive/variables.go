package interactive

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform/terraform"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/tf"
)

func PromptForMissingVariables(shouldRerun bool, missingVariablesError modconfig.MissingVariableError, ctx context.Context) error {
	// is there are missing variables, we will prompt for the values then rerun
	shouldRerun = true
	fmt.Println()
	fmt.Println("Variables defined with no value set.")
	for _, v := range missingVariablesError.MissingVariables {
		r, err := promptForVariable(ctx, v.ShortName, v.Description)
		if err != nil {
			return err
		}
		addInteractiveVariableToViper(v.ShortName, r)

	}
	return nil
}

func promptForVariable(ctx context.Context, name, description string) (string, error) {
	uiInput := &tf.UIInput{}
	rawValue, err := uiInput.Input(ctx, &terraform.InputOpts{
		Id:          fmt.Sprintf("var.%s", name),
		Query:       fmt.Sprintf("var.%s", name),
		Description: description,
	})

	return rawValue, err
}

func addInteractiveVariableToViper(name string, rawValue string) {
	varMap := viper.GetStringMap(constants.ConfigInteractiveVariables)
	varMap[name] = rawValue
	viper.Set(constants.ConfigInteractiveVariables, varMap)
}
