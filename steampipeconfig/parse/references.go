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
func AddReferences(resource modconfig.HclResource, block *hcl.Block, runCtx *RunContext) {
	for _, attr := range block.Body.(*hclsyntax.Body).Attributes {
		for _, v := range attr.Expr.Variables() {
			for _, blockType := range referenceBlockTypes {
				if referenceString, ok := hclhelpers.ResourceNameFromTraversal(blockType, v); ok {
					// find this resource in the current mod
					// build a potential resource reference
					ref := modconfig.ResourceReference{
						Name:   referenceString,
						Parent: runCtx.Mod.FullName,
					}
					refResource, ok := runCtx.Mod.AllResources[ref]
					if !ok {
						break
					}

					// TODO consider refs in another mod
					//parsedName, err := modconfig.ParseResourceName(reference)
					//if parsedName.Mod != "" && parsedName.Mod != runCtx.Mod.ShortName{
					//	break
					//}

					resource.AddReference(modconfig.NewResourceReference(refResource))
					break
				}
			}
		}
	}
}
