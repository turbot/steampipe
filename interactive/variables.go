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

func PromptForMissingVariables(ctx context.Context, missingVariables []*modconfig.Variable) error {
	fmt.Println()
	fmt.Println("Variables defined with no value set.")
	for _, v := range missingVariables {
		r, err := promptForVariable(ctx, v.ShortName, v.Description)
		if err != nil {
			return err
		}
		addInteractiveVariableToViper(v.ShortName, r)
	}
	return nil
}

func promptForVariable(ctx context.Context, name, description string) (string, error) {
	uiInput := &inputvars.UIInput{}
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
