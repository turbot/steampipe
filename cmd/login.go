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
	"time"
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
	for retries := 0; ; retries++ {
		code, err := promptUserForString("Enter verification code: ")
		error_helpers.FailOnError(err)

		// handle ctrl+d
		if len(code) == 0 {
			fmt.Println()
			os.Exit(0)
		}

		// use this code to get a login token and store it
		loginToken, err = cloud.GetLoginToken(ctx, id, code)
		if err == nil {
			return loginToken, nil
		}

		fmt.Print("\033[1A\033[K")
		if retries > 0 {
			fmt.Print("\033[1A\033[K")
		}
		if retries == 2 {
			return "", err
		}
		time.Sleep(20 * time.Millisecond)
		error_helpers.ShowWarning(err.Error())
	}

	return
}

//
//func getLoginToken(ctx context.Context, id string) (string, error) {
//
//	for retryCount := 0; retryCount < 3; retryCount++ {
//		// Wait for user to enter confirmation code they obtain through the UI login / approval
//		code, err := promptUserForString("Enter verification code: ", id)
//		error_helpers.FailOnError(err)
//
//		// handle ctrl+d
//		if len(code) == 0 {
//			fmt.Println()
//			return "", nil
//		}
//
//		// seek console output backwards 1 line
//		fmt.Println("\033[1A")
//
//	}
//	return "", fmt.Errorf("invalid code")
//}

func displayLoginMessage(ctx context.Context, token string) {
	userName, err := cloud.GetUserName(ctx, token)
	error_helpers.FailOnErrorWithMessage(err, "Failed to read user name")

	fmt.Printf("Login successful for user %s\n", userName)

}

//	func promptUserForString(promptStr string, id string) (token string, err error) {
//		ctx, cancel := context.WithCancel(context.Background())
//		retries := 0
//		contexthelpers.StartCancelHandler(cancel)
//
//		fmt.Println()
//		callExecutor := func(code string) {
//			// use this code to get a login token and store it
//			token, err = cloud.GetLoginToken(ctx, id, code)
//			if err == nil {
//				cancel()
//			}
//
//			fmt.Print("\033[1A\033[K") //one up, remove line
//
//			retries++
//			if retries == 3 {
//				fmt.Print("\033[1A\033[K") //one up, remove line
//				cancel()
//				return
//			}
//			error_helpers.ShowWarning(err.Error())
//
//		}
//		var completer = func(d prompt.Document) []prompt.Suggest { return nil }
//		p := prompt.New(callExecutor, completer, prompt.OptionPrefix(promptStr))
//
//		p.RunCtx(ctx)
//
//		return
//
// }
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
