package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/utils"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
)

// DashboardText is a struct representing a leaf dashboard node
type DashboardText struct {
	ResourceWithMetadataBase
	HclResourceBase
	ModTreeItemBase

	Value   *string `cty:"value" hcl:"value" column:"value,text" json:"value,omitempty"`
	Width   *int    `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type    *string `cty:"type" hcl:"type" column:"type,text" json:"-"`
	Display *string `cty:"display" hcl:"display" json:"-"`

	Base       *DashboardText       `hcl:"base" json:"-"`
	References []*ResourceReference `json:"-"`
	Mod        *Mod                 `cty:"mod" json:"-"`
}

func NewDashboardText(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)

	t := &DashboardText{
		HclResourceBase: HclResourceBase{
			ShortName:       shortName,
			FullName:        fullName,
			UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
			DeclRange:       block.DefRange,
			blockType:       block.Type,
		},
		ModTreeItemBase: ModTreeItemBase{
			Mod:      mod,
			fullName: fullName,
		},
	}
	t.SetAnonymous(block)
	return t
}

func (t *DashboardText) Equals(other *DashboardText) bool {
	diff := t.Diff(other)
	return !diff.HasChanges()
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
