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
	"log"
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
	var id, err = cloud.WebLogin()
	error_helpers.FailOnError(err)
	// Wait for user to input 4-digit code they obtain through the UI login / approval

	code, err := promptUserForString("Enter login code: ")
	error_helpers.FailOnError(err)

	// use this code to get a login token and store it
	token, err := cloud.GetLoginToken(id, code)
	error_helpers.FailOnError(err)

	// save token
	err = cloud.SaveToken(token)
	error_helpers.FailOnError(err)

	// ensure user has at least 1 workspace
	err = ensureWorkspace(ctx, token)
	error_helpers.FailOnError(err)

	fmt.Println("Login successful")
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

func ensureWorkspace(_ context.Context, token string) error {
	workspaces, _, err := cloud.GetUserWorkspaceHandles(token)
	if err != nil {
		return err
	}
	if len(workspaces) > 0 {
		return nil
	}

	workspaceHandle, err := promptUserForString("Enter handle for default workspace: ")
	error_helpers.FailOnError(err)

	log.Println(workspaceHandle)
	return nil
}
