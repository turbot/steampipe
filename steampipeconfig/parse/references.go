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
	"local"}

func AddReferences(resource modconfig.HclResource, block *hcl.Block) {
	// populate the 'References' field
	for _, attr := range block.Body.(*hclsyntax.Body).Attributes {
		for _, v := range attr.Expr.Variables() {
			for _, blockType := range referenceBlockTypes {
				if reference, ok := hclhelpers.ResourceNameFromTraversal(blockType, v); ok {
					resource.AddReference(reference)
					break
				}
			}
		}
	}
}
