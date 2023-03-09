package options

import (
	"fmt"
	"strings"
)

// Connection is a struct representing connection options
// json tags needed as this is stored in the connection state file
type Connection struct {
	Cache    *bool `hcl:"cache" json:"cache,omitempty"`
	CacheTTL *int  `hcl:"cache_ttl" json:"cache_ttl,omitempty"`

	// legacy properties included for backwards compatibility with v0.13
	LegacyCache    *bool `json:"Cache,omitempty"`
	LegacyCacheTTL *int  `json:"CacheTTL,omitempty"`
}

// MigrateLegacy migrates the legacy properties into new properties
func (c *Connection) MigrateLegacy() {
	c.Cache = c.LegacyCache
	c.CacheTTL = c.LegacyCacheTTL
}

func (c *Connection) ConfigMap() map[string]interface{} {
	// not implemented - we do not pass this config to viper
	return map[string]interface{}{}
}

// Merge merges other options over the top of this options object
// i.e. if a property is set in otherOptions, it takes precedence
func (c *Connection) Merge(otherOptions Options) {
	switch o := otherOptions.(type) {
	case *Connection:
		if o.Cache != nil {
			c.Cache = o.Cache
		}
		if o.CacheTTL != nil {
			c.CacheTTL = o.CacheTTL
		}
	}
}

func (c *Connection) Equals(other *Connection) bool {
	return c.String() == other.String()
}

func (c *Connection) String() string {
	if c == nil {
		return ""
	}
	var str []string
	if c.Cache == nil {
		str = append(str, "  Cache: nil")
	} else {
		str = append(str, fmt.Sprintf("  Cache: %v", *c.Cache))
	}
	if c.CacheTTL == nil {
		str = append(str, "  CacheTTL: nil")
	} else {
		str = append(str, fmt.Sprintf("  CacheTTL: %d", *c.CacheTTL))
	}
	return strings.Join(str, "\n")
}
