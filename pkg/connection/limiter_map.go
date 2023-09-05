package connection

import (
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"golang.org/x/exp/maps"
)

// map of plugin short name to Limiter map for the plugin
type PluginLimiterMap map[string]LimiterMap
func (l PluginLimiterMap) Equals(other PluginLimiterMap) bool {
	return maps.EqualFunc(l, other, func(m1, m2 LimiterMap) bool { return m1.Equals(m2) })
}



type LimiterMap map[string]*modconfig.RateLimiter

func (l LimiterMap) Equals(other LimiterMap) bool {
	return maps.EqualFunc(l, other, func(l1, l2 *modconfig.RateLimiter) bool { return l1.Equals(l2) })
}

// ToPluginMap converts limiter map keyed by limiter name to a map of limiter maps keyed by plugin
func (l LimiterMap) ToPluginMap() PluginLimiterMap{
	res := make(PluginLimiterMap)
	for name, limiter := range l {
		limitersForPlugin := res[limiter.Plugin]
		if limitersForPlugin == nil {
			limitersForPlugin = make(LimiterMap)
		}
		limitersForPlugin[name] = limiter
		res[limiter.Plugin] = limitersForPlugin
	}
	return res
}
