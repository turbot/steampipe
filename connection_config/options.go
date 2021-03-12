package connection_config

// hcl options types
const (
	HclOptionsConnection = "connection"
	HclOptionsDatabase   = "database"
	HclOptionsGeneral    = "general"
	HclOptionsConsole    = "console"
)

type Options interface {
	IsOptions()
}

// ConnectionOptions
type ConnectionOptions struct {
	Cache    *bool `hcl:"cache"`
	CacheTTL *int  `hcl:"cache_ttl"`
	MaxFiles int   `hcl:"rlimit_files"`
}

func (f ConnectionOptions) IsOptions() {}

// ConsoleOptions
type ConsoleOptions struct {
	Header    bool   `hcl:"header"`
	Multi     bool   `hcl:"multi"`
	Output    string `hcl:"output"`
	Separator string `hcl:"separator"`
	Timing    string `hcl:"timing"`
}

func (f ConsoleOptions) IsOptions() {}

// GeneralOptions
type GeneralOptions struct {
	LogLevel    string `hcl:"log_level"`
	UpdateCheck bool   `hcl:"update_check"`
}

func (f GeneralOptions) IsOptions() {}

// DatabaseOptions
type DatabaseOptions struct {
	Port   int    `hcl:"port"`
	Listen string `hcl:"listen"`
}

func (f DatabaseOptions) IsOptions() {}
