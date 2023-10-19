package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/hcl_helpers"
	"github.com/turbot/pipe-fittings/ociinstaller"
	"golang.org/x/exp/maps"
	"strings"
)

type Plugin struct {
	Instance        string         `hcl:"name,label" db:"plugin_instance"`
	Alias           string         `hcl:"source,optional"`
	MemoryMaxMb     *int           `hcl:"memory_max_mb,optional" db:"memory_max_mb"`
	Limiters        []*RateLimiter `hcl:"limiter,block" db:"limiters"`
	FileName        *string        `db:"file_name"`
	StartLineNumber *int           `db:"start_line_number"`
	EndLineNumber   *int           `db:"end_line_number"`
	// the image ref as a string
	Plugin string `db:"plugin"`
}

// NewImplicitPlugin creates a default plugin config struct for a connection
// this is called when there is no explicit plugin config defined
// for a plugin which is used by a connection
func NewImplicitPlugin(connection *Connection, imageRef string) *Plugin {
	return &Plugin{
		// NOTE: set instance to image ref
		Instance: imageRef,
		Alias:    connection.PluginAlias,
		Plugin:   imageRef,
	}
}

func (l *Plugin) OnDecoded(block *hcl.Block) {
	pluginRange := hcl_helpers.BlockRange(block)
	l.FileName = &pluginRange.Filename
	l.StartLineNumber = &pluginRange.Start.Line
	l.EndLineNumber = &pluginRange.End.Line
	l.Plugin = ResolvePluginImageRef(l.Alias)
}

// IsDefault returns whether this config was created as a default
// i.e. a connection reference this plugin but there was no plugin config
// in this case the Instance will be the imageRef
func (l *Plugin) IsDefault() bool {
	return l.Instance == l.Plugin
}

func (l *Plugin) FriendlyName() string {
	return ociinstaller.NewSteampipeImageRef(l.Plugin).GetFriendlyName()
}

func (l *Plugin) GetMaxMemoryBytes() int64 {
	memoryMaxMb := 0
	if l.MemoryMaxMb != nil {
		memoryMaxMb = *l.MemoryMaxMb
	}
	return int64(1024 * 1024 * memoryMaxMb)
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

// ResolvePluginImageRef resolves the plugin image ref from the plugin alias
// (this handles the special case of locally developed plugins in the plugins/local folder)
func ResolvePluginImageRef(pluginAlias string) string {
	var imageRef string
	if strings.HasPrefix(pluginAlias, `local/`) {
		imageRef = pluginAlias
	} else {
		// ok so there is no plugin block reference - build the plugin image ref from the PluginAlias field
		imageRef = ociinstaller.NewSteampipeImageRef(pluginAlias).DisplayImageRef()
	}
	//  are there any instances for this plugin
	return imageRef
}
