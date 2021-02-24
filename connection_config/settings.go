package connection_config

import (
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
)

type Settings struct {
	Cache    *bool `hcl:"cache"`
	CacheTTL *int  `hcl:"cache_ttl"`
}

func (s Settings) PopulateViper(v *viper.Viper) {
	v.Set(constants.SettingCache, s.Cache)
	v.Set(constants.SettingCacheTTL, s.CacheTTL)

}
