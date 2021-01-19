package connection_config

type ConnectionConfig struct {
	Connections map[string]*Connection
}

func newConfig() *ConnectionConfig {
	return &ConnectionConfig{
		Connections: make(map[string]*Connection),
	}
}
