package modconfig

type ConnectionGroup struct {
	// connection name
	Name string `hcl:"name,label"`
	// Name of plugin
	Plugin      string   `hcl:"plugin"`
	Connections []string `hcl:"connections"`
}
