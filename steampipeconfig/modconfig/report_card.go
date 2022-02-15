package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/utils"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// ReportCard is a struct representing a leaf reporting node
type ReportCard struct {
	ReportLeafNodeBase
	ResourceWithMetadataBase

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	// these properties are JSON serialised by the parent LeafRun
	Title *string `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width *int    `cty:"width" hcl:"width" column:"width,text"  json:"-"`
	Type  *string `cty:"type" hcl:"type" column:"type,text" json:"type,omitempty"`
	Icon  *string `cty:"icon" hcl:"icon" column:"icon,text" json:"icon,omitempty"`

	// QueryProvider
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"sql"`
	Query                 *Query      `hcl:"query" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"args,omitempty"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"params,omitempty"`

	Base *ReportCard `hcl:"base" json:"-"`

	DeclRange hcl.Range  `json:"-"`
	Mod       *Mod       `cty:"mod" json:"-"`
	Paths     []NodePath `column:"path,jsonb" json:"-"`

	parents  []ModTreeItem
	metadata *ResourceMetadata
}

func NewReportCard(block *hcl.Block, mod *Mod) *ReportCard {
	shortName := GetAnonymousResourceShortName(block, mod)
	c := &ReportCard{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}
	c.SetAnonymous(block)

	return c

}

func (c *ReportCard) Equals(other *ReportCard) bool {
	diff := c.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (c *ReportCard) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'card.<shortName>'
func (c *ReportCard) Name() string {
	return c.FullName
}

// OnDecoded implements HclResource
func (c *ReportCard) OnDecoded(*hcl.Block) hcl.Diagnostics {
	c.setBaseProperties()
	return nil
}

func (c *ReportCard) setBaseProperties() {
	if c.Base == nil {
		return
	}
	if c.Title == nil {
		c.Title = c.Base.Title
	}
	if c.Type == nil {
		c.Type = c.Base.Type
	}
	if c.Icon == nil {
		c.Icon = c.Base.Icon
	}
	if c.Width == nil {
		c.Width = c.Base.Width
	}
	if c.SQL == nil {
		c.SQL = c.Base.SQL
	}
}

// AddReference implements HclResource
func (c *ReportCard) AddReference(*ResourceReference) {}

// GetMod implements HclResource
func (c *ReportCard) GetMod() *Mod {
	return c.Mod
}

// GetDeclRange implements HclResource
func (c *ReportCard) GetDeclRange() *hcl.Range {
	return &c.DeclRange
}

// AddParent implements ModTreeItem
func (c *ReportCard) AddParent(parent ModTreeItem) error {
	c.parents = append(c.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (c *ReportCard) GetParents() []ModTreeItem {
	return c.parents
}

// GetChildren implements ModTreeItem
func (c *ReportCard) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements ModTreeItem
func (c *ReportCard) GetTitle() string {
	return typehelpers.SafeString(c.Title)
}

// GetDescription implements ModTreeItem
func (c *ReportCard) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (c *ReportCard) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (c *ReportCard) GetPaths() []NodePath {
	// lazy load
	if len(c.Paths) == 0 {
		c.SetPaths()
	}

	return c.Paths
}

// SetPaths implements ModTreeItem
func (c *ReportCard) SetPaths() {
	for _, parent := range c.parents {
		for _, parentPath := range parent.GetPaths() {
			c.Paths = append(c.Paths, append(parentPath, c.Name()))
		}
	}
}

// GetMetadata implements ResourceWithMetadata
func (c *ReportCard) GetMetadata() *ResourceMetadata {
	return c.metadata
}

// SetMetadata implements ResourceWithMetadata
func (c *ReportCard) SetMetadata(metadata *ResourceMetadata) {
	c.metadata = metadata
}

func (c *ReportCard) Diff(other *ReportCard) *ReportTreeItemDiffs {
	res := &ReportTreeItemDiffs{
		Item: c,
		Name: c.Name(),
	}
	if !utils.SafeStringsEqual(c.FullName, other.FullName) {
		res.AddPropertyDiff("Name")
	}

	if !utils.SafeStringsEqual(c.Title, other.Title) {
		res.AddPropertyDiff("Title")
	}

	if !utils.SafeStringsEqual(c.SQL, other.SQL) {
		res.AddPropertyDiff("SQL")
	}

	if !utils.SafeIntEqual(c.Width, other.Width) {
		res.AddPropertyDiff("Width")
	}

	if !utils.SafeStringsEqual(c.Type, other.Type) {
		res.AddPropertyDiff("Type")
	}

	if !utils.SafeStringsEqual(c.Icon, other.Icon) {
		res.AddPropertyDiff("Icon")
	}

	res.populateChildDiffs(c, other)

	return res
}

// GetWidth implements ReportLeafNode
func (c *ReportCard) GetWidth() int {
	if c.Width == nil {
		return 0
	}
	return *c.Width
}

// GetUnqualifiedName implements ReportLeafNode
func (c *ReportCard) GetUnqualifiedName() string {
	return c.UnqualifiedName
}

// GetParams implements QueryProvider
func (c *ReportCard) GetParams() []*ParamDef {
	return c.Params
}

// GetArgs implements QueryProvider
func (c *ReportCard) GetArgs() *QueryArgs {
	return c.Args
}

// GetSQL implements QueryProvider, ReportLeafNode
func (c *ReportCard) GetSQL() string {
	return typehelpers.SafeString(c.SQL)
}

// GetQuery implements QueryProvider
func (c *ReportCard) GetQuery() *Query {
	return c.Query
}

// GetPreparedStatementName implements QueryProvider
func (c *ReportCard) GetPreparedStatementName() string {
	// lazy load
	if c.PreparedStatementName == "" {
		c.PreparedStatementName = preparedStatementName(c)
	}
	return c.PreparedStatementName
}

// GetModName implements QueryProvider
func (c *ReportCard) GetModName() string {
	return c.Mod.NameWithVersion()
}

// SetArgs implements QueryProvider
func (c *ReportCard) SetArgs(args *QueryArgs) {
	// nothing
}

// SetParams implements QueryProvider
func (c *ReportCard) SetParams(params []*ParamDef) {
	c.Params = params
}
