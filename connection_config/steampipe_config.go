package connection_config

import (
	"github.com/spf13/viper"
)

// SteampipeConfig :: Connection map and Steampipe settings
type SteampipeConfig struct {
	// map of connection name to partially parsed connection config
	Connections map[string]*Connection
	// Steampipe options
	FdwOptions     *FdwOptions
	PluginOptions  *PluginOptions
	ConsoleOptions *ConsoleOptions
}

func newSteampipeConfig() *SteampipeConfig {
	return &SteampipeConfig{
		Connections: make(map[string]*Connection),
	}
}

func (c SteampipeConfig) PopulateViper(v *viper.Viper) {
	if c.FdwOptions != nil {
		c.FdwOptions.PopulateViper(v)
	}
	if c.PluginOptions != nil {
		c.PluginOptions.PopulateViper(v)
	}
	if c.ConsoleOptions != nil {
		c.ConsoleOptions.PopulateViper(v)
	}
}

func (c SteampipeConfig) SetOptions(options Options) {
	switch o := options.(type) {
	case *FdwOptions:
		c.FdwOptions = o
	case *PluginOptions:
		c.PluginOptions = o
	case *ConsoleOptions:
		c.ConsoleOptions = o
	}
}

// global steampipe config
var Config *SteampipeConfig
