package options

// options.Connection
type Connection struct {
	Cache    *bool `hcl:"cache"`
	CacheTTL *int  `hcl:"cache_ttl"`
}

func (c *Connection) ConfigMap() map[string]interface{} {
	// not implemented - we do not pass this config to viper
	return map[string]interface{}{}
}

func (c *Connection) Equals(other *Connection) bool {
	return c.Cache == other.Cache &&
		*c.CacheTTL == *other.CacheTTL
}
