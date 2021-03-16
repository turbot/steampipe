package options

import "github.com/turbot/steampipe/constants"

// Database
type Database struct {
	Port   *int    `hcl:"port"`
	Listen *string `hcl:"listen"`
}

// Populate :: nothing to do
func (d Database) Populate() {}

// ConfigMap :: create a config map to pass to viper
func (c *Database) ConfigMap() map[string]interface{} {
	// we can add null values as SteampipeConfig.ConfigMap() will ignore them
	return map[string]interface{}{
		constants.ArgPort:          c.Port,
		constants.ArgListenAddress: c.Listen,
	}

}
