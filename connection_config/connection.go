package connection_config

// Connection :: structure representing the partially parsed connection.
type Connection struct {
	// connection name
	Name string
	// FQN of plugin
	Plugin string
	// unparsed HCL of plugin specific connection config
	Config string
}
