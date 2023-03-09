package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type ModTreeItemImpl struct {
	HclResourceImpl
	// required to allow partial decoding
	ModTreeItemRemain hcl.Body `hcl:",remain" json:"-"`

	Mod   *Mod       `cty:"mod" json:"-"`
	Paths []NodePath `column:"path,jsonb" json:"-"`

	parents  []ModTreeItem
	children []ModTreeItem
}

// AddParent implements ModTreeItem
func (b *ModTreeItemImpl) AddParent(parent ModTreeItem) error {
	b.parents = append(b.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (b *ModTreeItemImpl) GetParents() []ModTreeItem {
	return b.parents
}

// GetChildren implements ModTreeItem
func (b *ModTreeItemImpl) GetChildren() []ModTreeItem {
	return b.children
}
func (b *ModTreeItemImpl) GetPaths() []NodePath {
	// lazy load
	if len(b.Paths) == 0 {
		b.SetPaths()
	}
	return b.Paths
}

// SetPaths implements ModTreeItem
func (b *ModTreeItemImpl) SetPaths() {
	for _, parent := range b.parents {
		for _, parentPath := range parent.GetPaths() {
			b.Paths = append(b.Paths, append(parentPath, b.FullName))
		}
	}
}
func (b *ModTreeItemImpl) GetMod() *Mod {
	return b.Mod
}

// GetModTreeItemBase implements ModTreeItem
func (b *ModTreeItemImpl) GetModTreeItemImpl() *ModTreeItemImpl {
	return b
}

// CtyValue implements CtyValueProvider
func (b *ModTreeItemImpl) CtyValue() (cty.Value, error) {
	if b.disableCtySerialise {
		return cty.Zero, nil
	}
	return GetCtyValue(b)
}
