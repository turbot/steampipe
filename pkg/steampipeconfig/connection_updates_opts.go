package steampipeconfig

type connectionUpdatesSettings struct {
	ForceUpdateConnectionNames []string
	// if we need to fetch rate limiter defs for all plugins, this will be populated
	// map of plugin to exemplar connection
	FetchRateLimitersForAllPlugins bool
	PluginExemplarConnections      map[string]string
}

type ConnectionUpdatesOption func(opt *connectionUpdatesSettings)

func WithForceUpdate(connections []string) ConnectionUpdatesOption {
	return func(opt *connectionUpdatesSettings) {
		opt.ForceUpdateConnectionNames = connections
	}
}
func WithFetchRateLimiterDefs(PluginExemplarConnections map[string]string) ConnectionUpdatesOption {
	return func(opt *connectionUpdatesSettings) {
		opt.FetchRateLimitersForAllPlugins = true
		opt.PluginExemplarConnections = PluginExemplarConnections
	}
}
