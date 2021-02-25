package connection_config

type FdwOptions struct {
	Cache    *bool
	CacheTTL *int
}

func (p FdwOptions) equals(other *FdwOptions) bool {
	//todo
	return false
}

type PluginOptions struct {
	RLimitFiles int
}

type ConsoleOptions struct {
	MultiLine bool
}
