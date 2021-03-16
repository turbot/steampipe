package options

// hcl options block types
const (
	ConnectionBlock = "connection"
	DatabaseBlock   = "database"
	GeneralBlock    = "general"
	ConsoleBlock    = "console"
)

type Options interface {
	// once we have parsed the hcl, we may need to convert the parsed values
	// - for example we accept true/false/on/off for bool values - convert these to bool
	Populate()
	// map of config keys to values - used to populate viper
	ConfigMap() map[string]interface{}
}
