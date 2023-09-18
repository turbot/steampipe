package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe/pkg/ociinstaller"
)

type Plugin struct {
	Instance        string         `hcl:"name,label" db:"plugin_instance"`
	Source          string         `hcl:"source,optional"`
	MaxMemoryMb     *int           `hcl:"max_memory_mb,optional" db:"max_memory_mb"`
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
func NewImplicitPlugin(connection *Connection) *Plugin {
	imageRef := ociinstaller.NewSteampipeImageRef(connection.PluginAlias)
	return &Plugin{
		// NOTE: set label to image ref
		Instance: imageRef.DisplayImageRef(),
		Source:   connection.PluginAlias,
		Plugin:   imageRef.DisplayImageRef(),
		imageRef: imageRef,
	}
}

func (l *Plugin) OnDecoded(block *hcl.Block) {
	l.FileName = &block.DefRange.Filename
	l.StartLineNumber = &block.Body.(*hclsyntax.Body).SrcRange.Start.Line
	l.EndLineNumber = &block.Body.(*hclsyntax.Body).SrcRange.End.Line
	l.imageRef = ociinstaller.NewSteampipeImageRef(l.Source)
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
