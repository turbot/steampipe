package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func DecodeLimiter(block *hcl.Block) (*modconfig.RateLimiter, hcl.Diagnostics) {
	var limiter = &modconfig.RateLimiter{
		Name: block.Labels[0],
	}
	diags := gohcl.DecodeBody(block.Body, nil, limiter)
	limiter.FileName = &block.DefRange.Filename
	limiter.StartLineNumber = &block.Body.(*hclsyntax.Body).SrcRange.Start.Line
	limiter.EndLineNumber = &block.Body.(*hclsyntax.Body).SrcRange.End.Line
	limiter.Status = modconfig.LimiterStatusActive
	limiter.Source = modconfig.LimiterSourceConfig

	return limiter, diags
}
