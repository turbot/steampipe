package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

func GetAnonymousResourceShortName(block *hcl.Block, mod *Mod) string {
	var shortName string

	anonymous := len(block.Labels) == 0
	if anonymous {
		shortName = mod.GetUniqueName(fmt.Sprintf("anonymous_%s", block.Type))
	} else {
		shortName = block.Labels[0]
	}
	return shortName
}
