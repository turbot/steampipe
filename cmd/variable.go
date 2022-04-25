package cmd

import (
	"fmt"

	"github.com/turbot/steampipe/display"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/workspace"

	"github.com/spf13/cobra"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

//  Variable management commands
func variableCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "variable [command]",
		Args:  cobra.NoArgs,
		Short: "Steampipe variable management",
		Long:  `Steampipe variable management.`,
	}

	cmd.AddCommand(variableListCmd())
	cmd.Flags().BoolP(constants.ArgHelp, "h", false, "Help for variable")

	return cmd
}

// List variables
func variableListCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Run:   runVariableListCmd,
		Short: "List currently installed variables",
		Long: `List currently installed variables.

List all Steampipe variables installed for this user.

Example:

  # List installed variables
  steampipe variable list

`,
	}

	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag("outdated", "", false, "Check each variable in the list for updates").
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for variable list").
		AddStringFlag(constants.ArgOutput, "", constants.OutputFormatTable, "Select a console output format: table or json")

	return cmd
}

func runVariableListCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	defer func() {
		if r := recover(); r != nil {
			utils.ShowError(ctx, helpers.ToError(r))
			exitCode = constants.ExitCodeUnknownErrorPanic
		}
	}()

	// validate output arg
	output := viper.GetString(constants.ArgOutput)
	if !helpers.StringSliceContains([]string{constants.OutputFormatTable, constants.OutputFormatJSON}, output) {
		utils.ShowError(ctx, fmt.Errorf("output flag must be either 'json' or 'table'"))
		return
	}

	workspacePath := viper.GetString(constants.ArgWorkspaceChDir)

	vars, err := workspace.LoadVariables(ctx, workspacePath)
	// load the workspace
	utils.FailOnErrorWithMessage(err, "failed to load workspace")

	if viper.GetString(constants.ArgOutput) == constants.OutputFormatJSON {
		display.ShowVarsListJson(vars)
	} else {

		display.ShowVarsListTable(vars)
	}
}
