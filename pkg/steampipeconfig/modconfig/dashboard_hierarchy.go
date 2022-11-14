package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
)

// DashboardHierarchy is a struct representing a leaf dashboard node
type DashboardHierarchy struct {
	ResourceWithMetadataBase
	QueryProviderBase
	ModTreeItemBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	Nodes DashboardNodeList `cty:"node_list"  hcl:"nodes,optional" column:"nodes,jsonb" json:"-"`
	Edges DashboardEdgeList `cty:"edge_list" hcl:"edges,optional" column:"edges,jsonb" json:"-"`

	Categories map[string]*DashboardCategory `cty:"categories" json:"categories"`
	Width      *int                          `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type       *string                       `cty:"type" hcl:"type" column:"type,text" json:"-"`
	Display    *string                       `cty:"display" hcl:"display" json:"-"`

	Base       *DashboardHierarchy  `hcl:"base" json:"-"`
	References []*ResourceReference `json:"-"`
	Paths      []NodePath           `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewDashboardHierarchy(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)

	h := &DashboardHierarchy{
		Categories: make(map[string]*DashboardCategory),
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
	h.SetAnonymous(block)
	return h
}

func (h *DashboardHierarchy) Equals(other *DashboardHierarchy) bool {
	diff := h.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource
func (h *DashboardHierarchy) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	h.setBaseProperties(resourceMapProvider)

	// populate nodes and edges
	return initialiseEdgesAndNodes(h, resourceMapProvider)
}

// AddReference implements ResourceWithMetadata
func (h *DashboardHierarchy) AddReference(ref *ResourceReference) {
	h.References = append(h.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (h *DashboardHierarchy) GetReferences() []*ResourceReference {
	return h.References
}

// TODO KAI PUT IN 1 PLACE FOR ALL EDGE PROVIDERS
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

// GetEdges implements EdgeAndNodeProvider
func (h *DashboardHierarchy) GetEdges() DashboardEdgeList {
	return h.Edges
}

// GetNodes implements EdgeAndNodeProvider
func (h *DashboardHierarchy) GetNodes() DashboardNodeList {
	return h.Nodes
}

// SetEdges implements EdgeAndNodeProvider
func (h *DashboardHierarchy) SetEdges(edges DashboardEdgeList) {
	h.Edges = edges
}

// SetNodes implements EdgeAndNodeProvider
func (h *DashboardHierarchy) SetNodes(nodes DashboardNodeList) {
	h.Nodes = nodes
}

// AddCategory implements EdgeAndNodeProvider
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

func (h *DashboardHierarchy) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(h.Base, resourceMapProvider); !resolved {
		return
	} else {
		h.Base = base.(*DashboardHierarchy)
	}

	if h.Title == nil {
		h.Title = h.Base.Title
	}

	if h.Type == nil {
		h.Type = h.Base.Type
	}

	if h.Display == nil {
		h.Display = h.Base.Display
	}

	if h.Width == nil {
		h.Width = h.Base.Width
	}

	if h.SQL == nil {
		h.SQL = h.Base.SQL
	}

	if h.Query == nil {
		h.Query = h.Base.Query
	}

	if h.Args == nil {
		h.Args = h.Base.Args
	}

	if h.Params == nil {
		h.Params = h.Base.Params
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

	h.MergeRuntimeDependencies(h.Base)
}
