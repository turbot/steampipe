package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

// Local is a struct representing a Local resource
type Local struct {
	ShortName       string
	FullName        string `cty:"name"`
	UnqualifiedName string

	Value     cty.Value
	DeclRange hcl.Range
	Mod       *Mod `cty:"mod"`
	metadata  *ResourceMetadata
	Paths     []NodePath `column:"path,jsonb"`
	parents   []ModTreeItem
}

func NewLocal(name string, val cty.Value, declRange hcl.Range, mod *Mod) *Local {
	l := &Local{
		ShortName:       name,
		UnqualifiedName: fmt.Sprintf("local.%s", name),
		FullName:        fmt.Sprintf("%s.local.%s", mod.ShortName, name),
		Value:           val,
		Mod:             mod,
		DeclRange:       declRange,
	}
	return l
}

// Name implements HclResource, ResourceWithMetadata
func (l *Local) Name() string {
	return l.FullName
}

// OnDecoded implements HclResource
func (l *Local) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	return nil
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (l *Local) GetUnqualifiedName() string {
	return l.UnqualifiedName
}

// GetMod implements ModTreeItem
func (l *Local) GetMod() *Mod {
	return l.Mod
}

// CtyValue implements HclResource
func (l *Local) CtyValue() (cty.Value, error) {
	return l.Value, nil
}

// GetDeclRange implements HclResource
func (l *Local) GetDeclRange() *hcl.Range {
	return &l.DeclRange
}

// BlockType implements HclResource
func (*Local) BlockType() string {
	return BlockTypeLocals
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

// GetDescription implements ModTreeItem
func (l *Local) GetDescription() string {
	return ""
}

// GetTitle implements HclResource
func (l *Local) GetTitle() string {
	return typehelpers.SafeString(l.FullName)
}

// GetTags implements HclResource
func (l *Local) GetTags() map[string]string {
	return map[string]string{}
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
