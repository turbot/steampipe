package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

func ParseLocals(block *hcl.Block) (*modconfig.Query, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var q = &modconfig.Query{}

	diags = gohcl.DecodeBody(block.Body, nil, q)
	if diags.HasErrors() {
		return nil, diags
	}

	q.ShortName = &block.Labels[0]
	return q, nil
}
