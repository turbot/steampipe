package connection_config

type Connection struct {
	Name   string
	Plugin string
	Config map[string]string
}

func NewConnection() *Connection {
	return &Connection{
		Name:   "",
		Plugin: "",
		Config: make(map[string]string),
	}
}
