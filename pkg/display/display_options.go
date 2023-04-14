package display

import (
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
)

type displayConfiguration struct {
	timing bool
}

// NewDisplayConfiguration creates a default configuration with timing set to
// true if both --timing is true and --output is table
func NewDisplayConfiguration() *displayConfiguration {
	return &displayConfiguration{
		timing: cmdconfig.Viper().GetBool(constants.ArgTiming) && (cmdconfig.Viper().GetString(constants.ArgOutput) == constants.OutputFormatTable),
	}
}

type DisplayOption = func(config *displayConfiguration)

// WithTimingDisabled forcefully disables display of timing data
func WithTimingDisabled() DisplayOption {
	return func(o *displayConfiguration) {
		o.timing = false
	}
}
