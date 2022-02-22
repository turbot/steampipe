package modconfig

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
)

func GetAnonymousResourceShortName(block *hcl.Block, parent ModTreeItem) string {
	var shortName string

	anonymous := len(block.Labels) == 0
	if anonymous {
		shortName = GetUniqueName(block.Type, parent)
	} else {
		shortName = block.Labels[0]
	}
	return shortName
}

// GetUniqueName returns a name unique within the scope of this execution tree
func GetUniqueName(blockType string, parent ModTreeItem) string {

	// count how many children of this block type the parent has
	childCount := 0

	for _, child := range parent.GetChildren() {
		parsedName, err := ParseResourceName(child.Name())
		if err != nil {
			// we do not expect this
			continue
		}
		if parsedName.ItemType == blockType {
			childCount++
		}
	}
	sanitisedParentName := strings.Replace(parent.GetUnqualifiedName(), ".", "_", -1)
	return fmt.Sprintf("%s_anonymous_%s_%d", sanitisedParentName, blockType, childCount)
}
