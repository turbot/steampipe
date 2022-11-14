package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
)

// DashboardGraph is a struct representing a leaf dashboard node
type DashboardGraph struct {
	ResourceWithMetadataBase
	QueryProviderBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Nodes DashboardNodeList `cty:"node_list"  hcl:"nodes,optional" column:"nodes,jsonb" json:"-"`
	Edges DashboardEdgeList `cty:"edge_list" hcl:"edges,optional" column:"edges,jsonb" json:"-"`

	Categories map[string]*DashboardCategory `cty:"categories" json:"categories"`
	Direction  *string                       `cty:"direction" hcl:"direction" column:"direction,text" json:"direction"`

	// these properties are JSON serialised by the parent LeafRun
	Width   *int    `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type    *string `cty:"type" hcl:"type" column:"type,text" json:"-"`
	Display *string `cty:"display" hcl:"display" json:"-"`

	Base       *DashboardGraph      `hcl:"base" json:"-"`
	References []*ResourceReference `json:"-"`
	Paths      []NodePath           `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewDashboardGraph(block *hcl.Block, mod *Mod, shortName string) HclResource {
	h := &DashboardGraph{
		QueryProviderBase: QueryProviderBase{
			Mod: mod,
			HclResourceBase: HclResourceBase{
				ShortName:       shortName,
				FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
				UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
		},
		Categories: make(map[string]*DashboardCategory),
	}
	h.SetAnonymous(block)
	return h
}

func (g *DashboardGraph) Equals(other *DashboardGraph) bool {
	diff := g.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource
func (g *DashboardGraph) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	g.setBaseProperties(resourceMapProvider)

	// populate nodes and edges
	return initialiseEdgesAndNodes(g, resourceMapProvider)
}

// AddReference implements ResourceWithMetadata
func (g *DashboardGraph) AddReference(ref *ResourceReference) {
	g.References = append(g.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (g *DashboardGraph) GetReferences() []*ResourceReference {
	return g.References
}

// AddParent implements ModTreeItem
func (g *DashboardGraph) AddParent(parent ModTreeItem) error {
	g.parents = append(g.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (g *DashboardGraph) GetParents() []ModTreeItem {
	return g.parents
}

// GetChildren implements ModTreeItem
func (g *DashboardGraph) GetChildren() []ModTreeItem {
	// return nodes and edges (if any)
	children := make([]ModTreeItem, len(g.Nodes)+len(g.Edges))
	for i, n := range g.Nodes {
		children[i] = n
	}
	offset := len(g.Nodes)
	for i, e := range g.Edges {
		children[i+offset] = e
	}
	return children
}

// GetPaths implements ModTreeItem
func (g *DashboardGraph) GetPaths() []NodePath {
	// lazy load
	if len(g.Paths) == 0 {
		g.SetPaths()
	}

	return g.Paths
}

// SetPaths implements ModTreeItem
func (g *DashboardGraph) SetPaths() {
	for _, parent := range g.parents {
		for _, parentPath := range parent.GetPaths() {
			g.Paths = append(g.Paths, append(parentPath, g.Name()))
		}
	}
}

func (g *DashboardGraph) Diff(other *DashboardGraph) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: g,
		Name: g.Name(),
	}

	if !utils.SafeStringsEqual(g.Type, other.Type) {
		res.AddPropertyDiff("Type")
	}

	if !utils.SafeStringsEqual(g.Direction, other.Direction) {
		res.AddPropertyDiff("Direction")
	}

	if len(g.Categories) != len(other.Categories) {
		res.AddPropertyDiff("Categories")
	} else {
		for name, c := range g.Categories {
			if !c.Equals(other.Categories[name]) {
				res.AddPropertyDiff("Categories")
			}
		}
	}

	res.populateChildDiffs(g, other)
	res.queryProviderDiff(g, other)
	res.dashboardLeafNodeDiff(g, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (g *DashboardGraph) GetWidth() int {
	if g.Width == nil {
		return 0
	}
	return *g.Width
}

// GetDisplay implements DashboardLeafNode
func (g *DashboardGraph) GetDisplay() string {
	return typehelpers.SafeString(g.Display)
}

// GetDocumentation implements DashboardLeafNode, ModTreeItem
func (g *DashboardGraph) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (g *DashboardGraph) GetType() string {
	return typehelpers.SafeString(g.Type)
}

// GetEdges implements EdgeAndNodeProvider
func (g *DashboardGraph) GetEdges() DashboardEdgeList {
	return g.Edges
}

// GetNodes implements EdgeAndNodeProvider
func (g *DashboardGraph) GetNodes() DashboardNodeList {
	return g.Nodes
}

// SetEdges implements EdgeAndNodeProvider
func (g *DashboardGraph) SetEdges(edges DashboardEdgeList) {
	g.Edges = edges
}

// SetNodes implements EdgeAndNodeProvider
func (g *DashboardGraph) SetNodes(nodes DashboardNodeList) {
	g.Nodes = nodes
}

// AddCategory implements EdgeAndNodeProvider
func (g *DashboardGraph) AddCategory(category *DashboardCategory) hcl.Diagnostics {
	categoryName := category.ShortName
	if _, ok := g.Categories[categoryName]; ok {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s has duplicate category %s", g.Name(), categoryName),
			Subject:  category.GetDeclRange(),
		}}
	}
	g.Categories[categoryName] = category
	return nil
}

func (g *DashboardGraph) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(g.Base, resourceMapProvider); !resolved {
		return
	} else {
		g.Base = base.(*DashboardGraph)
	}

	if g.Title == nil {
		g.Title = g.Base.Title
	}

	if g.Type == nil {
		g.Type = g.Base.Type
	}

	if g.Display == nil {
		g.Display = g.Base.Display
	}

	if g.Width == nil {
		g.Width = g.Base.Width
	}

	if g.SQL == nil {
		g.SQL = g.Base.SQL
	}

	if g.Query == nil {
		g.Query = g.Base.Query
	}

	if g.Args == nil {
		g.Args = g.Base.Args
	}

	if g.Params == nil {
		g.Params = g.Base.Params
	}

	if g.Categories == nil {
		g.Categories = g.Base.Categories
	} else {
		g.Categories = utils.MergeMaps(g.Categories, g.Base.Categories)
	}

	if g.Direction == nil {
		g.Direction = g.Base.Direction
	}

	if g.Edges == nil {
		g.Edges = g.Base.Edges
	} else {
		g.Edges.Merge(g.Base.Edges)
	}
	if g.Nodes == nil {
		g.Nodes = g.Base.Nodes
	} else {
		g.Nodes.Merge(g.Base.Nodes)
	}

	g.MergeRuntimeDependencies(g.Base)
}
