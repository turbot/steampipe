package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

// DashboardNode is a struct representing a leaf dashboard node
type DashboardNode struct {
	ResourceWithMetadataBase
	QueryProviderBase
	ModTreeItemBase

	// required to allow partial decoding
	Remain     hcl.Body             `hcl:",remain" json:"-"`
	Category   *DashboardCategory   `cty:"category" hcl:"category" column:"category,jsonb" json:"category,omitempty"`
	Base       *DashboardNode       `hcl:"base" json:"-"`
	References []*ResourceReference `json:"-"`
	Paths      []NodePath           `column:"path,jsonb" json:"-"`
}

func NewDashboardNode(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)
	c := &DashboardNode{
		QueryProviderBase: QueryProviderBase{
			modNameWithVersion: mod.NameWithVersion(),
			HclResourceBase: HclResourceBase{
				ShortName:       shortName,
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

func (n *DashboardNode) Equals(other *DashboardNode) bool {
	diff := n.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource—
func (n *DashboardNode) OnDecoded(_ *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	n.setBaseProperties(resourceMapProvider)

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

// AddReference implements ResourceWithMetadata
func (n *DashboardNode) AddReference(ref *ResourceReference) {
	n.References = append(n.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (n *DashboardNode) GetReferences() []*ResourceReference {
	return n.References
}

func (n *DashboardNode) Diff(other *DashboardNode) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: n,
		Name: n.Name(),
	}

	if !n.Category.Equals(other.Category) {
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

// IsSnapshotPanel implements SnapshotPanel
func (*DashboardNode) IsSnapshotPanel() {}

func (n *DashboardNode) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(n.Base, resourceMapProvider); !resolved {
		return
	} else {
		n.Base = base.(*DashboardNode)
	}

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

	if n.Params == nil {
		n.Params = n.Base.Params
	}

	if n.Category == nil {
		n.Category = n.Base.Category
	}

	n.MergeRuntimeDependencies(n.Base)
}
