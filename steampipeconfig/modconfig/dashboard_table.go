package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/utils"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// DashboardTable is a struct representing a leaf dashboard node
type DashboardTable struct {
	DashboardLeafNodeBase
	ResourceWithMetadataBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	Title      *string                          `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width      *int                             `cty:"width" hcl:"width" column:"width,text"  json:"-"`
	Type       *string                          `cty:"type" hcl:"type" column:"type,text"  json:"type,omitempty"`
	ColumnList DashboardTableColumnList         `cty:"column_list" hcl:"column,block" column:"columns,jsonb" json:"-"`
	Columns    map[string]*DashboardTableColumn `cty:"columns" json:"columns"`

	// QueryProvider
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"sql"`
	Query                 *Query      `hcl:"query" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"args"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"params"`

	Base      *DashboardTable `hcl:"base" json:"-"`
	DeclRange hcl.Range       `json:"-"`
	Mod       *Mod            `cty:"mod" json:"-"`
	Paths     []NodePath      `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewDashboardTable(block *hcl.Block, mod *Mod) *DashboardTable {
	shortName := GetAnonymousResourceShortName(block, mod)
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
func (t *DashboardTable) OnDecoded(*hcl.Block) hcl.Diagnostics {
	t.setBaseProperties()
	// populate columns map
	if len(t.ColumnList) > 0 {
		t.Columns = make(map[string]*DashboardTableColumn, len(t.ColumnList))
		for _, c := range t.ColumnList {
			t.Columns[c.Name] = c
		}
	}
	return nil
}

func (t *DashboardTable) setBaseProperties() {
	if t.Base == nil {
		return
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

	if t.ColumnList == nil {
		t.ColumnList = t.Base.ColumnList
	} else {
		t.ColumnList.Merge(t.Base.ColumnList)
	}
}

// AddReference implements HclResource
func (t *DashboardTable) AddReference(*ResourceReference) {}

// GetMod implements HclResource
func (t *DashboardTable) GetMod() *Mod {
	return t.Mod
}

// GetDeclRange implements HclResource
func (t *DashboardTable) GetDeclRange() *hcl.Range {
	return &t.DeclRange
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

// GetTitle implements ModTreeItem, DashboardLeafNode
func (t *DashboardTable) GetTitle() string {
	return typehelpers.SafeString(t.Title)
}

// GetDescription implements ModTreeItem
func (t *DashboardTable) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (t *DashboardTable) GetTags() map[string]string {
	return nil
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
	if !utils.SafeStringsEqual(t.FullName, other.FullName) {
		res.AddPropertyDiff("Name")
	}
	if !utils.SafeStringsEqual(t.Title, other.Title) {
		res.AddPropertyDiff("Title")
	}
	if !utils.SafeStringsEqual(t.SQL, other.SQL) {
		res.AddPropertyDiff("SQL")
	}

	if !utils.SafeIntEqual(t.Width, other.Width) {
		res.AddPropertyDiff("Width")
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

	return res
}

// GetWidth implements DashboardLeafNode
func (t *DashboardTable) GetWidth() int {
	if t.Width == nil {
		return 0
	}
	return *t.Width
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

// GetSQL implements QueryProvider, DashboardLeafNode
func (t *DashboardTable) GetSQL() string {
	return typehelpers.SafeString(t.SQL)
}

// GetQuery implements QueryProvider
func (t *DashboardTable) GetQuery() *Query {
	return t.Query
}

// GetPreparedStatementName implements QueryProvider
func (t *DashboardTable) GetPreparedStatementName() string {
	// lazy load
	if t.PreparedStatementName == "" {
		t.PreparedStatementName = preparedStatementName(t)
	}
	return t.PreparedStatementName
}

// GetModName implements QueryProvider
func (t *DashboardTable) GetModName() string {
	return t.Mod.NameWithVersion()
}

// SetArgs implements QueryProvider
func (t *DashboardTable) SetArgs(args *QueryArgs) {
	// nothing
}

// SetParams implements QueryProvider
func (t *DashboardTable) SetParams(params []*ParamDef) {
	t.Params = params
}
