package connection_config

import (
	"github.com/spf13/viper"
)

// SteampipeConfig :: Connection map and Steampipe settings
type SteampipeConfig struct {
	// map of connection name to partially parsed connection config
	Connections map[string]*Connection
	// Steampipe settings
	Settings *Settings
}

func newSteampipeConfig() *SteampipeConfig {
	return &SteampipeConfig{
		Connections: make(map[string]*Connection),
	}
}

func (c SteampipeConfig) PopulateViper(v *viper.Viper) {
	if c.Settings != nil {
		c.Settings.PopulateViper(v)
	}
}

// global steampipe config
var Config *SteampipeConfig
