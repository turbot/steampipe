package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe/pkg/ociinstaller"
)

type Plugin struct {
	Label           string         `hcl:"name,label"`
	Source          string         `hcl:"source,optional"`
	MaxMemoryMb     int            `hcl:"max_memory_mb,optional"`
	Limiters        []*RateLimiter `hcl:"limiter,block"`
	FileName        *string
	StartLineNumber *int
	EndLineNumber   *int
	imageRef        *ociinstaller.SteampipeImageRef
}

// NewDefaultPlugin creates a default plugin config struct for a connection
// this is called when there is no explicit plugin config defined
// for a plugin which is used by a connection
func NewDefaultPlugin(connection *Connection) *Plugin {
	imageRef := ociinstaller.NewSteampipeImageRef(connection.PluginAlias)
	return &Plugin{
		// NOTE: set label to image ref
		Label:    imageRef.DisplayImageRef(),
		Source:   connection.PluginAlias,
		imageRef: imageRef,
	}
}

func (l *Plugin) OnDecoded(block *hcl.Block) {
	l.FileName = &block.DefRange.Filename
	l.StartLineNumber = &block.Body.(*hclsyntax.Body).SrcRange.Start.Line
	l.EndLineNumber = &block.Body.(*hclsyntax.Body).SrcRange.End.Line
	l.imageRef = ociinstaller.NewSteampipeImageRef(l.Source)
}

// IsDefault returns whether this config was created as a default
// i.e. a connection reference this plugin but there was no plugin config
// in this case the Label will be the ImageRef
func (l *Plugin) IsDefault() bool {
	return l.Label == l.GetImageRef()
}

func (l *Plugin) GetMaxMemoryBytes() int64 {
	return int64(1024 * 1024 * l.MaxMemoryMb)
}

func (l *Plugin) GetImageRef() string {
	return l.imageRef.DisplayImageRef()
}
