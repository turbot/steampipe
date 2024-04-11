package display

import (
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
)

type displayConfiguration struct {
	timing bool
}

// newDisplayConfiguration creates a default configuration with timing set to
// true if both --timing is not 'off' and --output is table
func newDisplayConfiguration() *displayConfiguration {
	timingFlag := cmdconfig.Viper().GetString(constants.ArgTiming) != constants.ArgOff
	isInteractive := cmdconfig.Viper().GetBool(constants.ConfigKeyInteractive)
	outputTable := cmdconfig.Viper().GetString(constants.ArgOutput) == constants.OutputFormatTable

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
