package connection_config

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

// global steampipe config
var Config *SteampipeConfig
