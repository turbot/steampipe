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
	modconfig.BlockTypeParam,
	"local"}

// AddReferences populates the 'References' resource field, used for the introspection tables
func AddReferences(resource modconfig.HclResource, block *hcl.Block) {
	for _, attr := range block.Body.(*hclsyntax.Body).Attributes {
		for _, v := range attr.Expr.Variables() {
			for _, referenceBlockType := range referenceBlockTypes {
				if referenceString, ok := hclhelpers.ResourceNameFromTraversal(referenceBlockType, v); ok {
					resource.AddReference(modconfig.ResourceReference{
						Name:      referenceString,
						BlockType: block.Type,
						BlockName: block.Labels[0],
						Attribute: attr.Name,
					})
					break
				}
			}
		}
	}
}
