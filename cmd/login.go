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

	token, err := getToken(ctx, id)
	error_helpers.FailOnError(err)

	// save token
	err = cloud.SaveToken(token)
	error_helpers.FailOnError(err)

	displayLoginMessage(ctx, token)
}

func getToken(ctx context.Context, id string) (loginToken string, err error) {
	fmt.Println()
	retries := 0
	for {
		var code string
		code, err = promptUserForString("Enter verification code: ")
		error_helpers.FailOnError(err)
		if code != "" {
			// use this code to get a login token and store it
			loginToken, err = cloud.GetLoginToken(ctx, id, code)
			if err == nil {
				return loginToken, nil
			}
			// a code was entered but it failed - inc retry count
			retries++
		}

		// if we have used our retries, break out before displaying wanring - we will display an error
		if retries == 3 {
			return "", fmt.Errorf("Too many attempts.")
		}

		if err != nil {
			error_helpers.ShowWarning(err.Error())
		}
	}

	return
}

func displayLoginMessage(ctx context.Context, token string) {
	userName, err := cloud.GetUserName(ctx, token)
	error_helpers.FailOnErrorWithMessage(err, "Failed to read user name")

	fmt.Println()
	fmt.Printf("Logged in as: %s\n", constants.Bold(userName))
	fmt.Println()
}

func promptUserForString(prompt string) (string, error) {
	fmt.Print(prompt)

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		// handle ctrl+d
		fmt.Println()
		os.Exit(0)
	}

	err := scanner.Err()
	if err != nil {
		return "", err
	}
	code := scanner.Text()

	return code, nil
}
