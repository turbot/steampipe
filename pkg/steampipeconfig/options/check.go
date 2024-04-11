package options

import (
	"fmt"
	"golang.org/x/exp/maps"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
)

// General
type Check struct {
	Output    *string `hcl:"output" cty:"check_output"`
	Separator *string `hcl:"separator" cty:"check_separator"`
	Header    *bool   `hcl:"header" cty:"check_header"`
	Timing    *string `hcl:"timing" cty:"check_timing"`
}

func (t *Check) SetBaseProperties(otherOptions Options) {
	if helpers.IsNil(otherOptions) {
		return
	}
	if o, ok := otherOptions.(*Check); ok {
		if t.Output == nil && o.Output != nil {
			t.Output = o.Output
		}
		if t.Separator == nil && o.Separator != nil {
			t.Separator = o.Separator
		}
		if t.Separator == nil && o.Separator != nil {
			t.Separator = o.Separator
		}
		if t.Header == nil && o.Header != nil {
			t.Header = o.Header
		}
	}
}

// ConfigMap creates a config map that can be merged with viper
func (t *Check) ConfigMap() map[string]interface{} {
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
	if t.Timing != nil {
		res[constants.ArgTiming] = t.Timing
	}
	return res
}

// Merge :: merge other options over the the top of this options object
// i.e. if a property is set in otherOptions, it takes precedence
func (t *Check) Merge(otherOptions Options) {
	switch o := otherOptions.(type) {
	case *Check:
		if o.Output != nil {
			t.Output = o.Output
		}
		if o.Separator != nil {
			t.Separator = o.Separator
		}
		if o.Header != nil {
			t.Header = o.Header
		}
		if o.Timing != nil {
			t.Timing = o.Timing
		}
	}
}

func (t *Check) String() string {
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
	if t.Timing == nil {
		str = append(str, "  Timing: nil")
	} else {
		str = append(str, fmt.Sprintf("  Timing: %v", *t.Timing))
	}
	return strings.Join(str, "\n")
}

func (t *Check) SetTiming(flag string, r hcl.Range) hcl.Diagnostics {
	// check the value is valid
	if _, ok := constants.CheckTimingValueLookup[flag]; !ok {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Invalid timing value '%s', check options support: %s", flag, strings.Join(maps.Keys(constants.CheckTimingValueLookup), ", ")),
				Subject:  &r,
			},
		}
	}
	t.Timing = &flag

	return nil
}
