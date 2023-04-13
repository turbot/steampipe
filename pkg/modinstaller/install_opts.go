package modinstaller

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
)

type InstallOpts struct {
	WorkspacePath    string
	Command          string
	ModArgs          []string
	DryRun           bool
	Force            bool
	CreateDefaultMod bool
}

func NewInstallOpts(modsToInstall ...string) *InstallOpts {
	cmdName := viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command).Name()
	// only if the command is mod install, create default mod
	createDefault := cmdName == "install"
	opts := &InstallOpts{
		WorkspacePath:    viper.GetString(constants.ArgModLocation),
		DryRun:           viper.GetBool(constants.ArgDryRun),
		Force:            viper.GetBool(constants.ArgForce),
		ModArgs:          modsToInstall,
		Command:          cmdName,
		CreateDefaultMod: createDefault,
	}
	return opts
}
