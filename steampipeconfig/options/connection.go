package options

import "github.com/turbot/go-kit/types"

// options.Connection
type Connection struct {
	// string containing a bool - supports true/false/off/on etc
	CacheBoolString *string `hcl:"cache"`
	CacheTTL        *int    `hcl:"cache_ttl"`

	// fields which we populate by converting the parsed values
	Cache *bool
}

// Populate :: convert strings representing bool values into bool pointers
func (c *Connection) Populate() {
	// convert CacheBoolString to a bool ptr
	c.Cache = types.ToBoolPtr(c.CacheBoolString)
}

func (c *Connection) ConfigMap() map[string]interface{} {
	// not implemented - we do not pass this config to viper
	return map[string]interface{}{}
}

func (c *Connection) Equals(other *Connection) bool {
	return c.Cache == other.Cache &&
		*c.CacheTTL == *other.CacheTTL
}
