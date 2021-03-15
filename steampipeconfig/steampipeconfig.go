package steampipeconfig

import "github.com/turbot/steampipe/steampipeconfig/options"

// SteampipeConfig :: Connection map and Steampipe settings
type SteampipeConfig struct {
	// map of connection name to partially parsed connection config
	Connections map[string]*Connection

	// Steampipe options
	DefaultConnectionOptions *options.Connection
	DatabaseOptions          *options.Database
	ConsoleOptions           *options.Console
	GeneralOptions           *options.General
}

func newSteampipeConfig() *SteampipeConfig {
	return &SteampipeConfig{
		Connections: make(map[string]*Connection),
	}
}

// ConfigMap :: create a config map to pass to viper
func (c *SteampipeConfig) ConfigMap() map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range c.DefaultConnectionOptions.ConfigMap() {
		res[k] = v
	}
	for k, v := range c.DatabaseOptions.ConfigMap() {
		res[k] = v
	}
	for k, v := range c.ConsoleOptions.ConfigMap() {
		res[k] = v
	}
	for k, v := range c.GeneralOptions.ConfigMap() {
		res[k] = v
	}
	return res
}

func (c *SteampipeConfig) SetOptions(opts options.Options) {
	switch o := opts.(type) {
	case *options.Connection:
		c.DefaultConnectionOptions = o
	case *options.Database:
		c.DatabaseOptions = o
	case *options.Console:
		c.ConsoleOptions = o
	case *options.General:
		c.GeneralOptions = o
	}
}

// global steampipe config
var Config *SteampipeConfig
