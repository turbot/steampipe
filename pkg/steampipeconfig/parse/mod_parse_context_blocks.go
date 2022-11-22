package parse

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func (r *ModParseContext) DetermineBlockName(block *hcl.Block) string {
	var shortName string

	// have we cached a name for this block (i.e. is this the second decode pass)
	if name, ok := r.GetCachedBlockShortName(block); ok {
		return name
	}

	// if there is a parent set in the parent stack, this block is a child of that parent
	parentName := r.PeekParent()

	anonymous := len(block.Labels) == 0
	if anonymous {
		shortName = r.getUniqueName(block.Type, parentName)
	} else {
		shortName = block.Labels[0]
	}
	// build unqualified name
	unqualifiedName := fmt.Sprintf("%s.%s", block.Type, shortName)
	r.addChildBlockForParent(parentName, unqualifiedName)
	// cache this name for the second decode pass
	r.cacheBlockName(block, unqualifiedName)
	return shortName
}

func (r *ModParseContext) GetCachedBlockName(block *hcl.Block) (string, bool) {
	name, ok := r.blockNameMap[r.blockHash(block)]
	return name, ok
}

func (r *ModParseContext) GetCachedBlockShortName(block *hcl.Block) (string, bool) {
	unqualifiedName, ok := r.blockNameMap[r.blockHash(block)]
	if ok {
		parsedName, err := modconfig.ParseResourceName(unqualifiedName)
		if err != nil {
			return "", false
		}
		return parsedName.Name, true
	}
	return "", false
}

func (r *ModParseContext) GetDecodedResourceForBlock(block *hcl.Block) (modconfig.HclResource, bool) {
	if name, ok := r.GetCachedBlockName(block); ok {
		// see whether the mod contains this resource already
		parsedName, err := modconfig.ParseResourceName(name)
		if err == nil {
			return modconfig.GetResource(r.CurrentMod, parsedName)
		}
	}
	return nil, false
}

func (r *ModParseContext) cacheBlockName(block *hcl.Block, shortName string) {
	r.blockNameMap[r.blockHash(block)] = shortName
}

func (r *ModParseContext) blockHash(block *hcl.Block) string {
	return helpers.GetMD5Hash(block.DefRange.String())
}

// getUniqueName returns a name unique within the scope of this execution tree
func (r *ModParseContext) getUniqueName(blockType string, parent string) string {
	// count how many children of this block type the parent has
	childCount := 0

	for _, childName := range r.blockChildMap[parent] {
		parsedName, err := modconfig.ParseResourceName(childName)
		if err != nil {
			// we do not expect this
			continue
		}
		if parsedName.ItemType == blockType {
			childCount++
		}
	}
	sanitisedParentName := strings.Replace(parent, ".", "_", -1)
	return fmt.Sprintf("%s_anonymous_%s_%d", sanitisedParentName, blockType, childCount)
}

func (r *ModParseContext) addChildBlockForParent(parent, child string) {
	r.blockChildMap[parent] = append(r.blockChildMap[parent], child)
}
