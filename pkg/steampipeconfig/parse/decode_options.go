package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
)

// DecodeOptions decodes an options block
func DecodeOptions(block *hcl.Block) (options.Options, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var dest options.Options
	switch block.Labels[0] {
	case options.ConnectionBlock:
		dest = &options.Connection{}
	case options.DatabaseBlock:
		dest = &options.Database{}
	case options.TerminalBlock:
		dest = &options.Terminal{}
	case options.GeneralBlock:
		dest = &options.General{}
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Invalid options type '%s'", block.Labels[0]),
			Subject:  &block.DefRange,
		})
		return nil, diags
	}

	diags = gohcl.DecodeBody(block.Body, nil, dest)
	if diags.HasErrors() {
		return nil, diags
	}

	return dest, nil
}
