package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe/steampipeconfig/hclhelpers"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// list of block types we store references for
var referenceBlockTypes = []string{
	string(modconfig.BlockTypeMod),
	modconfig.BlockTypeQuery,
	modconfig.BlockTypeControl,
	modconfig.BlockTypeBenchmark,
	modconfig.BlockTypeReport,
	modconfig.BlockTypeContainer,
	modconfig.BlockTypeChart,
	modconfig.BlockTypeCounter,
	modconfig.BlockTypeHierarchy,
	modconfig.BlockTypeImage,
	modconfig.BlockTypeTable,
	modconfig.BlockTypeText,
	modconfig.BlockTypeParam,
	"local"}

// AddReferences populates the 'References' resource field, used for the introspection tables
func AddReferences(resource modconfig.HclResource, block *hcl.Block, runCtx *RunContext) hcl.Diagnostics {
	// NOTE: exclude locals
	if block.Type == modconfig.BlockTypeLocals {
		return nil
	}

	var diags hcl.Diagnostics
	for _, attr := range block.Body.(*hclsyntax.Body).Attributes {
		for _, v := range attr.Expr.Variables() {
			for _, referenceBlockType := range referenceBlockTypes {
				if referenceString, ok := hclhelpers.ResourceNameFromTraversal(referenceBlockType, v); ok {
					reference := &modconfig.ResourceReference{
						To:        referenceString,
						From:      resource.Name(),
						BlockType: block.Type,
						BlockName: block.Labels[0],
						Attribute: attr.Name,
					}
					moreDiags := addResourceMetadata(reference, attr.SrcRange, runCtx)
					diags = append(diags, moreDiags...)
					resource.AddReference(reference)
					break
				}
			}
		}
	}
	return diags
}
