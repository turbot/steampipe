package options

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/constants"
)

// Terminal
type Terminal struct {
	Output    *string `hcl:"output"`
	Separator *string `hcl:"separator"`
	Header    *bool   `hcl:"header"`
	Multi     *bool   `hcl:"multi"`
	Timing    *bool   `hcl:"timing"`
}

// ConfigMap :: create a config map to pass to viper
func (c *Terminal) ConfigMap() map[string]interface{} {
	// only add keys which are non null
	res := map[string]interface{}{}
	if c.Output != nil {
		res[constants.ArgOutput] = c.Output
	}
	if c.Separator != nil {
		res[constants.ArgSeparator] = c.Separator
	}
	if c.Header != nil {
		res[constants.ArgHeader] = c.Header
	}
	if c.Multi != nil {
		res[constants.ArgMultiLine] = c.Multi
	}
	if c.Timing != nil {
		res[constants.ArgTimer] = c.Timing
	}
	return res
}

func (c *Terminal) String() string {
	if c == nil {
		return ""
	}
	var str []string
	if c.Output == nil {
		str = append(str, "LogLevel: nil")
	} else {
		str = append(str, fmt.Sprintf("Output: %s", *c.Output))
	}
	if c.Separator == nil {
		str = append(str, "Separator: nil")
	} else {
		str = append(str, fmt.Sprintf("Separator: %s", *c.Separator))
	}
	if c.Header == nil {
		str = append(str, "Header: nil")
	} else {
		str = append(str, fmt.Sprintf("Header: %v", *c.Header))
	}
	if c.Multi == nil {
		str = append(str, "Multi: nil")
	} else {
		str = append(str, fmt.Sprintf("Multi: %v", *c.Multi))
	}
	if c.Timing == nil {
		str = append(str, "Timing: nil")
	} else {
		str = append(str, fmt.Sprintf("Timing: %v", *c.Timing))
	}
	return strings.Join(str, "\n")
}
