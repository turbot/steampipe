package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

// Local is a struct representing a Local resource
type Local struct {
	ModTreeItemImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Value    cty.Value
	metadata *ResourceMetadata
}

func NewLocal(name string, val cty.Value, declRange hcl.Range, mod *Mod) *Local {
	fullName := fmt.Sprintf("%s.local.%s", mod.ShortName, name)
	l := &Local{
		Value: val,
		ModTreeItemImpl: ModTreeItemImpl{
			HclResourceImpl: HclResourceImpl{
				ShortName:       name,
				UnqualifiedName: fmt.Sprintf("local.%s", name),
				FullName:        fullName,
				DeclRange:       declRange,
				blockType:       BlockTypeLocals,
				// disable cty serialisation of base properties
				disableCtySerialise: true,
			},
			Mod: mod,
		},
	}
	return l
}

// CtyValue implements CtyValueProvider
func (l *Local) CtyValue() (cty.Value, error) {
	return l.Value, nil
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
