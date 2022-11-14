package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

// Local is a struct representing a Local resource
type Local struct {
	HclResourceBase
	ModTreeItemBase

	Value    cty.Value
	metadata *ResourceMetadata
}

func NewLocal(name string, val cty.Value, declRange hcl.Range, mod *Mod) *Local {
	fullName := fmt.Sprintf("%s.local.%s", mod.ShortName, name)
	l := &Local{
		Value: val,
		HclResourceBase: HclResourceBase{
			ShortName:       name,
			UnqualifiedName: fmt.Sprintf("local.%s", name),
			FullName:        fullName,
			DeclRange:       declRange,
			blockType:       BlockTypeLocals,
		},
		ModTreeItemBase: ModTreeItemBase{
			Mod:      mod,
			fullName: fullName,
		},
	}
	return l
}

func (l *Local) Diff(other *Local) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: l,
		Name: l.Name(),
	}

	if !utils.SafeStringsEqual(l.FullName, other.FullName) {
		res.AddPropertyDiff("Name")
	}

	if !utils.SafeStringsEqual(l.Value, other.Value) {
		res.AddPropertyDiff("Value")
	}

	res.populateChildDiffs(l, other)
	return res
}
