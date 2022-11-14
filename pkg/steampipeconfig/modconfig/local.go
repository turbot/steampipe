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

	Value    cty.Value
	Mod      *Mod `cty:"mod"`
	metadata *ResourceMetadata
	Paths    []NodePath `column:"path,jsonb"`
	parents  []ModTreeItem
}

func NewLocal(name string, val cty.Value, declRange hcl.Range, mod *Mod) *Local {
	l := &Local{
		Value: val,
		Mod:   mod,
		HclResourceBase: HclResourceBase{
			ShortName:       name,
			UnqualifiedName: fmt.Sprintf("local.%s", name),
			FullName:        fmt.Sprintf("%s.local.%s", mod.ShortName, name),
			DeclRange:       declRange,
			blockType:       BlockTypeLocals,
		},
	}
	return l
}

// GetMod implements ModTreeItem
func (l *Local) GetMod() *Mod {
	return l.Mod
}

// AddParent implements ModTreeItem
func (l *Local) AddParent(parent ModTreeItem) error {
	l.parents = append(l.parents, parent)

	return nil
}

// GetParents implements ModTreeItem
func (l *Local) GetParents() []ModTreeItem {
	return l.parents
}

// GetChildren implements ModTreeItem
func (l *Local) GetChildren() []ModTreeItem {
	return nil
}

// GetPaths implements ModTreeItem
func (l *Local) GetPaths() []NodePath {
	// lazy load
	if len(l.Paths) == 0 {
		l.SetPaths()
	}
	return l.Paths
}

// SetPaths implements ModTreeItem
func (l *Local) SetPaths() {
	for _, parent := range l.parents {
		for _, parentPath := range parent.GetPaths() {
			l.Paths = append(l.Paths, append(parentPath, l.Name()))
		}
	}
}

// GetDocumentation implement ModTreeItem
func (*Local) GetDocumentation() string {
	return ""
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
