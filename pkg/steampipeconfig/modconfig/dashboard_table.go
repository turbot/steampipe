package modconfig

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/constants"

	"github.com/turbot/steampipe/pkg/utils"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// DashboardTable is a struct representing a leaf dashboard node
type DashboardTable struct {
	ResourceWithMetadataBase
	QueryProviderBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	Title      *string                          `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width      *int                             `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type       *string                          `cty:"type" hcl:"type" column:"type,text" json:"-"`
	ColumnList DashboardTableColumnList         `cty:"column_list" hcl:"column,block" column:"columns,jsonb" json:"-"`
	Columns    map[string]*DashboardTableColumn `cty:"columns" json:"columns,omitempty"`
	Display    *string                          `cty:"display" hcl:"display" json:"display,omitempty"`

	// QueryProvider
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"-"`
	Query                 *Query      `hcl:"query" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb"json:"-"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"-"`

	Base       *DashboardTable      `hcl:"base" json:"-"`
	DeclRange  hcl.Range            `json:"-"`
	References []*ResourceReference `json:"-"`
	Mod        *Mod                 `cty:"mod" json:"-"`
	Paths      []NodePath           `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewDashboardTable(block *hcl.Block, mod *Mod, shortName string) HclResource {
	t := &DashboardTable{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}
	t.SetAnonymous(block)
	return t
}

// NewQueryDashboardTable creates a Table to wrap a query.
// This is used in order to execute queries as dashboards
func NewQueryDashboardTable(q ModTreeItem) (*DashboardTable, error) {
	parsedName, err := ParseResourceName(q.Name())
	if err != nil {
		return nil, err
	}

	queryProvider, ok := q.(QueryProvider)
	if !ok {
		return nil, fmt.Errorf("rersource passed to NewQueryDashboardTable must implement QueryProvider")
	}

	tableName := BuildFullResourceName(q.GetMod().ShortName, BlockTypeTable, parsedName.Name)
	c := &DashboardTable{
		ResourceWithMetadataBase: ResourceWithMetadataBase{
			metadata: &ResourceMetadata{},
		},
		ShortName:       parsedName.Name,
		FullName:        tableName,
		UnqualifiedName: fmt.Sprintf("%s.%s", BlockTypeTable, parsedName),
		Title:           utils.ToStringPointer(q.GetTitle()),
		Mod:             q.GetMod(),
		Query:           queryProvider.GetQuery(),
		SQL:             queryProvider.GetSQL(),
	}
	return c, nil
}

func (t *DashboardTable) Equals(other *DashboardTable) bool {
	diff := t.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (t *DashboardTable) CtyValue() (cty.Value, error) {
	return getCtyValue(t)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'table.<shortName>'
func (t *DashboardTable) Name() string {
	return t.FullName
}

// OnDecoded implements HclResource
func (t *DashboardTable) OnDecoded(_ *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	t.setBaseProperties(resourceMapProvider)
	// populate columns map
	if len(t.ColumnList) > 0 {
		t.Columns = make(map[string]*DashboardTableColumn, len(t.ColumnList))
		for _, c := range t.ColumnList {
			t.Columns[c.Name] = c
		}
	}
	return nil
}

// AddReference implements ResourceWithMetadata
func (t *DashboardTable) AddReference(ref *ResourceReference) {
	t.References = append(t.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (t *DashboardTable) GetReferences() []*ResourceReference {
	return t.References
}

// GetMod implements ModTreeItem
func (t *DashboardTable) GetMod() *Mod {
	return t.Mod
}

// GetDeclRange implements HclResource
func (t *DashboardTable) GetDeclRange() *hcl.Range {
	return &t.DeclRange
}

// BlockType implements HclResource
func (*DashboardTable) BlockType() string {
	return BlockTypeTable
}

// AddParent implements ModTreeItem
func (t *DashboardTable) AddParent(parent ModTreeItem) error {
	t.parents = append(t.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (t *DashboardTable) GetParents() []ModTreeItem {
	return t.parents
}

// GetChildren implements ModTreeItem
func (t *DashboardTable) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements HclResource, DashboardLeafNode
func (t *DashboardTable) GetTitle() string {
	return typehelpers.SafeString(t.Title)
}

// GetDescription implements ModTreeItem
func (t *DashboardTable) GetDescription() string {
	return ""
}

// GetTags implements HclResource
func (t *DashboardTable) GetTags() map[string]string {
	return map[string]string{}
}

// GetPaths implements ModTreeItem
func (t *DashboardTable) GetPaths() []NodePath {
	// lazy load
	if len(t.Paths) == 0 {
		t.SetPaths()
	}

	return t.Paths
}

// SetPaths implements ModTreeItem
func (t *DashboardTable) SetPaths() {
	for _, parent := range t.parents {
		for _, parentPath := range parent.GetPaths() {
			t.Paths = append(t.Paths, append(parentPath, t.Name()))
		}
	}
}

func (t *DashboardTable) Diff(other *DashboardTable) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: t,
		Name: t.Name(),
	}

	if !utils.SafeStringsEqual(t.Type, other.Type) {
		res.AddPropertyDiff("Type")
	}

	if len(t.ColumnList) != len(other.ColumnList) {
		res.AddPropertyDiff("Columns")
	} else {
		for i, c := range t.Columns {
			if !c.Equals(other.Columns[i]) {
				res.AddPropertyDiff("Columns")
			}
		}
	}

	res.populateChildDiffs(t, other)
	res.queryProviderDiff(t, other)
	res.dashboardLeafNodeDiff(t, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (t *DashboardTable) GetWidth() int {
	if t.Width == nil {
		return 0
	}
	return *t.Width
}

// GetDisplay implements DashboardLeafNode
func (t *DashboardTable) GetDisplay() string {
	return typehelpers.SafeString(t.Display)
}

// GetDocumentation implements DashboardLeafNode, ModTreeItem
func (*DashboardTable) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (t *DashboardTable) GetType() string {
	return typehelpers.SafeString(t.Type)
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (t *DashboardTable) GetUnqualifiedName() string {
	return t.UnqualifiedName
}

// GetParams implements QueryProvider
func (t *DashboardTable) GetParams() []*ParamDef {
	return t.Params
}

// GetArgs implements QueryProvider
func (t *DashboardTable) GetArgs() *QueryArgs {
	return t.Args
}

// GetSQL implements QueryProvider
func (t *DashboardTable) GetSQL() *string {
	return t.SQL
}

// GetQuery implements QueryProvider
func (t *DashboardTable) GetQuery() *Query {
	return t.Query
}

// SetArgs implements QueryProvider
func (t *DashboardTable) SetArgs(args *QueryArgs) {
	t.Args = args
}

// SetParams implements QueryProvider
func (t *DashboardTable) SetParams(params []*ParamDef) {
	t.Params = params
}

// GetPreparedStatementName implements QueryProvider
func (t *DashboardTable) GetPreparedStatementName() string {
	if t.PreparedStatementName != "" {
		return t.PreparedStatementName
	}
	t.PreparedStatementName = t.buildPreparedStatementName(t.ShortName, t.Mod.NameWithVersion(), constants.PreparedStatementTableSuffix)
	return t.PreparedStatementName
}

// GetResolvedQuery implements QueryProvider
func (t *DashboardTable) GetResolvedQuery(runtimeArgs *QueryArgs) (*ResolvedQuery, error) {
	// defer to base
	return t.getResolvedQuery(t, runtimeArgs)
}

func (t *DashboardTable) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(t.Base, resourceMapProvider); !resolved {
		return
	} else {
		t.Base = base.(*DashboardTable)
	}

	if t.Title == nil {
		t.Title = t.Base.Title
	}

	if t.Width == nil {
		t.Width = t.Base.Width
	}

	if t.SQL == nil {
		t.SQL = t.Base.SQL
	}

	if t.Type == nil {
		t.Type = t.Base.Type
	}

	if t.Display == nil {
		t.Display = t.Base.Display
	}

	if t.ColumnList == nil {
		t.ColumnList = t.Base.ColumnList
	} else {
		t.ColumnList.Merge(t.Base.ColumnList)
	}
}
