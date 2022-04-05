package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/turbot/steampipe/cmd"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

var exitCode int

func main() {
	ctx := context.Background()
	utils.LogTime("main start")
	exitCode := constants.ExitCodeSuccessful
	defer func() {
		if r := recover(); r != nil {
			utils.ShowError(ctx, helpers.ToError(r))
		}
		utils.LogTime("main end")
		utils.DisplayProfileData()
		os.Exit(exitCode)
	}()

	// ensure steampipe is not being run as root
	checkRoot(ctx)

	// ensure UTF-8 langpacks are installed
	checkLangpacks(ctx)

	// increase the soft ULIMIT to match the hard limit
	err := setULimit()
	utils.FailOnErrorWithMessage(err, "failed to increase the file limit")

	cmd.InitCmd()

	// execute the command
	exitCode = cmd.Execute()
}

// set the current to the max to avoid any file handle shortages
func setULimit() error {
	ulimit, err := filehelpers.GetULimit()
	if err != nil {
		return err
	}

	// set the current ulimit to 8192 (or the max, if less)
	// this is to ensure we do not run out of file handler when watching files
	var newULimit uint64 = 8192
	if newULimit > ulimit.Max {
		newULimit = ulimit.Max
	}
	err = filehelpers.SetULimit(newULimit)
	return err
}

// this is to replicate the user security mechanism of out underlying
// postgresql engine.
func checkRoot(ctx context.Context) {
	if os.Geteuid() == 0 {
		exitCode = constants.ExitCodeUnknownErrorPanic
		utils.ShowError(ctx, fmt.Errorf(`Steampipe cannot be run as the "root" user.
To reduce security risk, use an unprivileged user account instead.`))
		os.Exit(exitCode)
	}

	/*
	 * Also make sure that real and effective uids are the same. Executing as
	 * a setuid program from a root shell is a security hole, since on many
	 * platforms a nefarious subroutine could setuid back to root if real uid
	 * is root.  (Since nobody actually uses postgres as a setuid program,
	 * trying to actively fix this situation seems more trouble than it's
	 * worth; we'll just expend the effort to check for it.)
	 */

	if os.Geteuid() != os.Getuid() {
		exitCode = constants.ExitCodeUnknownErrorPanic
		utils.ShowError(ctx, fmt.Errorf("real and effective user IDs must match."))
		os.Exit(exitCode)
	}
}

func checkLangpacks(ctx context.Context) {
	if !strings.Contains(os.Getenv("LC_CTYPE"), "UTF-8") {
		utils.ShowError(ctx, fmt.Errorf(`UTF-8 langpacks need to be installed to run steampipe`))
		os.Exit(1)
	}
}
