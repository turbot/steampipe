package modconfig

import (
	"fmt"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
)

// DashboardNode is a struct representing a leaf dashboard node
type DashboardNode struct {
	ResourceWithMetadataImpl
	QueryProviderImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Category *DashboardCategory `cty:"category" hcl:"category" column:"category,jsonb" json:"category,omitempty"`
	Base     *DashboardNode     `hcl:"base" json:"-"`
}

func NewDashboardNode(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)
	c := &DashboardNode{
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

func (n *DashboardNode) Equals(other *DashboardNode) bool {
	diff := n.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource—
func (n *DashboardNode) OnDecoded(_ *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	n.setBaseProperties()

	// when we reference resources (i.e. category),
	// not all properties are retrieved as they are no cty serialisable
	// repopulate category from resourceMapProvider
	if n.Category != nil {
		fullCategory, diags := enrichCategory(n.Category, n, resourceMapProvider)
		if diags.HasErrors() {
			return diags
		}
		n.Category = fullCategory
	}
	return nil
}

func (n *DashboardNode) Diff(other *DashboardNode) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: n,
		Name: n.Name(),
	}

	if (n.Category == nil) != (other.Category == nil) {
		res.AddPropertyDiff("Category")
	}
	if n.Category != nil && !n.Category.Equals(other.Category) {
		res.AddPropertyDiff("Category")
	}

	res.populateChildDiffs(n, other)
	res.queryProviderDiff(n, other)
	res.dashboardLeafNodeDiff(n, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (n *DashboardNode) GetWidth() int {
	return 0
}

// GetDisplay implements DashboardLeafNode
func (n *DashboardNode) GetDisplay() string {
	return ""
}

// GetType implements DashboardLeafNode
func (n *DashboardNode) GetType() string {
	return ""
}

// CtyValue implements CtyValueProvider
func (n *DashboardNode) CtyValue() (cty.Value, error) {
	return GetCtyValue(n)
}

func (n *DashboardNode) setBaseProperties() {
	if n.Base == nil {
		return
	}
	// copy base into the HclResourceImpl 'base' property so it is accessible to all nested structs
	n.base = n.Base
	// call into parent nested struct setBaseProperties
	n.QueryProviderImpl.setBaseProperties()

	if n.Title == nil {
		n.Title = n.Base.Title
	}

	if n.SQL == nil {
		n.SQL = n.Base.SQL
	}

	if n.Query == nil {
		n.Query = n.Base.Query
	}

	if n.Args == nil {
		n.Args = n.Base.Args
	}

	if n.Category == nil {
		n.Category = n.Base.Category
	}

	if n.Params == nil {
		n.Params = n.Base.Params
	}
}
