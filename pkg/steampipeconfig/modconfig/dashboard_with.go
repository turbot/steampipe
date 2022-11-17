package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/constants"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// DashboardWith is a struct representing a leaf dashboard node
type DashboardWith struct {
	ResourceWithMetadataBase
	QueryProviderBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	FullName        string `cty:"name" json:"name"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	// these properties are JSON serialised by the parent LeafRun
	Title *string `cty:"title" hcl:"title" column:"title,text" json:"-"`
	// QueryProvider
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"-"`
	Query                 *Query      `hcl:"query" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"-"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"-"`

	Base       *DashboardWith       `hcl:"base" json:"-"`
	DeclRange  hcl.Range            `json:"-"`
	References []*ResourceReference `json:"-"`
	Mod        *Mod                 `cty:"mod" json:"-"`
	Paths      []NodePath           `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewDashboardWith(block *hcl.Block, mod *Mod, shortName string) HclResource {
	// with blocks cannot be anonymous
	c := &DashboardWith{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}

	return c
}

func (e *DashboardWith) Equals(other *DashboardWith) bool {
	diff := e.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (e *DashboardWith) CtyValue() (cty.Value, error) {
	return getCtyValue(e)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'with.<shortName>'
func (e *DashboardWith) Name() string {
	return e.FullName
}

// OnDecoded implements HclResource
func (e *DashboardWith) OnDecoded(_ *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	e.setBaseProperties(resourceMapProvider)

	return nil
}

// AddReference implements ResourceWithMetadata
func (e *DashboardWith) AddReference(ref *ResourceReference) {
	e.References = append(e.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (e *DashboardWith) GetReferences() []*ResourceReference {
	return e.References
}

// GetMod implements ModTreeItem
func (e *DashboardWith) GetMod() *Mod {
	return e.Mod
}

// GetDeclRange implements HclResource
func (e *DashboardWith) GetDeclRange() *hcl.Range {
	return &e.DeclRange
}

// BlockType implements HclResource
func (*DashboardWith) BlockType() string {
	return BlockTypeQuery
}

// AddParent implements ModTreeItem
func (e *DashboardWith) AddParent(parent ModTreeItem) error {
	e.parents = append(e.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (e *DashboardWith) GetParents() []ModTreeItem {
	return e.parents
}

// GetChildren implements ModTreeItem
func (e *DashboardWith) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements HclResource
func (e *DashboardWith) GetTitle() string {
	return typehelpers.SafeString(e.Title)
}

// GetDescription implements ModTreeItem
func (e *DashboardWith) GetDescription() string {
	return ""
}

// GetTags implements HclResource
func (e *DashboardWith) GetTags() map[string]string {
	return map[string]string{}
}

// GetPaths implements ModTreeItem
func (e *DashboardWith) GetPaths() []NodePath {
	// lazy load
	if len(e.Paths) == 0 {
		e.SetPaths()
	}

	return e.Paths
}

// SetPaths implements ModTreeItem
func (e *DashboardWith) SetPaths() {
	for _, parent := range e.parents {
		for _, parentPath := range parent.GetPaths() {
			e.Paths = append(e.Paths, append(parentPath, e.Name()))
		}
	}
}

func (e *DashboardWith) Diff(other *DashboardWith) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: e,
		Name: e.Name(),
	}

	res.queryProviderDiff(e, other)

	return res
}

// GetDocumentation implements ModTreeItem
func (e *DashboardWith) GetDocumentation() string {
	return ""
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (e *DashboardWith) GetUnqualifiedName() string {
	return e.UnqualifiedName
}

// GetParams implements QueryProvider
func (e *DashboardWith) GetParams() []*ParamDef {
	return e.Params
}

// GetArgs implements QueryProvider
func (e *DashboardWith) GetArgs() *QueryArgs {
	return e.Args
}

// GetSQL implements QueryProvider
func (e *DashboardWith) GetSQL() *string {
	return e.SQL
}

// GetQuery implements QueryProvider
func (e *DashboardWith) GetQuery() *Query {
	return e.Query
}

// SetArgs implements QueryProvider
func (e *DashboardWith) SetArgs(args *QueryArgs) {
	e.Args = args
}

// SetParams implements QueryProvider
func (e *DashboardWith) SetParams(params []*ParamDef) {
	e.Params = params
}

// GetPreparedStatementName implements QueryProvider
func (e *DashboardWith) GetPreparedStatementName() string {
	if e.PreparedStatementName != "" {
		return e.PreparedStatementName
	}
	e.PreparedStatementName = e.buildPreparedStatementName(e.ShortName, e.Mod.NameWithVersion(), constants.PreparedStatementChartSuffix)
	return e.PreparedStatementName
}

// GetResolvedQuery implements QueryProvider
func (e *DashboardWith) GetResolvedQuery(runtimeArgs *QueryArgs) (*ResolvedQuery, error) {
	// defer to base
	return e.getResolvedQuery(e, runtimeArgs)
}

// IsSnapshotPanel implements SnapshotPanel
func (*DashboardWith) IsSnapshotPanel() {}

// GetWidth implements DashboardLeafNode
func (*DashboardWith) GetWidth() int {
	return 0
}

// GetDisplay implements DashboardLeafNode
func (*DashboardWith) GetDisplay() string {
	return ""
}

// GetType implements DashboardLeafNode
func (*DashboardWith) GetType() string {
	return ""
}

func (e *DashboardWith) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(e.Base, resourceMapProvider); !resolved {
		return
	} else {
		e.Base = base.(*DashboardWith)
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

	e.MergeRuntimeDependencies(e.Base)
}
