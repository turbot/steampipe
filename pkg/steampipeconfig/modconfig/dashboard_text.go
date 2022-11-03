package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/utils"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// DashboardText is a struct representing a leaf dashboard node
type DashboardText struct {
	ResourceWithMetadataBase

	FullName        string  `cty:"name" json:"-"`
	ShortName       string  `json:"-"`
	UnqualifiedName string  `json:"-"`
	Value           *string `cty:"value" hcl:"value" column:"value,text" json:"value,omitempty"`

	// these properties are JSON serialised by the parent LeafRun
	Title   *string `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width   *int    `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type    *string `cty:"type" hcl:"type" column:"type,text" json:"-"`
	Display *string `cty:"display" hcl:"display" json:"-"`

	Base       *DashboardText       `hcl:"base" json:"-"`
	DeclRange  hcl.Range            `json:"-"`
	References []*ResourceReference `json:"-"`
	Mod        *Mod                 `cty:"mod" json:"-"`
	Paths      []NodePath           `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewDashboardText(block *hcl.Block, mod *Mod, shortName string) HclResource {
	t := &DashboardText{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}
	t.SetAnonymous(block)
	return t
}

func (t *DashboardText) Equals(other *DashboardText) bool {
	diff := t.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (t *DashboardText) CtyValue() (cty.Value, error) {
	return getCtyValue(t)
}

// Name implements HclResource, ModTreeItem, DashboardLeafNode
// return name in format: 'text.<shortName>'
func (t *DashboardText) Name() string {
	return t.FullName
}

// OnDecoded implements HclResource
func (t *DashboardText) OnDecoded(*hcl.Block, ResourceMapsProvider) hcl.Diagnostics {
	t.setBaseProperties()
	return nil
}

// AddReference implements ResourceWithMetadata
func (t *DashboardText) AddReference(ref *ResourceReference) {
	t.References = append(t.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (t *DashboardText) GetReferences() []*ResourceReference {
	return t.References
}

// GetMod implements ModTreeItem
func (t *DashboardText) GetMod() *Mod {
	return t.Mod
}

// GetDeclRange implements HclResource
func (t *DashboardText) GetDeclRange() *hcl.Range {
	return &t.DeclRange
}

// BlockType implements HclResource
func (*DashboardText) BlockType() string {
	return BlockTypeText
}

// AddParent implements ModTreeItem
func (t *DashboardText) AddParent(parent ModTreeItem) error {
	t.parents = append(t.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (t *DashboardText) GetParents() []ModTreeItem {
	return t.parents
}

// GetChildren implements ModTreeItem
func (t *DashboardText) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements HclResource
func (t *DashboardText) GetTitle() string {
	return typehelpers.SafeString(t.Title)
}

// GetDescription implements ModTreeItem
func (t *DashboardText) GetDescription() string {
	return ""
}

// GetTags implements HclResource
func (t *DashboardText) GetTags() map[string]string {
	return map[string]string{}
}

// GetPaths implements ModTreeItem
func (t *DashboardText) GetPaths() []NodePath {
	// lazy load
	if len(t.Paths) == 0 {
		t.SetPaths()
	}

	return t.Paths
}

// SetPaths implements ModTreeItem
func (t *DashboardText) SetPaths() {
	for _, parent := range t.parents {
		for _, parentPath := range parent.GetPaths() {
			t.Paths = append(t.Paths, append(parentPath, t.Name()))
		}
	}
}

func (t *DashboardText) Diff(other *DashboardText) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: t,
		Name: t.Name(),
	}

	if !utils.SafeStringsEqual(t.Type, other.Type) {
		res.AddPropertyDiff("Type")
	}

	if !utils.SafeStringsEqual(t.Value, other.Value) {
		res.AddPropertyDiff("Value")
	}

	res.populateChildDiffs(t, other)
	res.dashboardLeafNodeDiff(t, other)
	return res
}

// GetWidth implements DashboardLeafNode
func (t *DashboardText) GetWidth() int {
	if t.Width == nil {
		return 0
	}
	return *t.Width
}

// GetDisplay implements DashboardLeafNode
func (t *DashboardText) GetDisplay() string {
	return typehelpers.SafeString(t.Display)
}

// GetDocumentation implements DashboardLeafNode, ModTreeItem
func (*DashboardText) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (t *DashboardText) GetType() string {
	return typehelpers.SafeString(t.Type)
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (t *DashboardText) GetUnqualifiedName() string {
	return t.UnqualifiedName
}

func (t *DashboardText) setBaseProperties() {
	if t.Base == nil {
		return
	}
	if t.Title == nil {
		t.Title = t.Base.Title
	}
	if t.Type == nil {
		t.Type = t.Base.Type
	}
	if t.Display == nil {
		t.Display = t.Base.Display
	}
	if t.Value == nil {
		t.Value = t.Base.Value
	}
	if t.Width == nil {
		t.Width = t.Base.Width
	}
}
