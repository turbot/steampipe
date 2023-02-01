package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

// DashboardGraph is a struct representing a leaf dashboard node
type DashboardGraph struct {
	ResourceWithMetadataImpl
	QueryProviderImpl
	WithProviderImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Nodes     DashboardNodeList `cty:"node_list" column:"nodes,jsonb" json:"-"`
	Edges     DashboardEdgeList `cty:"edge_list" column:"edges,jsonb" json:"-"`
	NodeNames []string          `json:"nodes"`
	EdgeNames []string          `json:"edges"`

	Categories map[string]*DashboardCategory `cty:"categories" json:"categories"`
	Direction  *string                       `cty:"direction" hcl:"direction" column:"direction,text" json:"direction"`

	// these properties are JSON serialised by the parent LeafRun
	Width   *int    `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type    *string `cty:"type" hcl:"type" column:"type,text" json:"-"`
	Display *string `cty:"display" hcl:"display" json:"-"`

	Base *DashboardGraph `hcl:"base" json:"-"`
}

func NewDashboardGraph(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)

	h := &DashboardGraph{
		Categories: make(map[string]*DashboardCategory),
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
	h.SetAnonymous(block)
	return h
}

func (g *DashboardGraph) Equals(other *DashboardGraph) bool {
	diff := g.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource
func (g *DashboardGraph) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	g.setBaseProperties()
	if len(g.Nodes) > 0 {
		g.NodeNames = g.Nodes.Names()
	}
	if len(g.Edges) > 0 {
		g.EdgeNames = g.Edges.Names()
	}
	return nil
}

// TODO [node_reuse] Add DashboardLeafNodeImpl and move this there https://github.com/turbot/steampipe/issues/2926

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

// GetType implements DashboardLeafNode
func (g *DashboardGraph) GetType() string {
	return typehelpers.SafeString(g.Type)
}

// GetEdges implements NodeAndEdgeProvider
func (g *DashboardGraph) GetEdges() DashboardEdgeList {
	return g.Edges
}

// GetNodes implements NodeAndEdgeProvider
func (g *DashboardGraph) GetNodes() DashboardNodeList {
	return g.Nodes
}

// SetEdges implements NodeAndEdgeProvider
func (g *DashboardGraph) SetEdges(edges DashboardEdgeList) {
	g.Edges = edges
}

// SetNodes implements NodeAndEdgeProvider
func (g *DashboardGraph) SetNodes(nodes DashboardNodeList) {
	g.Nodes = nodes
}

// AddCategory implements NodeAndEdgeProvider
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

// AddChild implements NodeAndEdgeProvider
func (g *DashboardGraph) AddChild(child HclResource) hcl.Diagnostics {
	switch c := child.(type) {
	case *DashboardNode:
		g.Nodes = append(g.Nodes, c)
	case *DashboardEdge:
		g.Edges = append(g.Edges, c)
	default:
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("DashboardGraph does not support children of type %s", child.BlockType()),
			Subject:  g.GetDeclRange(),
		}}
	}
	// set ourselves as parent
	child.(ModTreeItem).AddParent(g)
	return nil
}

// CtyValue implements CtyValueProvider
func (g *DashboardGraph) CtyValue() (cty.Value, error) {
	return GetCtyValue(g)
}

func (g *DashboardGraph) setBaseProperties() {
	if g.Base == nil {
		return
	}
	// copy base into the HclResourceImpl 'base' property so it is accessible to all nested structs
	g.base = g.Base
	// call into parent nested struct setBaseProperties
	g.QueryProviderImpl.setBaseProperties()

	if g.Type == nil {
		g.Type = g.Base.Type
	}

	if g.Display == nil {
		g.Display = g.Base.Display
	}

	if g.Width == nil {
		g.Width = g.Base.Width
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
}
