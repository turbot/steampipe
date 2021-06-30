package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

func ParseConnectionGroup(block *hcl.Block, fileData map[string][]byte) (*modconfig.ConnectionGroup, hcl.Diagnostics) {
	connectionGroup := &modconfig.ConnectionGroup{}
	diags := gohcl.DecodeBody(block.Body, nil, connectionGroup)
	if diags.HasErrors() {
		return nil, diags
	}
	connectionGroup.Name = block.Labels[0]
	return connectionGroup, diags
}
