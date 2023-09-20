package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/ociinstaller"
	"github.com/turbot/steampipe/pkg/steampipeconfig/hclhelpers"
	"golang.org/x/exp/maps"
)

type Plugin struct {
	Instance        string         `hcl:"name,label" db:"plugin_instance"`
	Alias           string         `hcl:"source,optional"`
	MaxMemoryMb     *int           `hcl:"memory_max_mb,optional" db:"memory_max_mb"`
	Limiters        []*RateLimiter `hcl:"limiter,block" db:"rate_limiters"`
	FileName        *string        `db:"file_name"`
	StartLineNumber *int           `db:"start_line_number"`
	EndLineNumber   *int           `db:"end_line_number"`
	// the image ref as a string
	Plugin   string `db:"plugin"`
	imageRef *ociinstaller.SteampipeImageRef
}

// NewImplicitPlugin creates a default plugin config struct for a connection
// this is called when there is no explicit plugin config defined
// for a plugin which is used by a connection
func NewImplicitPlugin(connection *Connection, imageRef *ociinstaller.SteampipeImageRef) *Plugin {
	return &Plugin{
		// NOTE: set label to image ref
		Instance: imageRef.DisplayImageRef(),
		Alias:    connection.PluginAlias,
		Plugin:   imageRef.DisplayImageRef(),
		imageRef: imageRef,
	}
}

func (l *Plugin) OnDecoded(block *hcl.Block) {
	pluginRange := hclhelpers.BlockRange(block)
	l.FileName = &pluginRange.Filename
	l.StartLineNumber = &pluginRange.Start.Line
	l.EndLineNumber = &pluginRange.End.Line
	l.imageRef = ociinstaller.NewSteampipeImageRef(l.Alias)
	l.Plugin = l.imageRef.DisplayImageRef()
}

// IsDefault returns whether this config was created as a default
// i.e. a connection reference this plugin but there was no plugin config
// in this case the Instance will be the imageRef
func (l *Plugin) IsDefault() bool {
	return l.Instance == l.GetImageRef()
}

func (l *Plugin) GetMaxMemoryBytes() int64 {
	maxMemoryMb := 0
	if l.MaxMemoryMb != nil {
		maxMemoryMb = *l.MaxMemoryMb
	}
	return int64(1024 * 1024 * maxMemoryMb)
}

func (l *Plugin) GetImageRef() string {
	return l.imageRef.DisplayImageRef()
}

func (l *Plugin) GetLimiterMap() map[string]*RateLimiter {
	res := make(map[string]*RateLimiter, len(l.Limiters))
	for _, l := range l.Limiters {
		res[l.Name] = l
	}
	return res
}

func (l *Plugin) Equals(other *Plugin) bool {

	return l.Instance == other.Instance &&
		l.Alias == other.Alias &&
		l.GetMaxMemoryBytes() == other.GetMaxMemoryBytes() &&
		l.Plugin == other.Plugin &&
		// compare limiters ignoring order
		maps.EqualFunc(l.GetLimiterMap(), other.GetLimiterMap(), func(l, r *RateLimiter) bool { return l.Equals(r) })

}
