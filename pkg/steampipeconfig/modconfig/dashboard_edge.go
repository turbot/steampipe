package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

// DashboardEdge is a struct representing a leaf dashboard node
type DashboardEdge struct {
	ResourceWithMetadataBase
	QueryProviderBase

	// required to allow partial decoding
	Remain     hcl.Body             `hcl:",remain" json:"-"`
	Category   *DashboardCategory   `cty:"category" hcl:"category" column:"category,jsonb" json:"category,omitempty"`
	Base       *DashboardEdge       `hcl:"base" json:"-"`
	References []*ResourceReference `json:"-"`
	Paths      []NodePath           `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewDashboardEdge(block *hcl.Block, mod *Mod, shortName string) HclResource {
	c := &DashboardEdge{
		QueryProviderBase: QueryProviderBase{
			Mod: mod,
			HclResourceBase: HclResourceBase{ShortName: shortName,
				FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
				UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
				DeclRange:       block.DefRange,
				blockType:       block.Type,
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

// AddParent implements ModTreeItem
func (e *DashboardEdge) AddParent(parent ModTreeItem) error {
	e.parents = append(e.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (e *DashboardEdge) GetParents() []ModTreeItem {
	return e.parents
}

// GetChildren implements ModTreeItem
func (e *DashboardEdge) GetChildren() []ModTreeItem {
	return nil
}

// GetPaths implements ModTreeItem
func (e *DashboardEdge) GetPaths() []NodePath {
	// lazy load
	if len(e.Paths) == 0 {
		e.SetPaths()
	}

	return e.Paths
}

// SetPaths implements ModTreeItem
func (e *DashboardEdge) SetPaths() {
	for _, parent := range e.parents {
		for _, parentPath := range parent.GetPaths() {
			e.Paths = append(e.Paths, append(parentPath, e.Name()))
		}
	}
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
