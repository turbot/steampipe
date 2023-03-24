package options

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/pkg/constants"
)

// Database
type Database struct {
	Port         *int    `hcl:"port"`
	Listen       *string `hcl:"listen"`
	SearchPath   *string `hcl:"search_path"`
	StartTimeout *int    `hcl:"start_timeout"`

	SearchPathPrefix *string `hcl:"search_path_prefix"`
	Cache            *bool   `hcl:"cache"`
	CacheMaxTtl      *int    `hcl:"cache_max_ttl"`
	CacheMaxSizeMb   *int    `hcl:"cache_max_size_mb"`
}

// ConfigMap creates a config map that can be merged with viper
func (d *Database) ConfigMap() map[string]interface{} {
	// only add keys which are non null
	res := map[string]interface{}{}
	if d.Port != nil {
		res[constants.ArgDatabasePort] = d.Port
	}
	if d.Listen != nil {
		res[constants.ArgListenAddress] = d.Listen
	}
	if d.SearchPath != nil {
		// convert from string to array
		res[constants.ArgSearchPath] = searchPathToArray(*d.SearchPath)
	}
	if d.StartTimeout != nil {
		res[constants.ArgDatabaseStartTimeout] = d.StartTimeout
	} else {
		res[constants.ArgDatabaseStartTimeout] = constants.DBStartTimeout.Seconds()
	}

	if d.SearchPathPrefix != nil {
		// convert from string to array
		res[constants.ArgSearchPathPrefix] = searchPathToArray(*d.SearchPathPrefix)
	}
	if d.Cache != nil {
		res[constants.ArgServiceCacheEnabled] = d.Cache
	}
	if d.CacheMaxTtl != nil {
		res[constants.ArgCacheMaxTtl] = d.CacheMaxTtl
	}
	if d.CacheMaxSizeMb != nil {
		res[constants.ArgMaxCacheSizeMb] = d.CacheMaxSizeMb
	}
	return res
}

// Merge ::  merge other options over the the top of this options object
// i.e. if a property is set in otherOptions, it takes precedence
func (d *Database) Merge(otherOptions Options) {
	switch o := otherOptions.(type) {
	case *Database:
		if o.Port != nil {
			d.Port = o.Port
		}
		if o.Listen != nil {
			d.Listen = o.Listen
		}
		if o.SearchPath != nil {
			d.SearchPath = o.SearchPath
		}
	}
}

func (d *Database) String() string {
	if d == nil {
		return ""
	}
	var str []string
	if d.Port == nil {
		str = append(str, "  Port: nil")
	} else {
		str = append(str, fmt.Sprintf("  Port: %d", *d.Port))
	}
	if d.Listen == nil {
		str = append(str, "  Listen: nil")
	} else {
		str = append(str, fmt.Sprintf("  Listen: %s", *d.Listen))
	}
	if d.SearchPath == nil {
		str = append(str, "  SearchPath: nil")
	} else {
		str = append(str, fmt.Sprintf("  SearchPath: %s", *d.SearchPath))
	}
	if d.StartTimeout == nil {
		str = append(str, "  ServiceStartTimeout: nil")
	} else {
		str = append(str, fmt.Sprintf("  ServiceStartTimeout: %d", *d.StartTimeout))
	}
	return strings.Join(str, "\n")
}
