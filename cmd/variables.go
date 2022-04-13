package cmd

import (
	"encoding/json"
	"fmt"
	"log"

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

Examples:

  # List installed variables
  steampipe variable list

  # List variables that have updates available
  steampipe variable list --outdated`,
	}

	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag("outdated", "", false, "Check each variable in the list for updates").
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for variable list")

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

	workspacePath := viper.GetString(constants.ArgWorkspaceChDir)

	vars, err := workspace.LoadVariables(ctx, workspacePath)
	// load the workspace
	utils.FailOnErrorWithMessage(err, "failed to load workspace")

	log.Printf("[WARN] %v", vars)

	if viper.GetString(constants.ArgOutput) == constants.OutputFormatJSON {
		jsonOutput, err := json.MarshalIndent(vars, "", "  ")
		utils.FailOnErrorWithMessage(err, "failed to marshal variables to JSON")
		fmt.Println(jsonOutput)
	} else {

	}
	//headers := []string{"Name", "Version", "Connections"}
	//rows := [][]string{}
	//for _, item := range list {
	//	rows = append(rows, []string{item.Name, item.Version, strings.Join(item.Connections, ",")})
	//}
	//display.ShowWrappedTable(headers, rows, false)

}
