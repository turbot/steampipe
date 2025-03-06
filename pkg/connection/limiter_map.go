package connection

import (
	"github.com/turbot/pipe-fittings/v2/plugin"
	"golang.org/x/exp/maps"
)

// LimiterMap is a map of limiter name to limiter definition
type LimiterMap map[string]*plugin.RateLimiter

func NewLimiterMap(limiters []*plugin.RateLimiter) LimiterMap {
	res := make(LimiterMap)
	for _, l := range limiters {
		res[l.Name] = l
	}
	return res
}
func (l LimiterMap) Equals(other LimiterMap) bool {
	return maps.EqualFunc(l, other, func(l1, l2 *plugin.RateLimiter) bool { return l1.Equals(l2) })
}

// ToPluginLimiterMap converts limiter map keyed by limiter name to a map of limiter maps keyed by plugin image ref
func (l LimiterMap) ToPluginLimiterMap() PluginLimiterMap {
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
