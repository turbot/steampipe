package connection

import "github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"

type LimiterMap map[string]*modconfig.RateLimiter

// GetPluginsWithChangedLimiters returns a list of plugins (short name)
// who have changed limiter configs (added/deleted/update)
func (l LimiterMap) GetPluginsWithChangedLimiters(other LimiterMap) map[string]struct{} {
	var pluginsWithChangedLimiters = make(map[string]struct{})

	for name, limiter := range l {
		otherLimiter, ok := other[name]
		if !ok || !limiter.Equals(otherLimiter) {
			pluginsWithChangedLimiters[limiter.Plugin] = struct{}{}
		}
	}
	for name, otherLimiter := range other {
		if _, ok := l[name]; !ok {
			pluginsWithChangedLimiters[otherLimiter.Plugin] = struct{}{}
		}
	}
	return pluginsWithChangedLimiters
}
