package connection_config

import (
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
)

// hcl options types
const (
	HclOptionsFdw     = "fdw"
	HclOptionsPlugin  = "plugin"
	HclOptionsConsole = "console"
)

type Options interface {
	PopulateViper(v *viper.Viper)
}

type FdwOptions struct {
	Cache    *bool `hcl:"cache"`
	CacheTTL *int  `hcl:"cache_ttl"`
}

func (f FdwOptions) PopulateViper(v *viper.Viper) {
	v.Set(constants.OptionCache, f.Cache)
	v.Set(constants.OptionCacheTTL, f.CacheTTL)

}
func (f FdwOptions) equals(other *FdwOptions) bool {
	//todo
	return false
}

// NOTE: this must be consistent with the protobuf PluginOptions type defined in the sdk
type PluginOptions struct {
	RLimitFiles int `hcl:"rlimit_files"`
}

func (f PluginOptions) equals(other *PluginOptions) bool {
	//todo
	return false
}

func (f PluginOptions) PopulateViper(v *viper.Viper) {
	v.Set(constants.RLimitFiles, f.RLimitFiles)

}

type ConsoleOptions struct {
	MultiLine bool `hcl:"multi"`
}

func (f ConsoleOptions) PopulateViper(v *viper.Viper) {
	v.Set(constants.MultiLine, f.MultiLine)

}
