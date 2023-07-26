package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func DecodeLimiter(block *hcl.Block) (*modconfig.RateLimiter, hcl.Diagnostics) {
	var limiter = &modconfig.RateLimiter{
		Name: block.Labels[0],
	}
	diags := gohcl.DecodeBody(block.Body, nil, limiter)
	return limiter, diags
}
