package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func DecodePlugin(block *hcl.Block) (*modconfig.Plugin, hcl.Diagnostics) {
	var plugin = &modconfig.Plugin{
		Source: block.Labels[0],
	}
	diags := gohcl.DecodeBody(block.Body, nil, plugin)
	if !diags.HasErrors() {
		plugin.OnDecoded(block)
	}

	return plugin, diags
}
