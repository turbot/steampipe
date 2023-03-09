package options

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/pkg/constants"
)

// Terminal
type Terminal struct {
	Output           *string `hcl:"output"`
	Separator        *string `hcl:"separator"`
	Header           *bool   `hcl:"header"`
	Multi            *bool   `hcl:"multi"`
	Timing           *bool   `hcl:"timing"`
	SearchPath       *string `hcl:"search_path"`
	SearchPathPrefix *string `hcl:"search_path_prefix"`
	Watch            *bool   `hcl:"watch"`
	AutoComplete     *bool   `hcl:"autocomplete"`
}

// ConfigMap :: create a config map to pass to viper
func (t *Terminal) ConfigMap() map[string]interface{} {
	// only add keys which are non null
	res := map[string]interface{}{}
	if t.Output != nil {
		res[constants.ArgOutput] = t.Output
	}
	if t.Separator != nil {
		res[constants.ArgSeparator] = t.Separator
	}
	if t.Header != nil {
		res[constants.ArgHeader] = t.Header
	}
	if t.Multi != nil {
		res[constants.ArgMultiLine] = t.Multi
	}
	if t.Timing != nil {
		res[constants.ArgTiming] = t.Timing
	}
	if t.SearchPath != nil {
		// convert from string to array
		res[constants.ArgSearchPath] = searchPathToArray(*t.SearchPath)
	}
	if t.SearchPathPrefix != nil {
		// convert from string to array
		res[constants.ArgSearchPathPrefix] = searchPathToArray(*t.SearchPathPrefix)
	}
	if t.Watch != nil {
		res[constants.ArgWatch] = t.Watch
	}
	if t.AutoComplete != nil {
		res[constants.ArgAutoComplete] = t.AutoComplete
	}
	return res
}

// merge other options over the the top of this options object
// i.e. if a property is set in otherOptions, it takes precedence
func (t *Terminal) Merge(otherOptions Options) {
	switch o := otherOptions.(type) {
	case *Terminal:
		if o.Output != nil {
			t.Output = o.Output
		}
		if o.Separator != nil {
			t.Separator = o.Separator
		}
		if o.Header != nil {
			t.Header = o.Header
		}
		if o.Multi != nil {
			t.Multi = o.Multi
		}
		if o.Timing != nil {
			t.Timing = o.Timing
		}
		if o.SearchPath != nil {
			t.SearchPath = o.SearchPath
		}
		if o.SearchPathPrefix != nil {
			t.SearchPathPrefix = o.SearchPathPrefix
		}
		if o.Watch != nil {
			t.Watch = o.Watch
		}
		if o.AutoComplete != nil {
			t.AutoComplete = o.AutoComplete
		}
	}
}

func (t *Terminal) String() string {
	if t == nil {
		return ""
	}
	var str []string
	if t.Output == nil {
		str = append(str, "  LogLevel: nil")
	} else {
		str = append(str, fmt.Sprintf("  Output: %s", *t.Output))
	}
	if t.Separator == nil {
		str = append(str, "  Separator: nil")
	} else {
		str = append(str, fmt.Sprintf("  Separator: %s", *t.Separator))
	}
	if t.Header == nil {
		str = append(str, "  Header: nil")
	} else {
		str = append(str, fmt.Sprintf("  Header: %v", *t.Header))
	}
	if t.Multi == nil {
		str = append(str, "  Multi: nil")
	} else {
		str = append(str, fmt.Sprintf("  Multi: %v", *t.Multi))
	}
	if t.Timing == nil {
		str = append(str, "  Timing: nil")
	} else {
		str = append(str, fmt.Sprintf("  Timing: %v", *t.Timing))
	}
	if t.SearchPath == nil {
		str = append(str, "  SearchPath: nil")
	} else {
		str = append(str, fmt.Sprintf("  SearchPath: %s", *t.SearchPath))
	}
	if t.SearchPathPrefix == nil {
		str = append(str, "  SearchPathPrefix: nil")
	} else {
		str = append(str, fmt.Sprintf("  SearchPathPrefix: %s", *t.SearchPathPrefix))
	}
	if t.Watch == nil {
		str = append(str, "  Watch: nil")
	} else {
		str = append(str, fmt.Sprintf("  Watch: %v", *t.Watch))
	}
	if t.AutoComplete == nil {
		str = append(str, "  AutoComplete: nil")
	} else {
		str = append(str, fmt.Sprintf("  AutoComplete: %v", *t.AutoComplete))
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
