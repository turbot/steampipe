package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/constants"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// DashboardEdge is a struct representing a leaf dashboard node
type DashboardEdge struct {
	ResourceWithMetadataBase
	QueryProviderBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	FullName        string             `cty:"name" json:"name"`
	ShortName       string             `json:"-"`
	UnqualifiedName string             `json:"-"`
	Category        *DashboardCategory `cty:"category" hcl:"category" column:"category,jsonb" json:"category,omitempty"`

	// these properties are JSON serialised by the parent LeafRun
	Title *string `cty:"title" hcl:"title" column:"title,text" json:"-"`
	// QueryProvider
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"-"`
	Query                 *Query      `hcl:"query" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"-"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"-"`

	Base       *DashboardEdge       `hcl:"base" json:"-"`
	DeclRange  hcl.Range            `json:"-"`
	References []*ResourceReference `json:"-"`
	Mod        *Mod                 `cty:"mod" json:"-"`
	Paths      []NodePath           `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewDashboardEdge(block *hcl.Block, mod *Mod, shortName string) HclResource {
	c := &DashboardEdge{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}

	c.SetAnonymous(block)
	return c
}

func (e *DashboardEdge) Equals(other *DashboardEdge) bool {
	diff := e.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (e *DashboardEdge) CtyValue() (cty.Value, error) {
	return getCtyValue(e)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'edge.<shortName>'
func (e *DashboardEdge) Name() string {
	return e.FullName
}

// OnDecoded implements HclResource
func (e *DashboardEdge) OnDecoded(_ *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	e.setBaseProperties(resourceMapProvider)

	// when we reference resources (i.e. category),
	// not all properties are retrieved as they are no cty serialisable
	// repopulate category from resourceMapProvider
	if e.Category != nil {
		fullCategory, diags := enrichCategory(e.Category, e, resourceMapProvider)
		if diags.HasErrors() {
			return diags
		}
		e.Category = fullCategory
	}
	return nil
}

// AddReference implements ResourceWithMetadata
func (e *DashboardEdge) AddReference(ref *ResourceReference) {
	e.References = append(e.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (e *DashboardEdge) GetReferences() []*ResourceReference {
	return e.References
}

// GetMod implements ModTreeItem
func (e *DashboardEdge) GetMod() *Mod {
	return e.Mod
}

// GetDeclRange implements HclResource
func (e *DashboardEdge) GetDeclRange() *hcl.Range {
	return &e.DeclRange
}

// BlockType implements HclResource
func (*DashboardEdge) BlockType() string {
	return BlockTypeEdge
}

// AddParent implements ModTreeItem
func (e *DashboardEdge) AddParent(parent ModTreeItem) error {
	e.parents = append(e.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (e *DashboardEdge) GetParents() []ModTreeItem {
	return e.parents
}

// GetChildren implements ModTreeItem
func (e *DashboardEdge) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements HclResource
func (e *DashboardEdge) GetTitle() string {
	return typehelpers.SafeString(e.Title)
}

// GetDescription implements ModTreeItem
func (e *DashboardEdge) GetDescription() string {
	return ""
}

// GetTags implements HclResource
func (e *DashboardEdge) GetTags() map[string]string {
	return map[string]string{}
}

// GetPaths implements ModTreeItem
func (e *DashboardEdge) GetPaths() []NodePath {
	// lazy load
	if len(e.Paths) == 0 {
		e.SetPaths()
	}

	return e.Paths
}

// SetPaths implements ModTreeItem
func (e *DashboardEdge) SetPaths() {
	for _, parent := range e.parents {
		for _, parentPath := range parent.GetPaths() {
			e.Paths = append(e.Paths, append(parentPath, e.Name()))
		}
	}
}

func (e *DashboardEdge) Diff(other *DashboardEdge) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: e,
		Name: e.Name(),
	}
	if (e.Category == nil) != (other.Category == nil) {
		res.AddPropertyDiff("Category")
	}

	if e.Category != nil && !e.Category.Equals(other.Category) {
		res.AddPropertyDiff("Category")
	}

	res.populateChildDiffs(e, other)
	res.queryProviderDiff(e, other)
	res.dashboardLeafNodeDiff(e, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (e *DashboardEdge) GetWidth() int {
	return 0
}

// GetDisplay implements DashboardLeafNode
func (e *DashboardEdge) GetDisplay() string {
	return ""
}

// GetDocumentation implements DashboardLeafNode
func (e *DashboardEdge) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (e *DashboardEdge) GetType() string {
	return ""
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (e *DashboardEdge) GetUnqualifiedName() string {
	return e.UnqualifiedName
}

// GetParams implements QueryProvider
func (e *DashboardEdge) GetParams() []*ParamDef {
	return e.Params
}

// GetArgs implements QueryProvider
func (e *DashboardEdge) GetArgs() *QueryArgs {
	return e.Args
}

// GetSQL implements QueryProvider
func (e *DashboardEdge) GetSQL() *string {
	return e.SQL
}

// GetQuery implements QueryProvider
func (e *DashboardEdge) GetQuery() *Query {
	return e.Query
}

// SetArgs implements QueryProvider
func (e *DashboardEdge) SetArgs(args *QueryArgs) {
	e.Args = args
}

// SetParams implements QueryProvider
func (e *DashboardEdge) SetParams(params []*ParamDef) {
	e.Params = params
}

// GetPreparedStatementName implements QueryProvider
func (e *DashboardEdge) GetPreparedStatementName() string {
	if e.PreparedStatementName != "" {
		return e.PreparedStatementName
	}
	e.PreparedStatementName = e.buildPreparedStatementName(e.ShortName, e.Mod.NameWithVersion(), constants.PreparedStatementChartSuffix)
	return e.PreparedStatementName
}

// GetResolvedQuery implements QueryProvider
func (e *DashboardEdge) GetResolvedQuery(runtimeArgs *QueryArgs) (*ResolvedQuery, error) {
	// defer to base
	return e.getResolvedQuery(e, runtimeArgs)
}

// IsSnapshotPanel implements SnapshotPanel
func (*DashboardEdge) IsSnapshotPanel() {}

func (e *DashboardEdge) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(e.Base, resourceMapProvider); !resolved {
		return
	} else {
		e.Base = base.(*DashboardEdge)
	}

	if e.Title == nil {
		e.Title = e.Base.Title
	}

	if e.SQL == nil {
		e.SQL = e.Base.SQL
	}

	if e.Query == nil {
		e.Query = e.Base.Query
	}

	if e.Args == nil {
		e.Args = e.Base.Args
	}

	if e.Params == nil {
		e.Params = e.Base.Params
	}

	if e.Category == nil {
		e.Category = e.Base.Category
	}

	e.MergeRuntimeDependencies(e.Base)
}
