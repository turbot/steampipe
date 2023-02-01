package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

type DashboardCategory struct {
	ResourceWithMetadataImpl
	ModTreeItemImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	// TACTICAL: include a title property (with a different name to the property in HclResourceImpl  for clarity)
	// This is purely to ensure the title is included in the panel properties of snapshots
	// Note: this will be parsed from HCL, but we must set this explicitly in setBaseProperties if there is a base
	CategoryTitle *string                               `cty:"title" hcl:"title" json:"title,omitempty"`
	Color         *string                               `cty:"color" hcl:"color" json:"color,omitempty"`
	Depth         *int                                  `cty:"depth" hcl:"depth" json:"depth,omitempty"`
	Icon          *string                               `cty:"icon" hcl:"icon" json:"icon,omitempty"`
	HREF          *string                               `cty:"href" hcl:"href" json:"href,omitempty"`
	Fold          *DashboardCategoryFold                `cty:"fold" hcl:"fold,block" json:"fold,omitempty"`
	PropertyList  DashboardCategoryPropertyList         `cty:"property_list" hcl:"property,block" column:"properties,jsonb" json:"-"`
	Properties    map[string]*DashboardCategoryProperty `cty:"properties" json:"properties,omitempty"`
	PropertyOrder []string                              `cty:"property_order" hcl:"property_order,optional" json:"property_order,omitempty"`
	Base          *DashboardCategory                    `hcl:"base" json:"-"`
}

func NewDashboardCategory(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)

	c := &DashboardCategory{
		ModTreeItemImpl: ModTreeItemImpl{
			HclResourceImpl: HclResourceImpl{
				ShortName:       shortName,
				FullName:        fullName,
				UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
				DeclRange:       BlockRange(block),
				blockType:       block.Type,
			},
			Mod: mod,
		},
	}
	c.SetAnonymous(block)
	return c
}

// OnDecoded implements HclResource
func (c *DashboardCategory) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	c.setBaseProperties()
	// populate properties map
	if len(c.PropertyList) > 0 {
		c.Properties = make(map[string]*DashboardCategoryProperty, len(c.PropertyList))
		for _, p := range c.PropertyList {
			c.Properties[p.ShortName] = p
		}
	}
	return nil
}

func (c *DashboardCategory) Equals(other *DashboardCategory) bool {
	if other == nil {
		return false
	}
	return !c.Diff(other).HasChanges()
}

func (c *DashboardCategory) setBaseProperties() {
	if c.Base == nil {
		return
	}
	// copy base into the HclResourceImpl 'base' property so it is accessible to all nested structs
	c.base = c.Base
	// call into parent nested struct setBaseProperties
	c.ModTreeItemImpl.setBaseProperties()

	// TACTICAL: DashboardCategory overrides the title property to ensure is included in the snapshot
	c.CategoryTitle = c.Title

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

// CtyValue implements CtyValueProvider
func (c *DashboardCategory) CtyValue() (cty.Value, error) {
	return GetCtyValue(c)
}
