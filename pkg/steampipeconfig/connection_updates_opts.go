package steampipeconfig

type connectionUpdatesSettings struct {
	ForceUpdateConnectionNames  []string
	FetchRateLimitersForPlugins map[string]struct{}
}

func newConnectionUpdatesSettings() *connectionUpdatesSettings {
	return &connectionUpdatesSettings{
		FetchRateLimitersForPlugins: make(map[string]struct{}),
	}
}

type ConnectionUpdatesOption func(opt *connectionUpdatesSettings)

func WithForceUpdate(connections []string) ConnectionUpdatesOption {
	return func(opt *connectionUpdatesSettings) {
		opt.ForceUpdateConnectionNames = connections
	}
}
func WithFetchRateLimiterDefs(fetchRateLimitersForPlugins map[string]struct{}) ConnectionUpdatesOption {
	return func(opt *connectionUpdatesSettings) {
		opt.FetchRateLimitersForPlugins = fetchRateLimitersForPlugins
	}
}
