package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	go_version "github.com/hashicorp/go-version"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/cmd"
	"github.com/turbot/steampipe/v2/pkg/cmdconfig"
	"github.com/turbot/steampipe/v2/pkg/constants"
	localconstants "github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
)

var exitCode int = constants.ExitCodeSuccessful

var (
	// these variables will be set by GoReleaser
	version = localconstants.DefaultVersion
	commit  = localconstants.DefaultCommit
	date    = localconstants.DefaultDate
	builtBy = localconstants.DefaultBuiltBy
)

func main() {
	ctx := context.Background()
	utils.LogTime("main start")
	defer func() {
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			if exitCode == 0 {
				exitCode = constants.ExitCodeUnknownErrorPanic
			}
		}
		utils.LogTime("main end")
		utils.DisplayProfileData(os.Stdout)
		os.Exit(exitCode)
	}()

	// add the auto-populated version properties into viper
	setVersionProperties()

	// ensure steampipe is not being run as root
	checkRoot(ctx)

	// ensure steampipe is not run on WSL1
	checkWsl1(ctx)

	// check OSX kernel version
	checkOSXVersion(ctx)

	cmdconfig.SetAppSpecificConstants()

	cmd.InitCmd()

	// execute the command
	exitCode = cmd.Execute()
}

// this is to replicate the user security mechanism of out underlying
// postgresql engine.
func checkRoot(ctx context.Context) {
	if os.Geteuid() == 0 {
		exitCode = constants.ExitCodeInvalidExecutionEnvironment
		error_helpers.ShowError(ctx, fmt.Errorf(`Steampipe cannot be run as the "root" user.
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
		exitCode = constants.ExitCodeInvalidExecutionEnvironment
		error_helpers.ShowError(ctx, fmt.Errorf("real and effective user IDs must match."))
		os.Exit(exitCode)
	}
}

func checkWsl1(ctx context.Context) {
	// store the 'uname -r' output
	output, err := exec.Command("uname", "-r").Output()
	if err != nil {
		error_helpers.ShowErrorWithMessage(ctx, err, "failed to check uname")
		return
	}
	// convert the output to a string of lowercase characters for ease of use
	op := strings.ToLower(string(output))

	// if WSL2, return
	if strings.Contains(op, "wsl2") {
		return
	}
	// if output contains 'microsoft' or 'wsl', check the kernel version
	if strings.Contains(op, "microsoft") || strings.Contains(op, "wsl") {

		// store the system kernel version
		sys_kernel, _, _ := strings.Cut(string(output), "-")
		sys_kernel_ver, err := go_version.NewVersion(sys_kernel)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "failed to check system kernel version")
			return
		}
		// if the kernel version >= 4.19, it's WSL Version 2.
		kernel_ver, err := go_version.NewVersion("4.19")
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "checking system kernel version")
			return
		}
		// if the kernel version >= 4.19, it's WSL version 2, else version 1
		if sys_kernel_ver.GreaterThanOrEqual(kernel_ver) {
			return
		} else {
			error_helpers.ShowError(ctx, fmt.Errorf("Steampipe requires WSL2, please upgrade and try again."))
			os.Exit(constants.ExitCodeInvalidExecutionEnvironment)
		}
	}
}

func checkOSXVersion(ctx context.Context) {
	// get the OS and return if not darwin
	if runtime.GOOS != "darwin" {
		return
	}

	// get kernel version
	output, err := exec.Command("uname", "-r").Output()
	if err != nil {
		error_helpers.ShowErrorWithMessage(ctx, err, "failed to get kernel version")
		return
	}

	// get the semver version from string
	version, err := semver.NewVersion(strings.TrimRight(string(output), "\n"))
	if err != nil {
		error_helpers.ShowErrorWithMessage(ctx, err, "failed to get version")
		return
	}
	catalina, err := semver.NewVersion("19.0.0")
	if err != nil {
		error_helpers.ShowErrorWithMessage(ctx, err, "failed to get version")
		return
	}

	// check if Darwin version is not less than Catalina(Darwin version 19.0.0)
	if version.Compare(catalina) == -1 {
		error_helpers.ShowError(ctx, fmt.Errorf("Steampipe requires MacOS 10.15 (Catalina) and above, please upgrade and try again."))
		os.Exit(constants.ExitCodeInvalidExecutionEnvironment)
	}
}

func setVersionProperties() {
	viper.SetDefault(constants.ConfigKeyVersion, version)
	viper.SetDefault(constants.ConfigKeyCommit, commit)
	viper.SetDefault(constants.ConfigKeyDate, date)
	viper.SetDefault(constants.ConfigKeyBuiltBy, builtBy)
}
