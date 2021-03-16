package options

import (
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/constants"
)

// Console
type Console struct {
	Output    *string `hcl:"output"`
	Separator *string `hcl:"separator"`
	// strings containing a bool - supports true/false/off/on etc
	HeaderBoolString *string `hcl:"header"`
	MultiBoolString  *string `hcl:"multi"`
	TimingBoolString *string `hcl:"timing"`

	// fields which we populate by converting the parsed values
	Header *bool
	Multi  *bool
	Timing *bool
}

// Populate :: convert strings representing bool values into bool pointers
func (c *Console) Populate() {
	c.Header = types.ToBoolPtr(c.HeaderBoolString)
	c.Multi = types.ToBoolPtr(c.MultiBoolString)
	c.Timing = types.ToBoolPtr(c.TimingBoolString)
}

// ConfigMap :: create a config map to pass to viper
func (c *Console) ConfigMap() map[string]interface{} {
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
	if c.Output != nil {
		res[constants.ArgOutput] = c.Output
	}
	if c.Output != nil {
		res[constants.ArgOutput] = c.Output
	}
	return res
}
