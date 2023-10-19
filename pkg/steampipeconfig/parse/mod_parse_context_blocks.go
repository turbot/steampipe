package parse

import (
	"fmt"
	"github.com/turbot/go-kit/hcl_helpers"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
)

func (m *ModParseContext) DetermineBlockName(block *hcl.Block) string {
	var shortName string

	// have we cached a name for this block (i.e. is this the second decode pass)
	if name, ok := m.GetCachedBlockShortName(block); ok {
		return name
	}

	// if there is a parent set in the parent stack, this block is a child of that parent
	parentName := m.PeekParent()

	anonymous := len(block.Labels) == 0
	if anonymous {
		shortName = m.getUniqueName(block.Type, parentName)
	} else {
		shortName = block.Labels[0]
	}
	// build unqualified name
	unqualifiedName := fmt.Sprintf("%s.%s", block.Type, shortName)
	m.addChildBlockForParent(parentName, unqualifiedName)
	// cache this name for the second decode pass
	m.cacheBlockName(block, unqualifiedName)
	return shortName
}

func (m *ModParseContext) GetCachedBlockName(block *hcl.Block) (string, bool) {
	name, ok := m.blockNameMap[m.blockHash(block)]
	return name, ok
}

func (m *ModParseContext) GetCachedBlockShortName(block *hcl.Block) (string, bool) {
	unqualifiedName, ok := m.blockNameMap[m.blockHash(block)]
	if ok {
		parsedName, err := modconfig.ParseResourceName(unqualifiedName)
		if err != nil {
			return "", false
		}
		return parsedName.Name, true
	}
	return "", false
}

func (m *ModParseContext) GetDecodedResourceForBlock(block *hcl.Block) (modconfig.HclResource, bool) {
	if name, ok := m.GetCachedBlockName(block); ok {
		// see whether the mod contains this resource already
		parsedName, err := modconfig.ParseResourceName(name)
		if err == nil {
			return m.CurrentMod.GetResource(parsedName)
		}
	}
	return nil, false
}

func (m *ModParseContext) cacheBlockName(block *hcl.Block, shortName string) {
	m.blockNameMap[m.blockHash(block)] = shortName
}

func (m *ModParseContext) blockHash(block *hcl.Block) string {
	return helpers.GetMD5Hash(hcl_helpers.BlockRange(block).String())
}

// getUniqueName returns a name unique within the scope of this execution tree
func (m *ModParseContext) getUniqueName(blockType string, parent string) string {
	// count how many children of this block type the parent has
	childCount := 0

	for _, childName := range m.blockChildMap[parent] {
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

func (m *ModParseContext) addChildBlockForParent(parent, child string) {
	m.blockChildMap[parent] = append(m.blockChildMap[parent], child)
}
