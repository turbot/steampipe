package modconfig

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/constants"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

// DashboardFlow is a struct representing a leaf dashboard node
type DashboardFlow struct {
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

	Base       *DashboardFlow       `hcl:"base" json:"-"`
	DeclRange  hcl.Range            `json:"-"`
	References []*ResourceReference `json:"-"`
	Mod        *Mod                 `cty:"mod" json:"-"`
	Paths      []NodePath           `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewDashboardFlow(block *hcl.Block, mod *Mod, shortName string) HclResource {
	h := &DashboardFlow{
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

func (f *DashboardFlow) Equals(other *DashboardFlow) bool {
	diff := f.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (f *DashboardFlow) CtyValue() (cty.Value, error) {
	return getCtyValue(f)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'chart.<shortName>'
func (f *DashboardFlow) Name() string {
	return f.FullName
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

// GetMod implements ModTreeItem
func (f *DashboardFlow) GetMod() *Mod {
	return f.Mod
}

// GetDeclRange implements HclResource
func (f *DashboardFlow) GetDeclRange() *hcl.Range {
	return &f.DeclRange
}

// BlockType implements HclResource
func (*DashboardFlow) BlockType() string {
	return BlockTypeFlow
}

// AddParent implements ModTreeItem
func (f *DashboardFlow) AddParent(parent ModTreeItem) error {
	f.parents = append(f.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (f *DashboardFlow) GetParents() []ModTreeItem {
	return f.parents
}

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

// GetTitle implements HclResource
func (f *DashboardFlow) GetTitle() string {
	return typehelpers.SafeString(f.Title)
}

// GetDescription implements ModTreeItem
func (f *DashboardFlow) GetDescription() string {
	return ""
}

// GetTags implements HclResource
func (f *DashboardFlow) GetTags() map[string]string {
	return map[string]string{}
}

// GetPaths implements ModTreeItem
func (f *DashboardFlow) GetPaths() []NodePath {
	// lazy load
	if len(f.Paths) == 0 {
		f.SetPaths()
	}

	return f.Paths
}

// SetPaths implements ModTreeItem
func (f *DashboardFlow) SetPaths() {
	for _, parent := range f.parents {
		for _, parentPath := range parent.GetPaths() {
			f.Paths = append(f.Paths, append(parentPath, f.Name()))
		}
	}
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

// GetDocumentation implements DashboardLeafNode, ModTreeItem
func (f *DashboardFlow) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (f *DashboardFlow) GetType() string {
	return typehelpers.SafeString(f.Type)
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (f *DashboardFlow) GetUnqualifiedName() string {
	return f.UnqualifiedName
}

// GetParams implements QueryProvider
func (f *DashboardFlow) GetParams() []*ParamDef {
	return f.Params
}

// GetArgs implements QueryProvider
func (f *DashboardFlow) GetArgs() *QueryArgs {
	return f.Args
}

// GetSQL implements QueryProvider
func (f *DashboardFlow) GetSQL() *string {
	return f.SQL
}

// GetQuery implements QueryProvider
func (f *DashboardFlow) GetQuery() *Query {
	return f.Query
}

// VerifyQuery implements QueryProvider
func (*DashboardFlow) VerifyQuery(QueryProvider) error {
	// query is optional - nothing to do
	return nil
}

// SetArgs implements QueryProvider
func (f *DashboardFlow) SetArgs(args *QueryArgs) {
	f.Args = args
}

// SetParams implements QueryProvider
func (f *DashboardFlow) SetParams(params []*ParamDef) {
	f.Params = params
}

// GetPreparedStatementName implements QueryProvider
func (f *DashboardFlow) GetPreparedStatementName() string {
	if f.PreparedStatementName != "" {
		return f.PreparedStatementName
	}
	f.PreparedStatementName = f.buildPreparedStatementName(f.ShortName, f.Mod.NameWithVersion(), constants.PreparedStatementFlowSuffix)
	return f.PreparedStatementName
}

// GetResolvedQuery implements QueryProvider
func (f *DashboardFlow) GetResolvedQuery(runtimeArgs *QueryArgs) (*ResolvedQuery, error) {
	// defer to base
	return f.getResolvedQuery(f, runtimeArgs)
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
