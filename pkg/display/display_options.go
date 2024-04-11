package display

import (
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
)

type displayConfiguration struct {
	timing string
}

// newDisplayConfiguration creates a default configuration with timing set to
// true if both --timing is not 'off' and --output is table
func newDisplayConfiguration() *displayConfiguration {
	return &displayConfiguration{
		timing: cmdconfig.Viper().GetString(constants.ArgTiming),
	}
}

type DisplayOption = func(config *displayConfiguration)

// WithTimingDisabled forcefully disables display of timing data
func WithTimingDisabled() DisplayOption {
	return func(o *displayConfiguration) {
		o.timing = constants.ArgOff
	}
}
