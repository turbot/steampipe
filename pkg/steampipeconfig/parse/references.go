package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe/pkg/steampipeconfig/hclhelpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// AddReferences populates the 'References' resource field, used for the introspection tables
func AddReferences(resource modconfig.HclResource, block *hcl.Block, parseCtx *ModParseContext) hcl.Diagnostics {
	resourceWithMetadata, ok := resource.(modconfig.ResourceWithMetadata)
	if !ok {
		return nil
	}

	var diags hcl.Diagnostics
	for _, attr := range block.Body.(*hclsyntax.Body).Attributes {
		for _, v := range attr.Expr.Variables() {
			for _, referenceBlockType := range modconfig.ReferenceBlocks {
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
					moreDiags := addResourceMetadata(reference, attr.SrcRange, parseCtx)
					diags = append(diags, moreDiags...)
					resourceWithMetadata.AddReference(reference)
					break
				}
			}
		}
	}
	return diags
}
