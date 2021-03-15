package connection_config

import "github.com/turbot/go-kit/types"

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
	// string containing a bool - supports true/false/off/on etc
	CacheBoolString *string `hcl:"cache"`
	CacheTTL        *int    `hcl:"cache_ttl"`
}

func (c ConnectionOptions) IsOptions() {}

func (c ConnectionOptions) equals(other *ConnectionOptions) bool {
	return c.Cache() == other.Cache() &&
		*c.CacheTTL == *other.CacheTTL
}

// convert CacheBoolString into a bool pointer
func (c ConnectionOptions) Cache() *bool {
	return types.ToBoolPtr(c.CacheBoolString)
}

// ConsoleOptions
type ConsoleOptions struct {
	Output    *string `hcl:"output"`
	Separator *string `hcl:"separator"`
	// strings containing a bool - supports true/false/off/on etc
	HeaderBoolString *string `hcl:"header"`
	MultiBoolString  *string `hcl:"multi"`
	TimingBoolString *string `hcl:"timing"`
}

// functions to convert strings representing bool values into bool pointers
func (c ConsoleOptions) Header() *bool {
	return types.ToBoolPtr(c.HeaderBoolString)
}

func (c ConsoleOptions) Multi() *bool {
	return types.ToBoolPtr(c.MultiBoolString)
}

func (c ConsoleOptions) Timing() *bool {
	return types.ToBoolPtr(c.MultiBoolString)
}

func (g ConsoleOptions) IsOptions() {}

// GeneralOptions
type GeneralOptions struct {
	LogLevel    *string `hcl:"log_level"`
	UpdateCheck *string `hcl:"update_check"`
}

func (g GeneralOptions) IsOptions() {}

// DatabaseOptions
type DatabaseOptions struct {
	Port   *int    `hcl:"port"`
	Listen *string `hcl:"listen"`
}

func (d DatabaseOptions) IsOptions() {}
