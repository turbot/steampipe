package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/utils"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// ReportTable is a struct representing a leaf reporting node
type ReportTable struct {
	HclResourceBase
	ResourceWithMetadataBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain"`

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	Title      *string                       `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width      *int                          `cty:"width" hcl:"width" column:"width,text"  json:"-"`
	ColumnList ReportTableColumnList         `cty:"column_list" hcl:"column,block" column:"columns,jsonb" json:"-"`
	Columns    map[string]*ReportTableColumn `cty:"columns" json:"columns"`

	// QueryProvider
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"sql"`
	Query                 *Query      `hcl:"query" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"args"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"params"`

	Base      *ReportTable `hcl:"base" json:"-"`
	DeclRange hcl.Range    `json:"-"`
	Mod       *Mod         `cty:"mod" json:"-"`
	Paths     []NodePath   `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewReportTable(block *hcl.Block, mod *Mod) *ReportTable {
	shortName := GetAnonymousResourceShortName(block, mod)
	t := &ReportTable{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}
	t.SetAnonymous(block)
	return t
}

func (t *ReportTable) Equals(other *ReportTable) bool {
	diff := t.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (t *ReportTable) CtyValue() (cty.Value, error) {
	return getCtyValue(t)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'table.<shortName>'
func (t *ReportTable) Name() string {
	return t.FullName
}

// OnDecoded implements HclResource
func (t *ReportTable) OnDecoded(*hcl.Block) hcl.Diagnostics {
	t.setBaseProperties()
	// populate columns map
	if len(t.ColumnList) > 0 {
		t.Columns = make(map[string]*ReportTableColumn, len(t.ColumnList))
		for _, c := range t.ColumnList {
			t.Columns[c.Name] = c
		}
	}
	return nil
}

func (t *ReportTable) setBaseProperties() {
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
	if t.ColumnList == nil {
		t.ColumnList = t.Base.ColumnList
	} else {
		t.ColumnList.Merge(t.Base.ColumnList)
	}
}

// AddReference implements HclResource
func (t *ReportTable) AddReference(*ResourceReference) {}

// GetMod implements HclResource
func (t *ReportTable) GetMod() *Mod {
	return t.Mod
}

// GetDeclRange implements HclResource
func (t *ReportTable) GetDeclRange() *hcl.Range {
	return &t.DeclRange
}

// AddParent implements ModTreeItem
func (t *ReportTable) AddParent(parent ModTreeItem) error {
	t.parents = append(t.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (t *ReportTable) GetParents() []ModTreeItem {
	return t.parents
}

// GetChildren implements ModTreeItem
func (t *ReportTable) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements ModTreeItem, ReportLeafNode
func (t *ReportTable) GetTitle() string {
	return typehelpers.SafeString(t.Title)
}

// GetDescription implements ModTreeItem
func (t *ReportTable) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (t *ReportTable) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (t *ReportTable) GetPaths() []NodePath {
	// lazy load
	if len(t.Paths) == 0 {
		t.SetPaths()
	}

	return t.Paths
}

// SetPaths implements ModTreeItem
func (t *ReportTable) SetPaths() {
	for _, parent := range t.parents {
		for _, parentPath := range parent.GetPaths() {
			t.Paths = append(t.Paths, append(parentPath, t.Name()))
		}
	}
}

func (t *ReportTable) Diff(other *ReportTable) *ReportTreeItemDiffs {
	res := &ReportTreeItemDiffs{
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

// GetWidth implements ReportLeafNode
func (t *ReportTable) GetWidth() int {
	if t.Width == nil {
		return 0
	}
	return *t.Width
}

// GetUnqualifiedName implements ReportLeafNode, ModTreeItem
func (t *ReportTable) GetUnqualifiedName() string {
	return t.UnqualifiedName
}

// GetParams implements QueryProvider
func (t *ReportTable) GetParams() []*ParamDef {
	return t.Params
}

// GetArgs implements QueryProvider
func (t *ReportTable) GetArgs() *QueryArgs {
	return t.Args
}

// GetSQL implements QueryProvider, ReportLeafNode
func (t *ReportTable) GetSQL() string {
	return typehelpers.SafeString(t.SQL)
}

// GetQuery implements QueryProvider
func (t *ReportTable) GetQuery() *Query {
	return t.Query
}

// GetPreparedStatementName implements QueryProvider
func (t *ReportTable) GetPreparedStatementName() string {
	// lazy load
	if t.PreparedStatementName == "" {
		t.PreparedStatementName = preparedStatementName(t)
	}
	return t.PreparedStatementName
}

// GetModName implements QueryProvider
func (t *ReportTable) GetModName() string {
	return t.Mod.NameWithVersion()
}

// SetArgs implements QueryProvider
func (t *ReportTable) SetArgs(args *QueryArgs) {
	// nothing
}

// SetParams implements QueryProvider
func (t *ReportTable) SetParams(params []*ParamDef) {
	t.Params = params
}
