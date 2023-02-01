package modconfig

import (
	"fmt"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
)

// DashboardEdge is a struct representing a leaf dashboard node
type DashboardEdge struct {
	ResourceWithMetadataImpl
	QueryProviderImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Category *DashboardCategory `cty:"category" hcl:"category" column:"category,jsonb" json:"category,omitempty"`
	Base     *DashboardEdge     `hcl:"base" json:"-"`
}

func NewDashboardEdge(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)

	c := &DashboardEdge{
		QueryProviderImpl: QueryProviderImpl{
			RuntimeDependencyProviderImpl: RuntimeDependencyProviderImpl{
				ModTreeItemImpl: ModTreeItemImpl{
					HclResourceImpl: HclResourceImpl{ShortName: shortName,
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

func (e *DashboardEdge) Equals(other *DashboardEdge) bool {
	diff := e.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource
func (e *DashboardEdge) OnDecoded(_ *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	e.setBaseProperties()

	// when we reference resources (i.e. category),
	// not all properties are retrieved as they are no cty serialisable
	// repopulate category from resourceMapProvider
	if e.Category != nil {
		fullCategory, diags := enrichCategory(e.Category, e, resourceMapProvider)
		if diags.HasErrors() {
			return diags
		}
		e.Category = fullCategory
	}
	return nil
}

func (e *DashboardEdge) Diff(other *DashboardEdge) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: e,
		Name: e.Name(),
	}
	if (e.Category == nil) != (other.Category == nil) {
		res.AddPropertyDiff("Category")
	}

	if e.Category != nil && !e.Category.Equals(other.Category) {
		res.AddPropertyDiff("Category")
	}

	res.populateChildDiffs(e, other)
	res.queryProviderDiff(e, other)
	res.dashboardLeafNodeDiff(e, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (e *DashboardEdge) GetWidth() int {
	return 0
}

// GetDisplay implements DashboardLeafNode
func (e *DashboardEdge) GetDisplay() string {
	return ""
}

// GetDocumentation implements DashboardLeafNode
func (e *DashboardEdge) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (e *DashboardEdge) GetType() string {
	return ""
}

// CtyValue implements CtyValueProvider
func (e *DashboardEdge) CtyValue() (cty.Value, error) {
	return GetCtyValue(e)
}

func (e *DashboardEdge) setBaseProperties() {
	if e.Base == nil {
		return
	}
	// copy base into the HclResourceImpl 'base' property so it is accessible to all nested structs
	e.base = e.Base
	// call into parent nested struct setBaseProperties
	e.QueryProviderImpl.setBaseProperties()

	if e.Category == nil {
		e.Category = e.Base.Category
	}
}
