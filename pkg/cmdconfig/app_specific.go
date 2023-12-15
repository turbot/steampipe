package cmdconfig

import (
	"github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/error_helpers"
	steampipeversion "github.com/turbot/steampipe/pkg/version"
	"os"
	"path/filepath"
	"strings"
)

// SetAppSpecificConstants sets app specific constants defined in pipe-fittings
func SetAppSpecificConstants() {
	app_specific.AppName = "steampipe"
	app_specific.AppVersion = steampipeversion.SteampipeVersion
	app_specific.AutoVariablesExtension = ".auto.spvars"
	app_specific.ClientConnectionAppNamePrefix = "steampipe_client"
	app_specific.ClientSystemConnectionAppNamePrefix = "steampipe_client_system"
	app_specific.DefaultVarsFileName = "steampipe.spvars"
	app_specific.DefaultWorkspaceDatabase = "local"
	app_specific.ModDataExtension = ".sp"
	app_specific.ModFileName = "mod.sp"
	app_specific.ServiceConnectionAppNamePrefix = "steampipe_service"
	app_specific.ConfigExtension = ".spc"
	app_specific.VariablesExtension = ".spvars"
	app_specific.WorkspaceIgnoreFile = ".steampipeignore"
	app_specific.WorkspaceDataDir = ".steampipe"
	app_specific.EnvAppPrefix = "STEAMPIPE_"
	// EnvInputVarPrefix is the prefix for environment variables that represent values for input variables.
	app_specific.EnvInputVarPrefix = "SP_VAR_"
	// set the command pre and post hooks
	cmdconfig.CustomPreRunHook = preRunHook
	cmdconfig.CustomPostRunHook = postRunHook

	// set the default install dir
	defaultInstallDir, err := files.Tildefy("~/.steampipe")
	error_helpers.FailOnError(err)
	app_specific.DefaultInstallDir = defaultInstallDir

	// set the default config path
	globalConfigPath := filepath.Join(defaultInstallDir, "config")
	// check whether install-dir env has been set - if so, respect it
	if envInstallDir, ok := os.LookupEnv(app_specific.EnvInstallDir); ok {
		globalConfigPath = filepath.Join(envInstallDir, "config")
	}
	app_specific.DefaultConfigPath = strings.Join([]string{".", globalConfigPath}, ":")
}
