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

	CategoryList DashboardFlowCategoryList         `cty:"category_list" hcl:"category,block" column:"category,jsonb" json:"-"`
	Categories   map[string]*DashboardFlowCategory `cty:"categories" json:"categories"`

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

func NewDashboardFlow(block *hcl.Block, mod *Mod, shortName string) *DashboardFlow {
	h := &DashboardFlow{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
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
func (f *DashboardFlow) OnDecoded(block *hcl.Block, resourceMapProvider ModResourcesProvider) hcl.Diagnostics {
	f.setBaseProperties(resourceMapProvider)
	// populate categories map
	if len(f.CategoryList) > 0 {
		f.Categories = make(map[string]*DashboardFlowCategory, len(f.CategoryList))
		for _, c := range f.CategoryList {
			f.Categories[c.Name] = c
		}
	}
	return nil
}

// AddReference implements HclResource
func (f *DashboardFlow) AddReference(ref *ResourceReference) {
	f.References = append(f.References, ref)
}

// GetReferences implements HclResource
func (f *DashboardFlow) GetReferences() []*ResourceReference {
	return f.References
}

// GetMod implements HclResource
func (f *DashboardFlow) GetMod() *Mod {
	return f.Mod
}

// GetDeclRange implements HclResource
func (f *DashboardFlow) GetDeclRange() *hcl.Range {
	return &f.DeclRange
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
	return nil
}

// GetTitle implements ModTreeItem
func (f *DashboardFlow) GetTitle() string {
	return typehelpers.SafeString(f.Title)
}

// GetDescription implements ModTreeItem
func (f *DashboardFlow) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (f *DashboardFlow) GetTags() map[string]string {
	return nil
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

	if len(f.CategoryList) != len(other.CategoryList) {
		res.AddPropertyDiff("Categories")
	} else {
		for i, c := range f.Categories {
			if !c.Equals(other.Categories[i]) {
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

// GetDocumentation implements DashboardLeafNode
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

// GetPreparedStatementExecuteSQL implements QueryProvider
func (f *DashboardFlow) GetPreparedStatementExecuteSQL(runtimeArgs *QueryArgs) (*ResolvedQuery, error) {
	// defer to base
	return f.getPreparedStatementExecuteSQL(f, runtimeArgs)
}

func (f *DashboardFlow) setBaseProperties(resourceMapProvider ModResourcesProvider) {
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

	if f.CategoryList == nil {
		f.CategoryList = f.Base.CategoryList
	} else {
		f.CategoryList.Merge(f.Base.CategoryList)
	}

	f.MergeRuntimeDependencies(f.Base)
}
