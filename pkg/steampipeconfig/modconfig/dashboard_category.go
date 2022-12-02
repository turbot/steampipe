package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

type DashboardCategory struct {
	ResourceWithMetadataBase

	ShortName       string `hcl:"name,label" json:"name"`
	FullName        string `cty:"name" json:"-"`
	UnqualifiedName string `json:"-"`

	Title         *string                               `cty:"title" hcl:"title" json:"title,omitempty"`
	Color         *string                               `cty:"color" hcl:"color" json:"color,omitempty"`
	Depth         *int                                  `cty:"depth" hcl:"depth" json:"depth,omitempty"`
	Icon          *string                               `cty:"icon" hcl:"icon" json:"icon,omitempty"`
	HREF          *string                               `cty:"href" hcl:"href" json:"href,omitempty"`
	Fold          *DashboardCategoryFold                `cty:"fold" hcl:"fold,block" json:"fold,omitempty"`
	PropertyList  DashboardCategoryPropertyList         `cty:"property_list" hcl:"property,block" column:"properties,jsonb" json:"-"`
	Properties    map[string]*DashboardCategoryProperty `cty:"properties" json:"properties,omitempty"`
	PropertyOrder []string                              `cty:"property_order" hcl:"property_order,optional" json:"property_order,omitempty"`
	Base          *DashboardCategory                    `hcl:"base" json:"-"`
	References    []*ResourceReference                  `json:"-"`
	Mod           *Mod                                  `cty:"mod" json:"-"`
	DeclRange     hcl.Range                             `json:"-"`
	Paths         []NodePath                            `column:"path,jsonb" json:"-"`
	Parents       []ModTreeItem                         `json:"-"`
}

func NewDashboardCategory(block *hcl.Block, mod *Mod, shortName string) HclResource {
	c := &DashboardCategory{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}
	c.SetAnonymous(block)
	return c
}

// Name implements HclResource
// return name in format: '<modname>.control.<shortName>'
func (c *DashboardCategory) Name() string {
	return c.FullName
}

// GetUnqualifiedName implements HclResource
func (c *DashboardCategory) GetUnqualifiedName() string {
	return c.UnqualifiedName
}

// CtyValue implements HclResource
func (c *DashboardCategory) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

// GetDeclRange implements HclResource
func (c *DashboardCategory) GetDeclRange() *hcl.Range {
	return &c.DeclRange
}

// BlockType implements HclResource
func (*DashboardCategory) BlockType() string {
	return BlockTypeCategory
}

// OnDecoded implements HclResource
func (c *DashboardCategory) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	c.setBaseProperties(resourceMapProvider)
	// populate properties map
	if len(c.PropertyList) > 0 {
		c.Properties = make(map[string]*DashboardCategoryProperty, len(c.PropertyList))
		for _, p := range c.PropertyList {
			c.Properties[p.ShortName] = p
		}
	}
	return nil
}

// AddReference implements ResourceWithMetadata
func (c *DashboardCategory) AddReference(ref *ResourceReference) {
	c.References = append(c.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (c *DashboardCategory) GetReferences() []*ResourceReference {
	return c.References
}

func (c *DashboardCategory) Equals(other *DashboardCategory) bool {
	if other == nil {
		return false
	}
	return !c.Diff(other).HasChanges()
}

func (c *DashboardCategory) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(c.Base, resourceMapProvider); !resolved {
		return
	} else {
		c.Base = base.(*DashboardCategory)
	}

	if c.Title == nil {
		c.Title = c.Base.Title
	}
	if c.Color == nil {
		c.Color = c.Base.Color
	}
	if c.Depth == nil {
		c.Depth = c.Base.Depth
	}
	if c.Icon == nil {
		c.Icon = c.Base.Icon
	}
	if c.HREF == nil {
		c.HREF = c.Base.HREF
	}
	if c.Fold == nil {
		c.Fold = c.Base.Fold
	}

	if c.PropertyList == nil {
		c.PropertyList = c.Base.PropertyList
	} else {
		c.PropertyList.Merge(c.Base.PropertyList)
	}

	if c.PropertyOrder == nil {
		c.PropertyOrder = c.Base.PropertyOrder
	}
}

// AddParent implements ModTreeItem
func (c *DashboardCategory) AddParent(parent ModTreeItem) error {
	c.Parents = append(c.Parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (c *DashboardCategory) GetParents() []ModTreeItem {
	return c.Parents
}

// GetTitle implements HclResource
func (c *DashboardCategory) GetTitle() string {
	return typehelpers.SafeString(c.Title)
}

// GetDescription implements ModTreeItem, DashboardLeafNode
func (c *DashboardCategory) GetDescription() string {
	return ""
}

// GetTags implements HclResource
func (c *DashboardCategory) GetTags() map[string]string {
	return map[string]string{}
}

// GetChildren implements ModTreeItem
func (c *DashboardCategory) GetChildren() []ModTreeItem {
	return nil
}

// GetDocumentation implements DashboardLeafNode, ModTreeItem
func (*DashboardCategory) GetDocumentation() string {
	return ""
}

// GetMod implements ModTreeItem
func (c *DashboardCategory) GetMod() *Mod {
	return c.Mod
}

// GetPaths implements ModTreeItem
func (c *DashboardCategory) GetPaths() []NodePath {
	// lazy load
	if len(c.Paths) == 0 {
		c.SetPaths()
	}

	return c.Paths
}

// SetPaths implements ModTreeItem
func (c *DashboardCategory) SetPaths() {
	for _, parent := range c.Parents {
		for _, parentPath := range parent.GetPaths() {
			c.Paths = append(c.Paths, append(parentPath, c.Name()))
		}
	}
}

func (c *DashboardCategory) Diff(other *DashboardCategory) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: c,
		Name: c.Name(),
	}

	if (c.Fold == nil) != (other.Fold == nil) {
		res.AddPropertyDiff("Fold")
	}
	if c.Fold != nil && !c.Fold.Equals(other.Fold) {
		res.AddPropertyDiff("Fold")
	}

	if len(c.PropertyList) != len(other.PropertyList) {
		res.AddPropertyDiff("Properties")
	} else {
		for i, p := range c.Properties {
			if !p.Equals(other.Properties[i]) {
				res.AddPropertyDiff("Properties")
			}
		}
	}

	if len(c.PropertyOrder) != len(other.PropertyOrder) {
		res.AddPropertyDiff("PropertyOrder")
	} else {
		for i, p := range c.PropertyOrder {
			if p != other.PropertyOrder[i] {
				res.AddPropertyDiff("PropertyOrder")
			}
		}
	}

	if !utils.SafeStringsEqual(c.Name, other.Name) {
		res.AddPropertyDiff("Name")
	}
	if !utils.SafeStringsEqual(c.Title, other.Title) {
		res.AddPropertyDiff("Title")
	}
	if !utils.SafeStringsEqual(c.Color, other.Color) {
		res.AddPropertyDiff("Color")
	}
	if !utils.SafeStringsEqual(c.Depth, other.Depth) {
		res.AddPropertyDiff("Depth")
	}
	if !utils.SafeStringsEqual(c.Icon, other.Icon) {
		res.AddPropertyDiff("Icon")
	}
	if !utils.SafeStringsEqual(c.HREF, other.HREF) {
		res.AddPropertyDiff("HREF")
	}

	return res
}
