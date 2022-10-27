package cmd

import (
	"bufio"
	"context"
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

func runLoginCmd(cmd *cobra.Command, _ []string) {
	ctx := cmd.Context()

	// start login flow - this will open a web page prompting user to login, and will give the user a code to enter
	var id, err = cloud.WebLogin(ctx)
	error_helpers.FailOnError(err)
	// Wait for user to input 4-digit code they obtain through the UI login / approval

	code, err := promptUserForString("Enter login code: ")
	error_helpers.FailOnError(err)

	// handle ctrl+d
	if len(code) == 0 {
		fmt.Println()
		return
	}

	// use this code to get a login token and store it
	token, err := cloud.GetLoginToken(ctx, id, code)
	error_helpers.FailOnError(err)

	// save token
	err = cloud.SaveToken(token)
	error_helpers.FailOnError(err)

	displayLoginMessage(ctx, token)
}

func displayLoginMessage(ctx context.Context, token string) {
	userName, err := cloud.GetUserName(ctx, token)
	error_helpers.FailOnErrorWithMessage(err, "Failed to read user name")

	fmt.Printf("Login successful for user %s\n", userName)

}

func promptUserForString(prompt string) (string, error) {
	fmt.Print(prompt)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	err := scanner.Err()
	if err != nil {
		return "", err
	}
	code := scanner.Text()

	return code, nil
}
