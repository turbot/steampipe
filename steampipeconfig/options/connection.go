package options

import (
	"fmt"
	"strings"
)

// options.Connection
type Connection struct {
	Cache    *bool `hcl:"cache"`
	CacheTTL *int  `hcl:"cache_ttl"`
}

func (c *Connection) ConfigMap() map[string]interface{} {
	// not implemented - we do not pass this config to viper
	return map[string]interface{}{}
}

// merge other options over the the top of this options object
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
	return c.Cache == other.Cache &&
		*c.CacheTTL == *other.CacheTTL
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
