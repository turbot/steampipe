package connection_config

type Connection struct {
	Name   string
	Plugin string
	Config string
}

func NewConnection() *Connection {
	return &Connection{
		Name:   "",
		Plugin: "",
	}
}
