package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/constants"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

// DashboardGraph is a struct representing a leaf dashboard node
type DashboardGraph struct {
	ResourceWithMetadataBase
	QueryProviderBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	CategoryList DashboardGraphCategoryList         `cty:"category_list" hcl:"category,block" column:"category,jsonb" json:"-"`
	Categories   map[string]*DashboardGraphCategory `cty:"categories" json:"categories"`

	Direction *string `cty:"direction" hcl:"direction" column:"direction,text" json:"direction"`

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

	Base       *DashboardGraph      `hcl:"base" json:"-"`
	DeclRange  hcl.Range            `json:"-"`
	References []*ResourceReference `json:"-"`
	Mod        *Mod                 `cty:"mod" json:"-"`
	Paths      []NodePath           `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewDashboardGraph(block *hcl.Block, mod *Mod, shortName string) *DashboardGraph {
	h := &DashboardGraph{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}
	h.SetAnonymous(block)
	return h
}

func (g *DashboardGraph) Equals(other *DashboardGraph) bool {
	diff := g.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (g *DashboardGraph) CtyValue() (cty.Value, error) {
	return getCtyValue(g)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'chart.<shortName>'
func (g *DashboardGraph) Name() string {
	return g.FullName
}

// OnDecoded implements HclResource
func (g *DashboardGraph) OnDecoded(block *hcl.Block, resourceMapProvider ModResourcesProvider) hcl.Diagnostics {
	g.setBaseProperties(resourceMapProvider)
	// populate categories map
	if len(g.CategoryList) > 0 {
		g.Categories = make(map[string]*DashboardGraphCategory, len(g.CategoryList))
		for _, c := range g.CategoryList {
			g.Categories[c.Name] = c
		}
	}
	return nil
}

// AddReference implements HclResource
func (g *DashboardGraph) AddReference(ref *ResourceReference) {
	g.References = append(g.References, ref)
}

// GetReferences implements HclResource
func (g *DashboardGraph) GetReferences() []*ResourceReference {
	return g.References
}

// GetMod implements HclResource
func (g *DashboardGraph) GetMod() *Mod {
	return g.Mod
}

// GetDeclRange implements HclResource
func (g *DashboardGraph) GetDeclRange() *hcl.Range {
	return &g.DeclRange
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
	return nil
}

// GetTitle implements ModTreeItem
func (g *DashboardGraph) GetTitle() string {
	return typehelpers.SafeString(g.Title)
}

// GetDescription implements ModTreeItem
func (g *DashboardGraph) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (g *DashboardGraph) GetTags() map[string]string {
	return nil
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

	if len(g.CategoryList) != len(other.CategoryList) {
		res.AddPropertyDiff("Categories")
	} else {
		for i, c := range g.Categories {
			if !c.Equals(other.Categories[i]) {
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

// GetDocumentation implements DashboardLeafNode
func (g *DashboardGraph) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (g *DashboardGraph) GetType() string {
	return typehelpers.SafeString(g.Type)
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (g *DashboardGraph) GetUnqualifiedName() string {
	return g.UnqualifiedName
}

// GetParams implements QueryProvider
func (g *DashboardGraph) GetParams() []*ParamDef {
	return g.Params
}

// GetArgs implements QueryProvider
func (g *DashboardGraph) GetArgs() *QueryArgs {
	return g.Args
}

// GetSQL implements QueryProvider
func (g *DashboardGraph) GetSQL() *string {
	return g.SQL
}

// GetQuery implements QueryProvider
func (g *DashboardGraph) GetQuery() *Query {
	return g.Query
}

// SetArgs implements QueryProvider
func (g *DashboardGraph) SetArgs(args *QueryArgs) {
	g.Args = args
}

// SetParams implements QueryProvider
func (g *DashboardGraph) SetParams(params []*ParamDef) {
	g.Params = params
}

// GetPreparedStatementName implements QueryProvider
func (g *DashboardGraph) GetPreparedStatementName() string {
	if g.PreparedStatementName != "" {
		return g.PreparedStatementName
	}
	g.PreparedStatementName = g.buildPreparedStatementName(g.ShortName, g.Mod.NameWithVersion(), constants.PreparedStatementGraphSuffix)
	return g.PreparedStatementName
}

// GetPreparedStatementExecuteSQL implements QueryProvider
func (g *DashboardGraph) GetPreparedStatementExecuteSQL(runtimeArgs *QueryArgs) (*ResolvedQuery, error) {
	// defer to base
	return g.getPreparedStatementExecuteSQL(g, runtimeArgs)
}

func (g *DashboardGraph) setBaseProperties(resourceMapProvider ModResourcesProvider) {
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

	if g.CategoryList == nil {
		g.CategoryList = g.Base.CategoryList
	} else {
		g.CategoryList.Merge(g.Base.CategoryList)
	}

	if g.Direction == nil {
		g.Direction = g.Base.Direction
	}

	g.MergeRuntimeDependencies(g.Base)
}
