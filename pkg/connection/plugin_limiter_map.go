package connection

import (
	"github.com/turbot/pipe-fittings/v2/plugin"
	"golang.org/x/exp/maps"
)

// PluginLimiterMap map of plugin image ref to Limiter map for the plugin
type PluginLimiterMap map[string]LimiterMap

func (l PluginLimiterMap) Equals(other PluginLimiterMap) bool {
	return maps.EqualFunc(l, other, func(m1, m2 LimiterMap) bool { return m1.Equals(m2) })
}

type PluginMap map[string]*plugin.Plugin

func (p PluginMap) ToPluginLimiterMap() PluginLimiterMap {
	var limiterPluginMap = make(PluginLimiterMap)
	for pluginInstance, p := range p {
		if len(p.Limiters) > 0 {
			limiterPluginMap[pluginInstance] = NewLimiterMap(p.Limiters)
		}
	}
	return limiterPluginMap
}
