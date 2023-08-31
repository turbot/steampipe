package modconfig

import "github.com/hashicorp/hcl/v2"

type Plugin struct {
	Source      string        `hcl:"source"`
	MaxMemoryMb int           `hcl:"max_memory_mb"`
	Limiters    []RateLimiter `hcl:"limiter,block"`
}

func (p Plugin) OnDecoded(block *hcl.Block) {

}
