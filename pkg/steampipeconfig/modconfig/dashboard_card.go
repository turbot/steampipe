package modconfig

import (
	"fmt"
	"github.com/zclconf/go-cty/cty"

	"github.com/turbot/steampipe/pkg/utils"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
)

// DashboardCard is a struct representing a leaf dashboard node
type DashboardCard struct {
	ResourceWithMetadataImpl
	QueryProviderImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Label *string `cty:"label" hcl:"label" column:"label,text" json:"label,omitempty"`
	Value *string `cty:"value" hcl:"value" column:"value,text" json:"value,omitempty"`
	Icon  *string `cty:"icon" hcl:"icon" column:"icon,text" json:"icon,omitempty"`
	HREF  *string `cty:"href" hcl:"href" json:"href,omitempty"`

	Width   *int           `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type    *string        `cty:"type" hcl:"type" column:"type,text" json:"-"`
	Display *string        `cty:"display" hcl:"display" json:"-"`
	Base    *DashboardCard `hcl:"base" json:"-"`

	metadata *ResourceMetadata
}

func NewDashboardCard(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)
	c := &DashboardCard{
		QueryProviderImpl: QueryProviderImpl{
			RuntimeDependencyProviderImpl: RuntimeDependencyProviderImpl{
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
			},
		},
	}

	c.SetAnonymous(block)
	return c
}

func (c *DashboardCard) Equals(other *DashboardCard) bool {
	diff := c.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource
func (c *DashboardCard) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	c.setBaseProperties()
	return nil
}

func (c *DashboardCard) Diff(other *DashboardCard) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: c,
		Name: c.Name(),
	}

	if !utils.SafeStringsEqual(c.Label, other.Label) {
		res.AddPropertyDiff("Label")
	}

	if !utils.SafeStringsEqual(c.Value, other.Value) {
		res.AddPropertyDiff("Value")
	}

	if !utils.SafeStringsEqual(c.Type, other.Type) {
		res.AddPropertyDiff("Type")
	}

	if !utils.SafeStringsEqual(c.Icon, other.Icon) {
		res.AddPropertyDiff("Icon")
	}

	if !utils.SafeStringsEqual(c.HREF, other.HREF) {
		res.AddPropertyDiff("HREF")
	}

	res.populateChildDiffs(c, other)
	res.queryProviderDiff(c, other)
	res.dashboardLeafNodeDiff(c, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (c *DashboardCard) GetWidth() int {
	if c.Width == nil {
		return 0
	}
	return *c.Width
}

// GetDisplay implements DashboardLeafNode
func (c *DashboardCard) GetDisplay() string {
	return typehelpers.SafeString(c.Display)
}

// GetDocumentation implements DashboardLeafNode, ModTreeItem
func (c *DashboardCard) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (c *DashboardCard) GetType() string {
	return typehelpers.SafeString(c.Type)
}

// ValidateQuery implements QueryProvider
func (c *DashboardCard) ValidateQuery() hcl.Diagnostics {
	// query is optional - nothing to do
	return nil
}

// CtyValue implements CtyValueProvider
func (c *DashboardCard) CtyValue() (cty.Value, error) {
	return GetCtyValue(c)
}

func (c *DashboardCard) setBaseProperties() {
	if c.Base == nil {
		return
	}
	// copy base into the HclResourceImpl 'base' property so it is accessible to all nested structs
	c.base = c.Base
	// call into parent nested struct setBaseProperties
	c.QueryProviderImpl.setBaseProperties()

	if c.Label == nil {
		c.Label = c.Base.Label
	}

	if c.Value == nil {
		c.Value = c.Base.Value
	}

	if c.Type == nil {
		c.Type = c.Base.Type
	}

	if c.Display == nil {
		c.Display = c.Base.Display
	}

	if c.Icon == nil {
		c.Icon = c.Base.Icon
	}

	if c.HREF == nil {
		c.HREF = c.Base.HREF
	}

	if c.Width == nil {
		c.Width = c.Base.Width
	}
}
