package connection_config

// SteampipeConfig :: Connection map and Steampipe settings
type SteampipeConfig struct {
	// map of connection name to partially parsed connection config
	Connections map[string]*Connection
	// Steampipe options
	ConnectionOptions *ConnectionOptions
	DatabaseOptions   *DatabaseOptions
	ConsoleOptions    *ConsoleOptions
	GeneralOptions    *GeneralOptions
}

func newSteampipeConfig() *SteampipeConfig {
	return &SteampipeConfig{
		Connections: make(map[string]*Connection),
	}
}

func (c *SteampipeConfig) SetOptions(options Options) {
	switch o := options.(type) {
	case *ConnectionOptions:
		c.ConnectionOptions = o
	case *DatabaseOptions:
		c.DatabaseOptions = o
	case *ConsoleOptions:
		c.ConsoleOptions = o
	case *GeneralOptions:
		c.GeneralOptions = o
	}
}

// global steampipe config
var Config *SteampipeConfig
