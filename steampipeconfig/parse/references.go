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
	modconfig.BlockTypeCard,
	modconfig.BlockTypeChart,
	modconfig.BlockTypeHierarchy,
	modconfig.BlockTypeImage,
	modconfig.BlockTypeInput,
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
					var blockName string
					if len(block.Labels) > 0 {
						blockName = block.Labels[0]
					}
					reference := &modconfig.ResourceReference{
						To:        referenceString,
						From:      resource.GetUnqualifiedName(),
						BlockType: block.Type,
						BlockName: blockName,
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
