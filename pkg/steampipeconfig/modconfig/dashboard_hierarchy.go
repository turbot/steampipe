package modconfig

import (
	"fmt"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
)

// DashboardHierarchy is a struct representing a leaf dashboard node
type DashboardHierarchy struct {
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
	Width      *int                          `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type       *string                       `cty:"type" hcl:"type" column:"type,text" json:"-"`
	Display    *string                       `cty:"display" hcl:"display" json:"-"`

	Base *DashboardHierarchy `hcl:"base" json:"-"`

	parents []ModTreeItem
}

func NewDashboardHierarchy(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)

	h := &DashboardHierarchy{
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

func (h *DashboardHierarchy) Equals(other *DashboardHierarchy) bool {
	diff := h.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource
func (h *DashboardHierarchy) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	h.setBaseProperties()
	if len(h.Nodes) > 0 {
		h.NodeNames = h.Nodes.Names()
	}
	if len(h.Edges) > 0 {
		h.EdgeNames = h.Edges.Names()
	}
	return nil
}

// TODO [node_reuse] Add DashboardLeafNodeImpl and move this there https://github.com/turbot/steampipe/issues/2926

// GetChildren implements ModTreeItem
func (h *DashboardHierarchy) GetChildren() []ModTreeItem {
	// return nodes and edges (if any)
	children := make([]ModTreeItem, len(h.Nodes)+len(h.Edges))
	for i, n := range h.Nodes {
		children[i] = n
	}
	offset := len(h.Nodes)
	for i, e := range h.Edges {
		children[i+offset] = e
	}
	return children
}

func (h *DashboardHierarchy) Diff(other *DashboardHierarchy) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: h,
		Name: h.Name(),
	}

	if !utils.SafeStringsEqual(h.Type, other.Type) {
		res.AddPropertyDiff("Type")
	}

	if len(h.Categories) != len(other.Categories) {
		res.AddPropertyDiff("Categories")
	} else {
		for name, c := range h.Categories {
			if !c.Equals(other.Categories[name]) {
				res.AddPropertyDiff("Categories")
			}
		}
	}

	res.populateChildDiffs(h, other)
	res.queryProviderDiff(h, other)
	res.dashboardLeafNodeDiff(h, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (h *DashboardHierarchy) GetWidth() int {
	if h.Width == nil {
		return 0
	}
	return *h.Width
}

// GetDisplay implements DashboardLeafNode
func (h *DashboardHierarchy) GetDisplay() string {
	return typehelpers.SafeString(h.Display)
}

// GetDocumentation implements DashboardLeafNode, ModTreeItem
func (h *DashboardHierarchy) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (h *DashboardHierarchy) GetType() string {
	return typehelpers.SafeString(h.Type)
}

// GetEdges implements NodeAndEdgeProvider
func (h *DashboardHierarchy) GetEdges() DashboardEdgeList {
	return h.Edges
}

// GetNodes implements NodeAndEdgeProvider
func (h *DashboardHierarchy) GetNodes() DashboardNodeList {
	return h.Nodes
}

// SetEdges implements NodeAndEdgeProvider
func (h *DashboardHierarchy) SetEdges(edges DashboardEdgeList) {
	h.Edges = edges
}

// SetNodes implements NodeAndEdgeProvider
func (h *DashboardHierarchy) SetNodes(nodes DashboardNodeList) {
	h.Nodes = nodes
}

// AddCategory implements NodeAndEdgeProvider
func (h *DashboardHierarchy) AddCategory(category *DashboardCategory) hcl.Diagnostics {
	categoryName := category.ShortName
	if _, ok := h.Categories[categoryName]; ok {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s has duplicate category %s", h.Name(), categoryName),
			Subject:  category.GetDeclRange(),
		}}
	}
	h.Categories[categoryName] = category
	return nil
}

// AddChild implements NodeAndEdgeProvider
func (h *DashboardHierarchy) AddChild(child HclResource) hcl.Diagnostics {
	switch c := child.(type) {
	case *DashboardNode:
		h.Nodes = append(h.Nodes, c)
	case *DashboardEdge:
		h.Edges = append(h.Edges, c)
	default:
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("DashboardHierarchy does not support children of type %s", child.BlockType()),
			Subject:  h.GetDeclRange(),
		}}
	}
	// set ourselves as parent
	child.(ModTreeItem).AddParent(h)

	return nil
}

// CtyValue implements CtyValueProvider
func (h *DashboardHierarchy) CtyValue() (cty.Value, error) {
	return GetCtyValue(h)
}

func (h *DashboardHierarchy) setBaseProperties() {
	if h.Base == nil {
		return
	}
	// copy base into the HclResourceImpl 'base' property so it is accessible to all nested structs
	h.base = h.Base
	// call into parent nested struct setBaseProperties
	h.QueryProviderImpl.setBaseProperties()

	if h.Type == nil {
		h.Type = h.Base.Type
	}

	if h.Display == nil {
		h.Display = h.Base.Display
	}

	if h.Width == nil {
		h.Width = h.Base.Width
	}

	if h.Categories == nil {
		h.Categories = h.Base.Categories
	} else {
		h.Categories = utils.MergeMaps(h.Categories, h.Base.Categories)
	}

	if h.Edges == nil {
		h.Edges = h.Base.Edges
	} else {
		h.Edges.Merge(h.Base.Edges)
	}

	if h.Nodes == nil {
		h.Nodes = h.Base.Nodes
	} else {
		h.Nodes.Merge(h.Base.Nodes)
	}
}
