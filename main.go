package main

import (
	"context"
	"fmt"
	"github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	localcmdconfig "github.com/turbot/steampipe/pkg/cmdconfig"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/go-version"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/steampipe/cmd"
	steampipe_version "github.com/turbot/steampipe/pkg/version"
)

var exitCode int = constants.ExitCodeSuccessful

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
		utils.DisplayProfileData()
		os.Exit(exitCode)
	}()

	// set app specific constants defined in pipe-fittings
	appInit()

	// ensure steampipe is not being run as root
	checkRoot(ctx)

	// ensure steampipe is not run on WSL1
	checkWsl1(ctx)

	// check OSX kernel version
	checkOSXVersion(ctx)

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
		sys_kernel_ver, err := version.NewVersion(sys_kernel)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "failed to check system kernel version")
			return
		}
		// if the kernel version >= 4.19, it's WSL AppVersion 2.
		kernel_ver, err := version.NewVersion("4.19")
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

// set app specific constants defined in pipe-fittings
func appInit() {

	// set the default install dir
	installDir, err := files.Tildefy("~/.steampipe")
	if err != nil {
		panic(err)
	}
	app_specific.DefaultInstallDir = installDir
	app_specific.AppName = "steampipe"
	app_specific.ClientConnectionAppNamePrefix = "steampipe_client"
	app_specific.ServiceConnectionAppNamePrefix = "steampipe_service"
	app_specific.ClientSystemConnectionAppNamePrefix = "steampipe_client_system"
	app_specific.AppVersion = steampipe_version.SteampipeVersion
	app_specific.DefaultWorkspaceDatabase = "local"
	app_specific.ModDataExtension = ".sp"
	app_specific.VariablesExtension = ".spvars"
	app_specific.AutoVariablesExtension = ".auto.spvars"
	app_specific.DefaultVarsFileName = "steampipe.spvars"
	app_specific.ModFileName = "mod.sp"
	app_specific.WorkspaceIgnoreFile = ".steampipeignore"
	app_specific.WorkspaceDataDir = ".steampipe"

	app_specific.EnvAppPrefix = "STEAMPIPE_"
	// EnvInputVarPrefix is the prefix for environment variables that represent values for input variables.
	app_specific.EnvInputVarPrefix = "SP_VAR_"
	// set the command pre and post hooks
	cmdconfig.CustomPreRunHook = localcmdconfig.PreRunHook
	cmdconfig.CustomPostRunHook = localcmdconfig.PostRunHook
}
