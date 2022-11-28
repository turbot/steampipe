package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
)

// DashboardFlow is a struct representing a leaf dashboard node
type DashboardFlow struct {
	ResourceWithMetadataBase
	QueryProviderBase
	ModTreeItemBase
	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Nodes DashboardNodeList `cty:"node_list"  hcl:"nodes,optional" column:"nodes,jsonb" json:"-"`
	Edges DashboardEdgeList `cty:"edge_list" hcl:"edges,optional" column:"edges,jsonb" json:"-"`

	Categories map[string]*DashboardCategory `cty:"categories" json:"categories"`

	Width   *int    `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type    *string `cty:"type" hcl:"type" column:"type,text" json:"-"`
	Display *string `cty:"display" hcl:"display" json:"-"`

	Base       *DashboardFlow       `hcl:"base" json:"-"`
	References []*ResourceReference `json:"-"`
}

func NewDashboardFlow(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)

	h := &DashboardFlow{
		Categories: make(map[string]*DashboardCategory),
		QueryProviderBase: QueryProviderBase{
			modNameWithVersion: mod.NameWithVersion(),
			HclResourceBase: HclResourceBase{
				ShortName:       shortName,
				FullName:        fullName,
				UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
				DeclRange:       block.DefRange,
			},
		},
		ModTreeItemBase: ModTreeItemBase{
			Mod:      mod,
			fullName: fullName,
		},
	}
	h.SetAnonymous(block)
	return h
}

func (f *DashboardFlow) Equals(other *DashboardFlow) bool {
	diff := f.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource
func (f *DashboardFlow) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	f.setBaseProperties(resourceMapProvider)

	// populate nodes and edges
	return initialiseEdgesAndNodes(f, resourceMapProvider)
}

// AddReference implements ResourceWithMetadata
func (f *DashboardFlow) AddReference(ref *ResourceReference) {
	f.References = append(f.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (f *DashboardFlow) GetReferences() []*ResourceReference {
	return f.References
}

// TODO KAI PUT IN 1 PLACE FOR ALL EDGE PROVIDERS
// GetChildren implements ModTreeItem
func (f *DashboardFlow) GetChildren() []ModTreeItem {
	// return nodes and edges (if any)
	children := make([]ModTreeItem, len(f.Nodes)+len(f.Edges))
	for i, n := range f.Nodes {
		children[i] = n
	}
	offset := len(f.Nodes)
	for i, e := range f.Edges {
		children[i+offset] = e
	}
	return children
}

func (f *DashboardFlow) Diff(other *DashboardFlow) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: f,
		Name: f.Name(),
	}

	if !utils.SafeStringsEqual(f.Type, other.Type) {
		res.AddPropertyDiff("Type")
	}

	if len(f.Categories) != len(other.Categories) {
		res.AddPropertyDiff("Categories")
	} else {
		for name, c := range f.Categories {
			if !c.Equals(other.Categories[name]) {
				res.AddPropertyDiff("Categories")
			}
		}
	}

	res.populateChildDiffs(f, other)
	res.queryProviderDiff(f, other)
	res.dashboardLeafNodeDiff(f, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (f *DashboardFlow) GetWidth() int {
	if f.Width == nil {
		return 0
	}
	return *f.Width
}

// GetDisplay implements DashboardLeafNode
func (f *DashboardFlow) GetDisplay() string {
	return typehelpers.SafeString(f.Display)
}

// GetType implements DashboardLeafNode
func (f *DashboardFlow) GetType() string {
	return typehelpers.SafeString(f.Type)
}

// VerifyQuery implements QueryProvider
func (*DashboardFlow) VerifyQuery(QueryProvider) error {
	// query is optional - nothing to do
	return nil
}

// GetEdges implements EdgeAndNodeProvider
func (f *DashboardFlow) GetEdges() DashboardEdgeList {
	return f.Edges
}

// GetNodes implements EdgeAndNodeProvider
func (f *DashboardFlow) GetNodes() DashboardNodeList {
	return f.Nodes
}

// SetEdges implements EdgeAndNodeProvider
func (f *DashboardFlow) SetEdges(edges DashboardEdgeList) {
	f.Edges = edges
}

// SetNodes implements EdgeAndNodeProvider
func (f *DashboardFlow) SetNodes(nodes DashboardNodeList) {
	f.Nodes = nodes
}

// AddCategory implements EdgeAndNodeProvider
func (f *DashboardFlow) AddCategory(category *DashboardCategory) hcl.Diagnostics {
	categoryName := category.ShortName
	if _, ok := f.Categories[categoryName]; ok {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s has duplicate category %s", f.Name(), categoryName),
			Subject:  category.GetDeclRange(),
		}}
	}
	f.Categories[categoryName] = category
	return nil
}

func (f *DashboardFlow) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(f.Base, resourceMapProvider); !resolved {
		return
	} else {
		f.Base = base.(*DashboardFlow)
	}

	if f.Title == nil {
		f.Title = f.Base.Title
	}

	if f.Type == nil {
		f.Type = f.Base.Type
	}

	if f.Display == nil {
		f.Display = f.Base.Display
	}

	if f.Width == nil {
		f.Width = f.Base.Width
	}

	if f.SQL == nil {
		f.SQL = f.Base.SQL
	}

	if f.Query == nil {
		f.Query = f.Base.Query
	}

	if f.Args == nil {
		f.Args = f.Base.Args
	}

	if f.Params == nil {
		f.Params = f.Base.Params
	}

	if f.Categories == nil {
		f.Categories = f.Base.Categories
	} else {
		f.Categories = utils.MergeMaps(f.Categories, f.Base.Categories)
	}

	if f.Edges == nil {
		f.Edges = f.Base.Edges
	} else {
		f.Edges.Merge(f.Base.Edges)
	}
	if f.Nodes == nil {
		f.Nodes = f.Base.Nodes
	} else {
		f.Nodes.Merge(f.Base.Nodes)
	}

	f.MergeRuntimeDependencies(f.Base)
}
