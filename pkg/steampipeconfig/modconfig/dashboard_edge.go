package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

// DashboardEdge is a struct representing a leaf dashboard node
type DashboardEdge struct {
	ResourceWithMetadataBase
	QueryProviderBase
	ModTreeItemBase

	// required to allow partial decoding
	Remain     hcl.Body             `hcl:",remain" json:"-"`
	Category   *DashboardCategory   `cty:"category" hcl:"category" column:"category,jsonb" json:"category,omitempty"`
	Base       *DashboardEdge       `hcl:"base" json:"-"`
	References []*ResourceReference `json:"-"`
	Paths      []NodePath           `column:"path,jsonb" json:"-"`
}

func NewDashboardEdge(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)

	c := &DashboardEdge{
		QueryProviderBase: QueryProviderBase{
			modNameWithVersion: mod.NameWithVersion(),
			HclResourceBase: HclResourceBase{ShortName: shortName,
				FullName:        fullName,
				UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
		},
		ModTreeItemBase: ModTreeItemBase{
			Mod:      mod,
			fullName: fullName,
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
	e.setBaseProperties(resourceMapProvider)

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

// AddReference implements ResourceWithMetadata
func (e *DashboardEdge) AddReference(ref *ResourceReference) {
	e.References = append(e.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (e *DashboardEdge) GetReferences() []*ResourceReference {
	return e.References
}

func (e *DashboardEdge) Diff(other *DashboardEdge) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: e,
		Name: e.Name(),
	}

	if !e.Category.Equals(other.Category) {
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

// IsSnapshotPanel implements SnapshotPanel
func (*DashboardEdge) IsSnapshotPanel() {}

func (e *DashboardEdge) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(e.Base, resourceMapProvider); !resolved {
		return
	} else {
		e.Base = base.(*DashboardEdge)
	}

	if e.Title == nil {
		e.Title = e.Base.Title
	}

	if e.SQL == nil {
		e.SQL = e.Base.SQL
	}

	if e.Query == nil {
		e.Query = e.Base.Query
	}

	if e.Args == nil {
		e.Args = e.Base.Args
	}

	if e.Params == nil {
		e.Params = e.Base.Params
	}

	if e.Category == nil {
		e.Category = e.Base.Category
	}

	e.MergeRuntimeDependencies(e.Base)
}
