package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/turbot/steampipe/pkg/cloud"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"os"
)

func loginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "login",
		TraverseChildren: true,
		Args:             cobra.NoArgs,
		Run:              runLoginCmd,
		Short:            "Login to Steampipe Cloud",
		Long:             `Login to Steampipe Cloud.`,
	}

	cmdconfig.OnCmd(cmd).AddBoolFlag(constants.ArgHelp, "h", false, "Help for dashboard")

	return cmd
}

func runLoginCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()

	// start login flow - this will open a web page prompting user to login, and will give the user a code to enter
	var err = cloud.WebLogin(ctx)
	error_helpers.FailOnError(err)
	// Wait for user to input 4-digit code they obtain through the UI login / approval

	code, err := promptUserForCode()
	error_helpers.FailOnError(err)

	// use this code to get a login token and store it
	token, err := cloud.GetLoginToken(ctx, code)
	error_helpers.FailOnError(err)

	// save token
	err = cloud.SaveToken(ctx, token)
	error_helpers.FailOnError(err)

	// ensure user has at least 1 workspace
	err = cloud.EnsureWorkspace(ctx, token)
	error_helpers.FailOnError(err)

	fmt.Println("Login successful")
}

func promptUserForCode() (string, error) {
	fmt.Printf("Enter login code code: ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	err := scanner.Err()
	if err != nil {
		return "", err
	}
	code := scanner.Text()

	return code, nil
}
