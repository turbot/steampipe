package steampipeconfig

type connectionUpdatesSettings struct {
	ForceUpdateConnectionNames          []string
	FetchRateLimiterDefsConnectionNames []string
}

type ConnectionUpdatesOption func(opt *connectionUpdatesSettings)

func WithForceUpdate(connections []string) ConnectionUpdatesOption {
	return func(opt *connectionUpdatesSettings) {
		opt.ForceUpdateConnectionNames = connections
	}
}
func WithFetchRateLimiterDefs(connections []string) ConnectionUpdatesOption {
	return func(opt *connectionUpdatesSettings) {
		opt.FetchRateLimiterDefsConnectionNames = connections
	}
}
