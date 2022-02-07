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
		childIndex := len(parent.GetChildren())
		parent_segment := strings.Replace(parent.GetUnqualifiedName(), ".", "_", -1)
		shortName = fmt.Sprintf("%s_child%d", parent_segment, childIndex)
	} else {
		shortName = block.Labels[0]
	}
	return shortName
}
