package connection

import (
	"github.com/turbot/pipe-fittings/plugin"
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
	for pluginInstance, p := range p {
		if len(p.Limiters) > 0 {
			limiterPluginMap[pluginInstance] = NewLimiterMap(p.Limiters)
		}
	}
	return limiterPluginMap
}

//func (p PluginMap) Diff(otherMap PluginMap) (added, deleted, changed map[string][]*modconfig.Plugin) {
//	// results are maps of connections keyed by plugin instance
//	added = make(map[string][]*modconfig.Plugin)
//	deleted = make(map[string][]*modconfig.Plugin)
//	changed = make(map[string][]*modconfig.Plugin)
//
//	for name, plugin := range p {
//		if otherConnection, ok := otherMap[name]; !ok {
//			deleted[plugin.Instance] = append(deleted[plugin.Instance], plugin)
//		} else {
//			// check for changes
//
//			// special case - if the plugin has changed, treat this as a deletion and a re-add
//			if plugin.Instance != otherConnection.Plugin {
//				added[otherConnection.Plugin] = append(added[otherConnection.Plugin], otherConnection)
//				deleted[plugin.Instance] = append(deleted[plugin.Instance], plugin)
//			} else {
//				if !plugin.Equals(otherConnection) {
//					changed[plugin.Instance] = append(changed[plugin.Instance], otherConnection)
//				}
//			}
//		}
//	}
//
//	for otherName, otherConnection := range otherMap {
//		if _, ok := p[otherName]; !ok {
//			added[otherConnection.Plugin] = append(added[otherConnection.Plugin], otherConnection)
//		}
//	}
//
//	return
//}
