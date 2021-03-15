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
	// only add keys which are non null
	res := map[string]interface{}{}
	if c.Port != nil {
		res[constants.ArgPort] = c.Port
	}
	if c.Listen != nil {
		res[constants.ArgListenAddress] = c.Listen
	}

	return res
}
