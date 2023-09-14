package connection

import (
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"golang.org/x/exp/maps"
)

// PluginLimiterMap map of plugin image ref to Limiter map for the plugin
type PluginLimiterMap map[string]LimiterMap

func (l PluginLimiterMap) Equals(other PluginLimiterMap) bool {
	return maps.EqualFunc(l, other, func(m1, m2 LimiterMap) bool { return m1.Equals(m2) })
}

type PluginMap map[string]*modconfig.Plugin

func (p PluginMap) ToPluginLimiterMap() PluginLimiterMap {
	var limiterPluginMap = make(PluginLimiterMap)
	for pluginConfigLabel, p := range p {
		if len(p.Limiters) > 0 {
			limiterPluginMap[pluginConfigLabel] = NewLimiterMap(p.Limiters)
		}
	}
	return limiterPluginMap
}
