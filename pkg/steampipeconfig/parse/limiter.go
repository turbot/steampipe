package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/pipe-fittings/plugin"
)

func DecodeLimiter(block *hcl.Block) (*plugin.RateLimiter, hcl.Diagnostics) {
	var limiter = &plugin.RateLimiter{
		// populate name from label
		Name: block.Labels[0],
	}
	diags := gohcl.DecodeBody(block.Body, nil, limiter)
	if !diags.HasErrors() {
		limiter.OnDecoded(block)
	}

	return limiter, diags
}
