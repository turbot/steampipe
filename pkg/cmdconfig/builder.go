package cmdconfig

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/utils"
)

type CmdBuilder struct {
	cmd      *cobra.Command
	bindings map[string]*pflag.Flag
}

// OnCmd starts a config builder wrapping over the provided *cobra.Command
func OnCmd(cmd *cobra.Command) *CmdBuilder {
	cfg := new(CmdBuilder)
	cfg.cmd = cmd
	cfg.bindings = map[string]*pflag.Flag{}

	// we will wrap over these two function - need references to call them
	originalPreRun := cfg.cmd.PreRun
	cfg.cmd.PreRun = func(cmd *cobra.Command, args []string) {
		utils.LogTime(fmt.Sprintf("cmd.%s.PreRun start", cmd.CommandPath()))
		defer utils.LogTime(fmt.Sprintf("cmd.%s.PreRun end", cmd.CommandPath()))
		// bind flags
		for flagName, flag := range cfg.bindings {
			if flag == nil {
				// we can panic here since this is bootstrap code and not execution path specific
				panic(fmt.Sprintf("flag for %s cannot be nil", flagName))
			}
			//nolint:golint,errcheck // nil check above
			viper.GetViper().BindPFlag(flagName, flag)
		}

		// now that we have done all the flag bindings, run the global pre run
		// this will load up and populate the global config, init the logger and
		// also run the daily task runner
		preRunHook(cmd, args)

		// run the original PreRun
		if originalPreRun != nil {
			originalPreRun(cmd, args)
		}
	}

	originalPostRun := cfg.cmd.PostRun
	cfg.cmd.PostRun = func(cmd *cobra.Command, args []string) {
		utils.LogTime(fmt.Sprintf("cmd.%s.PostRun start", cmd.CommandPath()))
		defer utils.LogTime(fmt.Sprintf("cmd.%s.PostRun end", cmd.CommandPath()))
		// run the original PostRun
		if originalPostRun != nil {
			originalPostRun(cmd, args)
		}

		// run the post run
		postRunHook(cmd, args)
	}

	// wrap over the original Run function
	originalRun := cfg.cmd.Run
	cfg.cmd.Run = func(cmd *cobra.Command, args []string) {
		utils.LogTime(fmt.Sprintf("cmd.%s.Run start", cmd.CommandPath()))
		defer utils.LogTime(fmt.Sprintf("cmd.%s.Run end", cmd.CommandPath()))

		// run the original Run
		if originalRun != nil {
			originalRun(cmd, args)
		}
	}

	return cfg
}

// AddStringFlag is a helper function to add a string flag to a command
func (c *CmdBuilder) AddStringFlag(name string, defaultValue string, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().String(name, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}

	return c
}

// AddIntFlag is a helper function to add an integer flag to a command
func (c *CmdBuilder) AddIntFlag(name string, defaultValue int, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().Int(name, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}

// AddBoolFlag ia s helper function to add a boolean flag to a command
func (c *CmdBuilder) AddBoolFlag(name string, defaultValue bool, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().Bool(name, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}

// AddCloudFlags is helper function to add the cloud flags to a command
func (c *CmdBuilder) AddCloudFlags() *CmdBuilder {
	return c.
		AddStringFlag(constants.ArgPipesHost, constants.DefaultPipesHost, "Turbot Pipes host").
		AddStringFlag(constants.ArgPipesToken, "", "Turbot Pipes authentication token").
		AddStringFlag(constants.ArgCloudHost, constants.DefaultPipesHost, "Turbot Pipes host", FlagOptions.Deprecated(constants.ArgPipesHost)).
		AddStringFlag(constants.ArgCloudToken, "", "Turbot Pipes authentication token", FlagOptions.Deprecated(constants.ArgPipesToken))
}

// AddWorkspaceDatabaseFlag is helper function to add the workspace-databse flag to a command
func (c *CmdBuilder) AddWorkspaceDatabaseFlag() *CmdBuilder {
	return c.
		AddStringFlag(constants.ArgWorkspaceDatabase, constants.DefaultWorkspaceDatabase, "Turbot Pipes workspace database")
}

// AddModLocationFlag is helper function to add the mod-location flag to a command
func (c *CmdBuilder) AddModLocationFlag() *CmdBuilder {
	cwd, err := os.Getwd()
	error_helpers.FailOnError(err)
	return c.
		AddStringFlag(constants.ArgModLocation, cwd, "Path to the workspace working directory")
}

// AddStringSliceFlag is a helper function to add a flag that accepts an array of strings
func (c *CmdBuilder) AddStringSliceFlag(name string, defaultValue []string, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().StringSlice(name, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}

// AddStringArrayFlag is a helper function to add a flag that accepts an array of strings
func (c *CmdBuilder) AddStringArrayFlag(name string, defaultValue []string, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().StringArray(name, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}

// AddStringMapStringFlag is a helper function to add a flag that accepts a map of strings
func (c *CmdBuilder) AddStringMapStringFlag(name string, defaultValue map[string]string, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().StringToString(name, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}
