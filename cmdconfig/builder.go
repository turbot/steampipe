package cmdconfig

import (
	"github.com/spf13/cobra"
)

type CmdBuilder struct {
	cmd      *cobra.Command
	vWrapper *ViperWrapper
}

// OnCmd :: starts a config builder wrapping over the provided *cobra.Command
func OnCmd(cmd *cobra.Command) *CmdBuilder {

	if cmd.Run == nil {
		panic("Run needs to be present for configuration")
	}

	cfg := new(CmdBuilder)
	cfg.cmd = cmd
	cfg.vWrapper = NewViperWrapper(cfg.cmd)

	InitViper(cfg.vWrapper)

	originalRun := cfg.cmd.Run

	cfg.cmd.Run = func(cmd *cobra.Command, args []string) {
		setConfig(cfg.vWrapper)
		originalRun(cmd, args)
	}

	return cfg
}

// Helper function to add a string flag to a command
func (c *CmdBuilder) AddStringFlag(name string, shorthand string, def string, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().StringP(name, shorthand, def, desc)
	c.vWrapper.BindPFlag(name, c.cmd.Flags().Lookup(name))
	for _, o := range opts {
		o(c.cmd, name, name, c.vWrapper)
	}

	return c
}

// Helper function to add an integer flag to a command
func (c *CmdBuilder) AddIntFlag(name, shorthand string, def int, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().IntP(name, shorthand, def, desc)
	c.vWrapper.BindPFlag(name, c.cmd.Flags().Lookup(name))
	for _, o := range opts {
		o(c.cmd, name, name, c.vWrapper)
	}
	return c
}

// Helper function to add a boolean flag to a command
func (c *CmdBuilder) AddBoolFlag(name, shorthand string, def bool, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().BoolP(name, shorthand, def, desc)
	c.vWrapper.BindPFlag(name, c.cmd.Flags().Lookup(name))
	for _, o := range opts {
		o(c.cmd, name, name, c.vWrapper)
	}
	return c
}

// Helper function to add a flag that accepts an array of strings
func (c *CmdBuilder) AddStringSliceFlag(name, shorthand string, def []string, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().StringSliceP(name, shorthand, def, desc)
	c.vWrapper.BindPFlag(name, c.cmd.Flags().Lookup(name))
	for _, o := range opts {
		o(c.cmd, name, name, c.vWrapper)
	}
	return c
}

// Helper function to add a flag that accepts a map of strings
func (c *CmdBuilder) AddStringMapStringFlag(name, shorthand string, def map[string]string, desc string, opts ...flagOpt) *CmdBuilder {
	c.cmd.Flags().StringToStringP(name, shorthand, def, desc)
	c.vWrapper.BindPFlag(name, c.cmd.Flags().Lookup(name))
	for _, o := range opts {
		o(c.cmd, name, name, c.vWrapper)
	}
	return c
}
