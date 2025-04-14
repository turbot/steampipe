package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	pparse "github.com/turbot/pipe-fittings/v2/parse"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/pipe-fittings/v2/schema"
)

func DecodePlugin(block *hcl.Block) (*plugin.Plugin, hcl.Diagnostics) {
	// manually decode child limiter blocks
	content, rest, diags := block.Body.PartialContent(pparse.PluginBlockSchema)
	if diags.HasErrors() {
		return nil, diags
	}
	body := rest.(*hclsyntax.Body)

	// decode attributes using 'rest' (these are automativally parsed so are not in schema)
	var p = &plugin.Plugin{
		// default source and name to label
		Instance: block.Labels[0],
		Alias:    block.Labels[0],
	}
	moreDiags := gohcl.DecodeBody(body, &hcl.EvalContext{}, p)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, diags
	}

	// decode limiter blocks using 'content'
	for _, block := range content.Blocks {
		switch block.Type {
		// only block defined in schema
		case schema.BlockTypeRateLimiter:
			limiter, moreDiags := pparse.DecodeLimiter(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			limiter.SetPlugin(p)
			p.Limiters = append(p.Limiters, limiter)
		}
	}
	if !diags.HasErrors() {
		p.OnDecoded(block)
	}

	return p, diags
}
