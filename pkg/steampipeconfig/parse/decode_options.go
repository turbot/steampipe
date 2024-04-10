package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
)

// DecodeOptions decodes an options block
func DecodeOptions(block *hcl.Block, overrides ...BlockMappingOverride) (options.Options, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	mapping := defaultOptionsBlockMapping()
	for _, applyOverride := range overrides {
		applyOverride(mapping)
	}
	optionsType := block.Labels[0]

	destination, ok := mapping[optionsType]
	if !ok {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Unexpected options type '%s'", block.Labels[0]),
			Subject:  hclhelpers.BlockRangePointer(block),
		})
		return nil, diags
	}

	if timingOptions, ok := destination.(options.CanSetTiming); ok {
		morediags := decodeTimingFlag(block, timingOptions)
		if morediags.HasErrors() {
			diags = append(diags, morediags...)
			return nil, diags
		}
	}
	diags = gohcl.DecodeBody(block.Body, nil, destination)
	if diags.HasErrors() {
		return nil, diags
	}

	return destination, nil
}

// for Query options block,  if timing attribute is set to "verbose", replace with true and set verbose to true
func decodeTimingFlag(block *hcl.Block, timingOptions options.CanSetTiming) hcl.Diagnostics {
	body := block.Body.(*hclsyntax.Body)
	timingAttribute := body.Attributes["timing"]
	if timingAttribute == nil {
		return nil
	}
	// remove the attribute so subsequent decoding does not see it
	delete(body.Attributes, "timing")

	scopeTraversal, ok := timingAttribute.Expr.(*hclsyntax.ScopeTraversalExpr)
	if !ok || len(scopeTraversal.Traversal) != 1 {
		return nil
	}
	value := scopeTraversal.Traversal[0].(hcl.TraverseRoot).Name
	return timingOptions.SetTiming(value, timingAttribute.Range())

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
