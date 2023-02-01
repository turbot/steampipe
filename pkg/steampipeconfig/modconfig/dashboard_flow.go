package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

// DashboardFlow is a struct representing a leaf dashboard node
type DashboardFlow struct {
	ResourceWithMetadataImpl
	QueryProviderImpl
	WithProviderImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Nodes     DashboardNodeList `cty:"node_list"  column:"nodes,jsonb" json:"-"`
	Edges     DashboardEdgeList `cty:"edge_list" column:"edges,jsonb" json:"-"`
	NodeNames []string          `json:"nodes"`
	EdgeNames []string          `json:"edges"`

	Categories map[string]*DashboardCategory `cty:"categories" json:"categories"`

	Width   *int    `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type    *string `cty:"type" hcl:"type" column:"type,text" json:"-"`
	Display *string `cty:"display" hcl:"display" json:"-"`

	Base *DashboardFlow `hcl:"base" json:"-"`
}

func NewDashboardFlow(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)

	h := &DashboardFlow{
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

func (f *DashboardFlow) Equals(other *DashboardFlow) bool {
	diff := f.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource
func (f *DashboardFlow) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	f.setBaseProperties()
	if len(f.Nodes) > 0 {
		f.NodeNames = f.Nodes.Names()
	}
	if len(f.Edges) > 0 {
		f.EdgeNames = f.Edges.Names()
	}
	return nil
}

// TODO [node_reuse] Add DashboardLeafNodeImpl and move this there https://github.com/turbot/steampipe/issues/2926
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

// ValidateQuery implements QueryProvider
func (*DashboardFlow) ValidateQuery() hcl.Diagnostics {
	// query is optional - nothing to do
	return nil
}

// GetEdges implements NodeAndEdgeProvider
func (f *DashboardFlow) GetEdges() DashboardEdgeList {
	return f.Edges
}

// GetNodes implements NodeAndEdgeProvider
func (f *DashboardFlow) GetNodes() DashboardNodeList {
	return f.Nodes
}

// SetEdges implements NodeAndEdgeProvider
func (f *DashboardFlow) SetEdges(edges DashboardEdgeList) {
	f.Edges = edges
}

// SetNodes implements NodeAndEdgeProvider
func (f *DashboardFlow) SetNodes(nodes DashboardNodeList) {
	f.Nodes = nodes
}

// AddCategory implements NodeAndEdgeProvider
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

// AddChild implements NodeAndEdgeProvider
func (f *DashboardFlow) AddChild(child HclResource) hcl.Diagnostics {
	switch c := child.(type) {
	case *DashboardNode:
		f.Nodes = append(f.Nodes, c)
	case *DashboardEdge:
		f.Edges = append(f.Edges, c)
	default:
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("DashboardFlow does not support children of type %s", child.BlockType()),
			Subject:  f.GetDeclRange(),
		}}
	}
	// set ourselves as parent
	child.(ModTreeItem).AddParent(f)
	return nil
}

// CtyValue implements CtyValueProvider
func (f *DashboardFlow) CtyValue() (cty.Value, error) {
	return GetCtyValue(f)
}

func (f *DashboardFlow) setBaseProperties() {
	if f.Base == nil {
		return
	}
	// copy base into the HclResourceImpl 'base' property so it is accessible to all nested structs
	f.base = f.Base
	// call into parent nested struct setBaseProperties
	f.QueryProviderImpl.setBaseProperties()

	if f.Type == nil {
		f.Type = f.Base.Type
	}

	if f.Display == nil {
		f.Display = f.Base.Display
	}

	if f.Width == nil {
		f.Width = f.Base.Width
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
}
