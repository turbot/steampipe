package modconfig

import (
	"github.com/zclconf/go-cty/cty"
)

type ModTreeItemBase struct {
	HclResourceBase

	Mod   *Mod       `cty:"mod" json:"-"`
	Paths []NodePath `column:"path,jsonb" json:"-"`

	parents  []ModTreeItem
	children []ModTreeItem
}

// AddParent implements ModTreeItem
func (b *ModTreeItemBase) AddParent(parent ModTreeItem) error {
	b.parents = append(b.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (b *ModTreeItemBase) GetParents() []ModTreeItem {
	return b.parents
}

// GetChildren implements ModTreeItem
func (b *ModTreeItemBase) GetChildren() []ModTreeItem {
	return b.children
}
func (b *ModTreeItemBase) GetPaths() []NodePath {
	// lazy load
	if len(b.Paths) == 0 {
		b.SetPaths()
	}
	return b.Paths
}

// SetPaths implements ModTreeItem
func (b *ModTreeItemBase) SetPaths() {
	for _, parent := range b.parents {
		for _, parentPath := range parent.GetPaths() {
			b.Paths = append(b.Paths, append(parentPath, b.FullName))
		}
	}
}
func (b *ModTreeItemBase) GetMod() *Mod {
	return b.Mod
}

func (b *ModTreeItemBase) IsTopLevelResource() bool {
	return len(b.parents) == 1 && b.parents[0] == b.Mod
}

// GetModTreeItemBase implements ModTreeItem
func (b *ModTreeItemBase) GetModTreeItemBase() *ModTreeItemBase {
	return b
}

// CtyValue implements CtyValueProvider
func (b *ModTreeItemBase) CtyValue() (cty.Value, error) {
	if b.disableCtySerialise {
		return cty.Zero, nil
	}
	return GetCtyValue(b)
}
