package options

import (
	"fmt"
	"strings"

	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/options"
)

type Database struct {
	Cache            *bool   `hcl:"cache"`
	CacheMaxTtl      *int    `hcl:"cache_max_ttl"`
	CacheMaxSizeMb   *int    `hcl:"cache_max_size_mb"`
	Listen           *string `hcl:"listen"`
	Port             *int    `hcl:"port"`
	SearchPath       *string `hcl:"search_path"`
	SearchPathPrefix *string `hcl:"search_path_prefix"`
	StartTimeout     *int    `hcl:"start_timeout"`
}

// ConfigMap creates a config map that can be merged with viper
func (d *Database) ConfigMap() map[string]interface{} {
	// only add keys which are non null
	res := map[string]interface{}{}
	if d.Listen != nil {
		res[constants.ArgDatabaseListenAddresses] = d.Listen
	}
	if d.Port != nil {
		res[constants.ArgDatabasePort] = d.Port
	}
	if d.SearchPath != nil {
		// convert from string to array
		res[constants.ConfigKeyServerSearchPath] = searchPathToArray(*d.SearchPath)
	}
	if d.SearchPathPrefix != nil {
		// convert from string to array
		res[constants.ConfigKeyServerSearchPathPrefix] = searchPathToArray(*d.SearchPathPrefix)
	}
	if d.StartTimeout != nil {
		res[constants.ArgDatabaseStartTimeout] = d.StartTimeout
	} else {
		res[constants.ArgDatabaseStartTimeout] = constants.DBStartTimeout.Seconds()
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
func (d *Database) Merge(otherOptions options.Options) {
	switch o := otherOptions.(type) {
	case *Database:
		if o.Listen != nil {
			d.Listen = o.Listen
		}
		if o.Port != nil {
			d.Port = o.Port
		}
		if o.SearchPath != nil {
			d.SearchPath = o.SearchPath
		}
		if o.StartTimeout != nil {
			d.StartTimeout = o.StartTimeout
		}
		if o.SearchPathPrefix != nil {
			d.SearchPathPrefix = o.SearchPathPrefix
		}
		if o.Cache != nil {
			d.Cache = o.Cache
		}
		if o.CacheMaxSizeMb != nil {
			d.CacheMaxSizeMb = o.CacheMaxSizeMb
		}
		if o.CacheMaxTtl != nil {
			d.CacheMaxTtl = o.CacheMaxTtl
		}
	}
}

func (d *Database) String() string {
	if d == nil {
		return ""
	}
	var str []string
	if d.Listen == nil {
		str = append(str, "  Listen: nil")
	} else {
		str = append(str, fmt.Sprintf("  Listen: %s", *d.Listen))
	}
	if d.Port == nil {
		str = append(str, "  Port: nil")
	} else {
		str = append(str, fmt.Sprintf("  Port: %d", *d.Port))
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
	if d.SearchPathPrefix == nil {
		str = append(str, "  SearchPathPrefix: nil")
	} else {
		str = append(str, fmt.Sprintf("  SearchPathPrefix: %s", *d.SearchPathPrefix))
	}
	if d.Cache == nil {
		str = append(str, "  Cache: nil")
	} else {
		str = append(str, fmt.Sprintf("  Cache: %t", *d.Cache))
	}
	if d.CacheMaxSizeMb == nil {
		str = append(str, "  CacheMaxSizeMb: nil")
	} else {
		str = append(str, fmt.Sprintf("  CacheMaxSizeMb: %d", *d.CacheMaxSizeMb))
	}
	if d.CacheMaxTtl == nil {
		str = append(str, "  CacheMaxTtl: nil")
	} else {
		str = append(str, fmt.Sprintf("  CacheMaxTtl: %d", *d.CacheMaxTtl))
	}
	return strings.Join(str, "\n")
}

func searchPathToArray(searchPathString string) []string {
	// convert comma separated list to array
	searchPath := strings.Split(searchPathString, ",")
	// strip whitespace
	for i, s := range searchPath {
		searchPath[i] = strings.TrimSpace(s)
	}
	return searchPath
}
