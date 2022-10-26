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
	QueryTimeout *int64  `hcl:"query_timeout"`
}

// ConfigMap :: create a config map to pass to viper
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
	if d.QueryTimeout != nil {
		res[constants.ArgDatabaseQueryTimeout] = d.QueryTimeout
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
		if o.QueryTimeout != nil {
			d.QueryTimeout = o.QueryTimeout
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
	if d.QueryTimeout == nil {
		str = append(str, "  QueryTimeout: nil")
	} else {
		str = append(str, fmt.Sprintf("  QueryTimeout: %d", *d.QueryTimeout))
	}
	return strings.Join(str, "\n")
}
