package cmdconfig

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
			viper.GetViper().BindPFlag(flagName, flag)
		}
		// run the original PreRun
		if originalPreRun != nil {
			originalPreRun(cmd, args)
		}
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
func (c *CmdBuilder) AddStringFlag(name string, shorthand string, defaultValue string, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().StringP(name, shorthand, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}

	return c
}

// AddIntFlag is a helper function to add an integer flag to a command
func (c *CmdBuilder) AddIntFlag(name, shorthand string, defaultValue int, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().IntP(name, shorthand, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}

// AddBoolFlag ia s helper function to add a boolean flag to a command
func (c *CmdBuilder) AddBoolFlag(name, shorthand string, defaultValue bool, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().BoolP(name, shorthand, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}

// AddStringSliceFlag is a helper function to add a flag that accepts an array of strings
func (c *CmdBuilder) AddStringSliceFlag(name, shorthand string, defaultValue []string, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().StringSliceP(name, shorthand, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}

// AddStringArrayFlag is a helper function to add a flag that accepts an array of strings
func (c *CmdBuilder) AddStringArrayFlag(name, shorthand string, defaultValue []string, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().StringArrayP(name, shorthand, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}

// AddStringMapStringFlag is a helper function to add a flag that accepts a map of strings
func (c *CmdBuilder) AddStringMapStringFlag(name, shorthand string, defaultValue map[string]string, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().StringToStringP(name, shorthand, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}
