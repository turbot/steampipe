package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Plugin struct {
	Source          string         `hcl:"source,optional"`
	MaxMemoryMb     int            `hcl:"max_memory_mb,optional"`
	Limiters        []*RateLimiter `hcl:"limiter,block"`
	FileName        *string
	StartLineNumber *int
	EndLineNumber   *int
}

func (l *Plugin) OnDecoded(block *hcl.Block) {
	l.FileName = &block.DefRange.Filename
	l.StartLineNumber = &block.Body.(*hclsyntax.Body).SrcRange.Start.Line
	l.EndLineNumber = &block.Body.(*hclsyntax.Body).SrcRange.End.Line
}

func (l *Plugin) GetMaxMemoryBytes() int64 {
	return int64(1024 * 1024 * l.MaxMemoryMb)
}
