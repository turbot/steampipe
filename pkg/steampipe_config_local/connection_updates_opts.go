package steampipe_config_local

type connectionUpdatesConfig struct {
	ForceUpdateConnectionNames []string
}

type ConnectionUpdatesOption func(opt *connectionUpdatesConfig)

func WithForceUpdate(connections []string) ConnectionUpdatesOption {
	return func(opt *connectionUpdatesConfig) {
		opt.ForceUpdateConnectionNames = connections
	}
}
