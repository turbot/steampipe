package options

import (
	"fmt"
	"strings"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
)

// General
type Query struct {
	Output       *string `hcl:"output" cty:"query_output"`
	Separator    *string `hcl:"separator" cty:"query_separator"`
	Header       *bool   `hcl:"header" cty:"query_header"`
	Multi        *bool   `hcl:"multi" cty:"query_multi"`
	Timing       *bool   `hcl:"timing" cty:"query_timing"`
	AutoComplete *bool   `hcl:"autocomplete" cty:"query_autocomplete"`
}

func (t *Query) SetBaseProperties(otherOptions Options) {
	if helpers.IsNil(otherOptions) {
		return
	}
	if o, ok := otherOptions.(*Query); ok {
		if t.Output == nil && o.Output != nil {
			t.Output = o.Output
		}
		if t.Separator == nil && o.Separator != nil {
			t.Separator = o.Separator
		}
		if t.Header == nil && o.Header != nil {
			t.Header = o.Header
		}
		if t.Multi == nil && o.Multi != nil {
			t.Multi = o.Multi
		}
		if t.Timing == nil && o.Timing != nil {
			t.Timing = o.Timing
		}
		if t.AutoComplete == nil && o.AutoComplete != nil {
			t.AutoComplete = o.AutoComplete
		}
	}
}

// ConfigMap creates a config map that can be merged with viper
func (t *Query) ConfigMap() map[string]interface{} {
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
	if t.AutoComplete != nil {
		res[constants.ArgAutoComplete] = t.AutoComplete
	}
	return res
}

// Merge :: merge other options over the the top of this options object
// i.e. if a property is set in otherOptions, it takes precedence
func (t *Query) Merge(otherOptions Options) {
	if _, ok := otherOptions.(*Query); !ok {
		return
	}
	switch o := otherOptions.(type) {
	case *Query:
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
		if o.AutoComplete != nil {
			t.AutoComplete = o.AutoComplete
		}
	}
}

func (t *Query) String() string {
	if t == nil {
		return ""
	}
	var str []string
	if t.Output == nil {
		str = append(str, "  Output: nil")
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
	if t.AutoComplete == nil {
		str = append(str, "  AutoComplete: nil")
	} else {
		str = append(str, fmt.Sprintf("  AutoComplete: %v", *t.AutoComplete))
	}
	return strings.Join(str, "\n")
}
