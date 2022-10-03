package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-version"
	_ "github.com/jackc/pgx/v4/stdlib"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmd"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/utils"
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

	// ensure steampipe is not run on WSL1
	checkWsl1(ctx)

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

func checkWsl1(ctx context.Context) {
	// store the 'uname -r' output
	output, err := exec.Command("uname", "-r").Output()
	if err != nil {
		utils.ShowErrorWithMessage(ctx, err, "failed to check uname")
		return
	}
	// convert the ouptut to a string of lowercase characters for ease of use
	op := strings.ToLower(string(output))

	// if WSL2, return
	if strings.Contains(op, "wsl2") {
		return
	}
	// if output contains 'microsoft' or 'wsl', check the kernel version
	if strings.Contains(op, "microsoft") || strings.Contains(op, "wsl") {

		// store the system kernel version
		sys_kernel, _, _ := strings.Cut(string(output), "-")
		sys_kernel_ver, err := version.NewVersion(sys_kernel)
		if err != nil {
			utils.ShowErrorWithMessage(ctx, err, "failed to check system kernel version")
			return
		}
		// if the kernel version >= 4.19, it's WSL Version 2.
		kernel_ver, err := version.NewVersion("4.19")
		if err != nil {
			utils.ShowErrorWithMessage(ctx, err, "checking system kernel version")
			return
		}
		// if the kernel version >= 4.19, it's WSL version 2, else version 1
		if sys_kernel_ver.GreaterThanOrEqual(kernel_ver) {
			return
		} else {
			utils.ShowError(ctx, fmt.Errorf("Steampipe requires WSL2, please upgrade and try again."))
			os.Exit(1)
		}
	}
}
