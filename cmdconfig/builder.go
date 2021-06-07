package cmdconfig

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/utils"
)

type CmdBuilder struct {
	cmd      *cobra.Command
	bindings map[string]*pflag.Flag
}

// OnCmd :: starts a config builder wrapping over the provided *cobra.Command
func OnCmd(cmd *cobra.Command) *CmdBuilder {
	cfg := new(CmdBuilder)
	cfg.cmd = cmd
	cfg.bindings = map[string]*pflag.Flag{}

	originalPreRun := cfg.cmd.PreRun
	originalRun := cfg.cmd.Run

	cfg.cmd.PreRun = func(cmd *cobra.Command, args []string) {
		// bind flags
		for flagName, flag := range cfg.bindings {
			viper.GetViper().BindPFlag(flagName, flag)
		}
		// run the original PreRun
		if originalPreRun != nil {
			originalPreRun(cmd, args)
		}
	}

	cfg.cmd.Run = func(cmd *cobra.Command, args []string) {
		utils.LogTime(fmt.Sprintf("cmd.%s.Run start", cmd.CommandPath()))
		defer utils.LogTime(fmt.Sprintf("cmd.%s.Run end", cmd.CommandPath()))
		originalRun(cmd, args)
	}

	return cfg
}

// Helper function to add a string flag to a command
func (c *CmdBuilder) AddStringFlag(name string, shorthand string, defaultValue string, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().StringP(name, shorthand, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}

	return c
}

// Helper function to add an integer flag to a command
func (c *CmdBuilder) AddIntFlag(name, shorthand string, defaultValue int, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().IntP(name, shorthand, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}

// Helper function to add a boolean flag to a command
func (c *CmdBuilder) AddBoolFlag(name, shorthand string, defaultValue bool, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().BoolP(name, shorthand, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}

// Helper function to add a flag that accepts an array of strings
func (c *CmdBuilder) AddStringSliceFlag(name, shorthand string, defaultValue []string, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().StringSliceP(name, shorthand, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}

// Helper function to add a flag that accepts a map of strings
func (c *CmdBuilder) AddStringMapStringFlag(name, shorthand string, defaultValue map[string]string, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().StringToStringP(name, shorthand, defaultValue, desc)
	c.bindings[name] = c.cmd.Flags().Lookup(name)
	for _, o := range opts {
		o(c.cmd, name, name)
	}
	return c
}
