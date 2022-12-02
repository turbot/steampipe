package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/constants"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

// DashboardHierarchy is a struct representing a leaf dashboard node
type DashboardHierarchy struct {
	ResourceWithMetadataBase
	QueryProviderBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	Nodes DashboardNodeList `cty:"node_list"  hcl:"nodes,optional" column:"nodes,jsonb" json:"-"`
	Edges DashboardEdgeList `cty:"edge_list" hcl:"edges,optional" column:"edges,jsonb" json:"-"`

	Categories map[string]*DashboardCategory `cty:"categories" json:"categories"`

	// these properties are JSON serialised by the parent LeafRun
	Title   *string `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width   *int    `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type    *string `cty:"type" hcl:"type" column:"type,text" json:"-"`
	Display *string `cty:"display" hcl:"display" json:"-"`

	// QueryProvider
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"-"`
	Query                 *Query      `hcl:"query" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"-"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"-"`

	Base       *DashboardHierarchy  `hcl:"base" json:"-"`
	DeclRange  hcl.Range            `json:"-"`
	References []*ResourceReference `json:"-"`
	Mod        *Mod                 `cty:"mod" json:"-"`
	Paths      []NodePath           `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewDashboardHierarchy(block *hcl.Block, mod *Mod, shortName string) HclResource {
	h := &DashboardHierarchy{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
		Categories:      make(map[string]*DashboardCategory),
	}
	h.SetAnonymous(block)
	return h
}

func (h *DashboardHierarchy) Equals(other *DashboardHierarchy) bool {
	diff := h.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (h *DashboardHierarchy) CtyValue() (cty.Value, error) {
	return getCtyValue(h)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'chart.<shortName>'
func (h *DashboardHierarchy) Name() string {
	return h.FullName
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

// GetMod implements ModTreeItem
func (h *DashboardHierarchy) GetMod() *Mod {
	return h.Mod
}

// GetDeclRange implements HclResource
func (h *DashboardHierarchy) GetDeclRange() *hcl.Range {
	return &h.DeclRange
}

// BlockType implements HclResource
func (*DashboardHierarchy) BlockType() string {
	return BlockTypeHierarchy
}

// AddParent implements ModTreeItem
func (h *DashboardHierarchy) AddParent(parent ModTreeItem) error {
	h.parents = append(h.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (h *DashboardHierarchy) GetParents() []ModTreeItem {
	return h.parents
}

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

// GetTitle implements HclResource
func (h *DashboardHierarchy) GetTitle() string {
	return typehelpers.SafeString(h.Title)
}

// GetDescription implements ModTreeItem
func (h *DashboardHierarchy) GetDescription() string {
	return ""
}

// GetTags implements HclResource
func (h *DashboardHierarchy) GetTags() map[string]string {
	return map[string]string{}
}

// GetPaths implements ModTreeItem
func (h *DashboardHierarchy) GetPaths() []NodePath {
	// lazy load
	if len(h.Paths) == 0 {
		h.SetPaths()
	}

	return h.Paths
}

// SetPaths implements ModTreeItem
func (h *DashboardHierarchy) SetPaths() {
	for _, parent := range h.parents {
		for _, parentPath := range parent.GetPaths() {
			h.Paths = append(h.Paths, append(parentPath, h.Name()))
		}
	}
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

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (h *DashboardHierarchy) GetUnqualifiedName() string {
	return h.UnqualifiedName
}

// GetParams implements QueryProvider
func (h *DashboardHierarchy) GetParams() []*ParamDef {
	return h.Params
}

// GetArgs implements QueryProvider
func (h *DashboardHierarchy) GetArgs() *QueryArgs {
	return h.Args
}

// GetSQL implements QueryProvider
func (h *DashboardHierarchy) GetSQL() *string {
	return h.SQL
}

// GetQuery implements QueryProvider
func (h *DashboardHierarchy) GetQuery() *Query {
	return h.Query
}

// VerifyQuery implements QueryProvider
func (*DashboardHierarchy) VerifyQuery(QueryProvider) error {
	// query is optional - nothing to do
	return nil
}

// SetArgs implements QueryProvider
func (h *DashboardHierarchy) SetArgs(args *QueryArgs) {
	h.Args = args
}

// SetParams implements QueryProvider
func (h *DashboardHierarchy) SetParams(params []*ParamDef) {
	h.Params = params
}

// GetPreparedStatementName implements QueryProvider
func (h *DashboardHierarchy) GetPreparedStatementName() string {
	if h.PreparedStatementName != "" {
		return h.PreparedStatementName
	}
	h.PreparedStatementName = h.buildPreparedStatementName(h.ShortName, h.Mod.NameWithVersion(), constants.PreparedStatementHierarchySuffix)
	return h.PreparedStatementName
}

// GetResolvedQuery implements QueryProvider
func (h *DashboardHierarchy) GetResolvedQuery(runtimeArgs *QueryArgs) (*ResolvedQuery, error) {
	// defer to base
	return h.getResolvedQuery(h, runtimeArgs)
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
