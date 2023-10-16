package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/go-kit/hcl_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
)

// DecodeOptions decodes an options block
func DecodeOptions(block *hcl.Block, overrides ...BlockMappingOverride) (options.Options, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	mapping := defaultOptionsBlockMapping()
	for _, applyOverride := range overrides {
		applyOverride(mapping)
	}

	destination, ok := mapping[block.Labels[0]]
	if !ok {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Unexpected options type '%s'", block.Labels[0]),
			Subject:  hcl_helpers.BlockRangePointer(block),
		})
		return nil, diags
	}

	diags = gohcl.DecodeBody(block.Body, nil, destination)
	if diags.HasErrors() {
		return nil, diags
	}

	return destination, nil
}

type OptionsBlockMapping = map[string]options.Options

func defaultOptionsBlockMapping() OptionsBlockMapping {
	mapping := OptionsBlockMapping{
		options.ConnectionBlock: &options.Connection{},
		options.DatabaseBlock:   &options.Database{},
		options.TerminalBlock:   &options.Terminal{},
		options.GeneralBlock:    &options.General{},
		options.QueryBlock:      &options.Query{},
		options.CheckBlock:      &options.Check{},
		options.DashboardBlock:  &options.GlobalDashboard{},
		options.PluginBlock:     &options.Plugin{},
	}
	return mapping
}

type BlockMappingOverride func(OptionsBlockMapping)

// WithOverride overrides the default block mapping for a single block type
func WithOverride(blockName string, destination options.Options) BlockMappingOverride {
	return func(mapping OptionsBlockMapping) {
		mapping[blockName] = destination
	}
}
