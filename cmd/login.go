package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/pipes"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/v2/pkg/cmdconfig"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
)

func loginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "login",
		TraverseChildren: true,
		Args:             cobra.NoArgs,
		Run:              runLoginCmd,
		Short:            "Login to Turbot Pipes",
		Long:             `Login to Turbot Pipes.`,
	}

	cmdconfig.OnCmd(cmd).
		AddCloudFlags().
		AddBoolFlag(pconstants.ArgHelp, false, "Help for dashboard", cmdconfig.FlagOptions.WithShortHand("h"))

	return cmd
}

func runLoginCmd(cmd *cobra.Command, _ []string) {
	ctx := cmd.Context()

	log.Printf("[TRACE] login, pipes host %s", viper.Get(pconstants.ArgPipesHost))
	log.Printf("[TRACE] opening login web page")
	// start login flow - this will open a web page prompting user to login, and will give the user a code to enter
	var id, err = pipes.WebLogin(ctx)
	if err != nil {
		error_helpers.ShowError(ctx, err)
		exitCode = constants.ExitCodeLoginCloudConnectionFailed
		return
	}

	token, err := getToken(ctx, id)
	if err != nil {
		error_helpers.ShowError(ctx, err)
		exitCode = constants.ExitCodeLoginCloudConnectionFailed
		return
	}

	// save token
	err = pipes.SaveToken(token)
	if err != nil {
		error_helpers.ShowError(ctx, err)
		exitCode = constants.ExitCodeLoginCloudConnectionFailed
		return
	}

	displayLoginMessage(ctx, token)
}

func getToken(ctx context.Context, id string) (loginToken string, err error) {
	log.Printf("[TRACE] prompt for verification code")

	fmt.Println()
	retries := 0
	for {
		var code string
		code, err = promptUserForString("Enter verification code: ")
		error_helpers.FailOnError(err)
		if code != "" {
			log.Printf("[TRACE] get login token")
			// use this code to get a login token and store it
			loginToken, err = pipes.GetLoginToken(ctx, id, code)
			if err == nil {
				return loginToken, nil
			}
		}
		if err != nil {
			// a code was entered but it failed - inc retry count
			log.Printf("[TRACE] GetLoginToken failed with %s", err.Error())
		}
		retries++

		// if we have used our retries, break out before displaying wanring - we will display an error
		if retries == 3 {
			return "", sperr.New("Too many attempts.")
		}

		if err != nil {
			error_helpers.ShowWarning(err.Error())
		}
		log.Printf("[TRACE] Retrying")
	}
}

func displayLoginMessage(ctx context.Context, token string) {
	userName, err := pipes.GetUserName(ctx, token)
	error_helpers.FailOnError(sperr.WrapWithMessage(err, "failed to read user name"))

	fmt.Println()
	fmt.Printf("Logged in as: %s\n", pconstants.Bold(userName))
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
		return "", sperr.Wrap(err)
	}
	code := scanner.Text()

	return code, nil
}
