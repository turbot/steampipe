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

func PluginForConnection(connection *Connection) *Plugin {
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

func (l *Plugin) GetMaxMemoryBytes() int64 {
	return int64(1024 * 1024 * l.MaxMemoryMb)
}

func (l *Plugin) GetImageRef() string {
	return l.imageRef.DisplayImageRef()
}
