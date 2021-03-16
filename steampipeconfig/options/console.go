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
	// we can add null values as SteampipeConfig.ConfigMap() will ignore them
	return map[string]interface{}{
		constants.ArgOutput:    c.Output,
		constants.ArgSeparator: c.Separator,
		constants.ArgHeader:    c.Header,
		constants.ArgOutput:    c.Output,
		constants.ArgOutput:    c.Output,
	}
}
