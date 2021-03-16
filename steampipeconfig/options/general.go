package options

import "github.com/turbot/steampipe/constants"

// General
type General struct {
	LogLevel    *string `hcl:"log_level"`
	UpdateCheck *string `hcl:"update_check"`
}

// Populate :: nothing to do
func (d General) Populate() {}

// ConfigMap :: create a config map to pass to viper
func (c General) ConfigMap() map[string]interface{} {
	// we can add null values as SteampipeConfig.ConfigMap() will ignore them
	return map[string]interface{}{
		constants.ArgLogLevel:    c.LogLevel,
		constants.ArgUpdateCheck: c.UpdateCheck,
	}
}
