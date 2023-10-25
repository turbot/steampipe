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
	timingFlag := cmdconfig.Viper().GetBool(constants.ArgTiming)
	isInteractive := cmdconfig.Viper().GetBool(constants_steampipe.ConfigKeyInteractive)
	outputTable := cmdconfig.Viper().GetString(constants.ArgOutput) == constants_steampipe.OutputFormatTable

	timing := timingFlag && (outputTable || isInteractive)

	return &displayConfiguration{
		timing: timing,
	}
}

type DisplayOption = func(config *displayConfiguration)

// WithTimingDisabled forcefully disables display of timing data
func WithTimingDisabled() DisplayOption {
	return func(o *displayConfiguration) {
		o.timing = false
	}
}
