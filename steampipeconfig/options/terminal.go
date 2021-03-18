package options

import (
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
