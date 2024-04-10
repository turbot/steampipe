package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
	"github.com/zclconf/go-cty/cty"
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
			Subject:  hclhelpers.BlockRangePointer(block),
		})
		return nil, diags
	}
	if len(block.Labels) > 0 && block.Labels[0] == options.QueryBlock {
		handleQueryTiming(block, destination)
	}
	diags = gohcl.DecodeBody(block.Body, nil, destination)
	if diags.HasErrors() {
		return nil, diags
	}

	return destination, nil
}

// for Query options block,  if timing attribute is set to "verbose", replace with true and set verbose to true
func handleQueryTiming(block *hcl.Block, destination options.Options) {
	body := block.Body.(*hclsyntax.Body)
	for _, attr := range body.Attributes {
		// if timing attribute is set to "verbose", replace with true and set verbose to true
		if attr.Name == "timing" {
			if scopeTraversal, ok := attr.Expr.(*hclsyntax.ScopeTraversalExpr); ok {
				if len(scopeTraversal.Traversal) == 1 && scopeTraversal.Traversal[0].(hcl.TraverseRoot).Name == "verbose" {
					attr.Expr = &hclsyntax.LiteralValueExpr{
						Val:      cty.BoolVal(true),
						SrcRange: attr.Expr.Range(),
					}
					verbose := true
					destination.(*options.Query).VerboseTiming = &verbose
				}
			}
		}
	}
}

type OptionsBlockMapping = map[string]options.Options

func defaultOptionsBlockMapping() OptionsBlockMapping {
	mapping := OptionsBlockMapping{
		options.DatabaseBlock:  &options.Database{},
		options.GeneralBlock:   &options.General{},
		options.QueryBlock:     &options.Query{},
		options.CheckBlock:     &options.Check{},
		options.DashboardBlock: &options.GlobalDashboard{},
		options.PluginBlock:    &options.Plugin{},
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
