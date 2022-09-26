package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/constants"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// DashboardNode is a struct representing a leaf dashboard node
type DashboardNode struct {
	ResourceWithMetadataBase
	QueryProviderBase

	// required to allow partial decoding
	Remain          hcl.Body `hcl:",remain" json:"-"`
	FullName        string   `cty:"name" json:"-"`
	ShortName       string   `json:"-"`
	UnqualifiedName string   `json:"-"`

	CategoryList DashboardCategoryList         `cty:"category_list" hcl:"category,block" column:"category,jsonb" json:"-"`
	Categories   map[string]*DashboardCategory `cty:"categories" json:"categories"`

	// these properties are JSON serialised by the parent LeafRun
	Title *string `cty:"title" hcl:"title" column:"title,text" json:"-"`

	// QueryProvider
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"-"`
	Query                 *Query      `hcl:"query" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"-"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"-"`

	Base       *DashboardNode       `hcl:"base" json:"-"`
	DeclRange  hcl.Range            `json:"-"`
	References []*ResourceReference `json:"-"`
	Mod        *Mod                 `cty:"mod" json:"-"`
	Paths      []NodePath           `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewDashboardNode(block *hcl.Block, mod *Mod, shortName string) HclResource {
	c := &DashboardNode{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}

	c.SetAnonymous(block)
	return c
}

func (n *DashboardNode) Equals(other *DashboardNode) bool {
	diff := n.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (n *DashboardNode) CtyValue() (cty.Value, error) {
	return getCtyValue(n)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'edge.<shortName>'
func (n *DashboardNode) Name() string {
	return n.FullName
}

// OnDecoded implements HclResource
func (n *DashboardNode) OnDecoded(_ *hcl.Block, resourceMapProvider ModResourcesProvider) hcl.Diagnostics {
	n.setBaseProperties(resourceMapProvider)
	// populate categories map
	if len(n.CategoryList) > 0 {
		n.Categories = make(map[string]*DashboardCategory, len(n.CategoryList))
		for _, c := range n.CategoryList {
			n.Categories[c.Name] = c
		}
	}
	return nil
}

// AddReference implements HclResource
func (n *DashboardNode) AddReference(ref *ResourceReference) {
	n.References = append(n.References, ref)
}

// GetReferences implements HclResource
func (n *DashboardNode) GetReferences() []*ResourceReference {
	return n.References
}

// GetMod implements HclResource
func (n *DashboardNode) GetMod() *Mod {
	return n.Mod
}

// GetDeclRange implements HclResource
func (n *DashboardNode) GetDeclRange() *hcl.Range {
	return &n.DeclRange
}

// AddParent implements ModTreeItem
func (n *DashboardNode) AddParent(parent ModTreeItem) error {
	n.parents = append(n.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (n *DashboardNode) GetParents() []ModTreeItem {
	return n.parents
}

// GetChildren implements ModTreeItem
func (n *DashboardNode) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements ModTreeItem
func (n *DashboardNode) GetTitle() string {
	return typehelpers.SafeString(n.Title)
}

// GetDescription implements ModTreeItem
func (n *DashboardNode) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (n *DashboardNode) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (n *DashboardNode) GetPaths() []NodePath {
	// lazy load
	if len(n.Paths) == 0 {
		n.SetPaths()
	}

	return n.Paths
}

// SetPaths implements ModTreeItem
func (n *DashboardNode) SetPaths() {
	for _, parent := range n.parents {
		for _, parentPath := range parent.GetPaths() {
			n.Paths = append(n.Paths, append(parentPath, n.Name()))
		}
	}
}

func (n *DashboardNode) Diff(other *DashboardNode) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: n,
		Name: n.Name(),
	}

	if len(n.CategoryList) != len(other.CategoryList) {
		res.AddPropertyDiff("Categories")
	} else {
		for i, c := range n.Categories {
			if !c.Equals(other.Categories[i]) {
				res.AddPropertyDiff("Categories")
			}
		}
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

// GetDocumentation implements DashboardLeafNode
func (n *DashboardNode) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (n *DashboardNode) GetType() string {
	return ""
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (n *DashboardNode) GetUnqualifiedName() string {
	return n.UnqualifiedName
}

// GetParams implements QueryProvider
func (n *DashboardNode) GetParams() []*ParamDef {
	return n.Params
}

// GetArgs implements QueryProvider
func (n *DashboardNode) GetArgs() *QueryArgs {
	return n.Args
}

// GetSQL implements QueryProvider
func (n *DashboardNode) GetSQL() *string {
	return n.SQL
}

// GetQuery implements QueryProvider
func (n *DashboardNode) GetQuery() *Query {
	return n.Query
}

// SetArgs implements QueryProvider
func (n *DashboardNode) SetArgs(args *QueryArgs) {
	n.Args = args
}

// SetParams implements QueryProvider
func (n *DashboardNode) SetParams(params []*ParamDef) {
	n.Params = params
}

// GetPreparedStatementName implements QueryProvider
func (n *DashboardNode) GetPreparedStatementName() string {
	if n.PreparedStatementName != "" {
		return n.PreparedStatementName
	}
	n.PreparedStatementName = n.buildPreparedStatementName(n.ShortName, n.Mod.NameWithVersion(), constants.PreparedStatementChartSuffix)
	return n.PreparedStatementName
}

// GetPreparedStatementExecuteSQL implements QueryProvider
func (n *DashboardNode) GetPreparedStatementExecuteSQL(runtimeArgs *QueryArgs) (*ResolvedQuery, error) {
	// defer to base
	return n.getPreparedStatementExecuteSQL(n, runtimeArgs)
}

func (n *DashboardNode) setBaseProperties(resourceMapProvider ModResourcesProvider) {
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

	if n.CategoryList == nil {
		n.CategoryList = n.Base.CategoryList
	} else {
		n.CategoryList.Merge(n.Base.CategoryList)
	}

	n.MergeRuntimeDependencies(n.Base)
}
