package options

// hcl options block types
const (
	ConnectionBlock = "connection"
	DatabaseBlock   = "database"
	GeneralBlock    = "general"
	TerminalBlock   = "terminal"
)

type Options interface {
	// map of config keys to values - used to populate viper
	ConfigMap() map[string]interface{}
	// merge with another options of same type
	Merge(otherOptions Options)
}
