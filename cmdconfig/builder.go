package cmdconfig

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type configBuilder struct {
	cmd   *cobra.Command
	viper *viper.Viper
}

// OnCmd :: starts a config builder wrapping over the provided *cobra.Command
func OnCmd(cmd *cobra.Command) *configBuilder {

	if cmd.Run == nil {
		panic("Run needs to be present for configuration")
	}

	cfg := new(configBuilder)
	cfg.cmd = cmd
	cfg.viper = viper.New()

	InitViper(cfg.viper)

	originalRun := cfg.cmd.Run

	cfg.cmd.Run = func(cmd *cobra.Command, args []string) {
		setConfig(cfg.viper)
		originalRun(cmd, args)
	}

	return cfg
}

// Helper function to add a string flag to a command
func (c *configBuilder) AddStringFlag(name string, shorthand string, def string, desc string, opts ...flagOpt) *configBuilder {
	fn := name
	c.cmd.Flags().StringP(name, shorthand, def, desc)
	c.viper.BindPFlag(fn, c.cmd.Flags().Lookup(name))
	for _, o := range opts {
		o(c.cmd, name, fn, c.viper)
	}

	return c
}

// Helper function to add an integer flag to a command
func (c *configBuilder) AddIntFlag(name, shorthand string, def int, desc string, opts ...flagOpt) *configBuilder {
	fn := name
	c.cmd.Flags().IntP(name, shorthand, def, desc)
	c.viper.BindPFlag(fn, c.cmd.Flags().Lookup(name))
	for _, o := range opts {
		o(c.cmd, name, fn, c.viper)
	}
	return c
}

// Helper function to add a boolean flag to a command
func (c *configBuilder) AddBoolFlag(name, shorthand string, def bool, desc string, opts ...flagOpt) *configBuilder {
	fn := name
	c.cmd.Flags().BoolP(name, shorthand, def, desc)
	c.viper.BindPFlag(fn, c.cmd.Flags().Lookup(name))
	for _, o := range opts {
		o(c.cmd, name, fn, c.viper)
	}
	return c
}

// Helper function to add a flag that accepts an array of strings
func (c *configBuilder) AddStringSliceFlag(name, shorthand string, def []string, desc string, opts ...flagOpt) *configBuilder {
	fn := name
	c.cmd.Flags().StringSliceP(name, shorthand, def, desc)
	c.viper.BindPFlag(fn, c.cmd.Flags().Lookup(name))
	for _, o := range opts {
		o(c.cmd, name, fn, c.viper)
	}
	return c
}

// Helper function to add a flag that accepts a map of strings
func (c *configBuilder) AddStringMapStringFlag(name, shorthand string, def map[string]string, desc string, opts ...flagOpt) *configBuilder {
	fn := name
	c.cmd.Flags().StringToStringP(name, shorthand, def, desc)
	c.viper.BindPFlag(fn, c.cmd.Flags().Lookup(name))
	for _, o := range opts {
		o(c.cmd, name, fn, c.viper)
	}
	return c
}
