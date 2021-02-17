package connection_config

// ConnectionConfigMap :: map of connection name to partially parsed connection config
type ConnectionConfigMap struct {
	Connections map[string]*Connection
}

func newConfigMap() *ConnectionConfigMap {
	return &ConnectionConfigMap{
		Connections: make(map[string]*Connection),
	}
}
