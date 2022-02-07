package modconfig

import "github.com/hashicorp/hcl/v2"

func GetAnonymousResourceShortName(block *hcl.Block, parent HclResource) string {
	var shortName string
	anonymous := len(block.Labels) == 0
	if anonymous {
		// if this resource is anonymous, the parent must be a ReportContainer
		reportContrainerParent, ok := parent.(*ReportContainer)
		if !ok {
			// shoul never happen
			panic("parent of an anonymous resource must be a ReportContainer")
		}
		shortName = reportContrainerParent.GetAnonymousChildName(block.Type)
	} else {
		shortName = block.Labels[0]
	}
	return shortName
}
