package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

func ParseControl(block *hcl.Block) (*modconfig.Control, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var c = &modconfig.Control{}

	diags = gohcl.DecodeBody(block.Body, nil, c)
	if diags.HasErrors() {
		return nil, diags
	}

	c.ShortName = &block.Labels[0]
	return c, nil
}
