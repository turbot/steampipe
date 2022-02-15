package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/utils"
	"github.com/zclconf/go-cty/cty"
)

// ReportCounter is a struct representing a leaf reporting node
type ReportCounter struct {
	ReportLeafNodeBase
	ResourceWithMetadataBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	// these properties are JSON serialised by the parent LeafRun
	Title *string `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width *int    `cty:"width" hcl:"width" column:"width,text"  json:"-"`
	Type  *string `cty:"type" hcl:"type" column:"type,text"  json:"type,omitempty"`
	Style *string `cty:"style" hcl:"style" column:"style,text" json:"style,omitempty"`

	// QueryProvider
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"sql"`
	Query                 *Query      `hcl:"query" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"args,omitempty"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"params,omitempty"`

	Base *ReportCounter `hcl:"base" json:"-"`

	DeclRange hcl.Range  `json:"-"`
	Mod       *Mod       `cty:"mod" json:"-"`
	Paths     []NodePath `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewReportCounter(block *hcl.Block, mod *Mod) *ReportCounter {
	shortName := GetAnonymousResourceShortName(block, mod)
	c := &ReportCounter{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}
	c.SetAnonymous(block)

	return c
}

func (c *ReportCounter) Equals(other *ReportCounter) bool {
	diff := c.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (c *ReportCounter) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'counter.<shortName>'
func (c *ReportCounter) Name() string {
	return c.FullName
}

// OnDecoded implements HclResource
func (c *ReportCounter) OnDecoded(*hcl.Block) hcl.Diagnostics {
	c.setBaseProperties()
	return nil
}

func (c *ReportCounter) setBaseProperties() {
	if c.Base == nil {
		return
	}
	if c.Title == nil {
		c.Title = c.Base.Title
	}
	if c.Type == nil {
		c.Type = c.Base.Type
	}
	if c.Style == nil {
		c.Style = c.Base.Style
	}

	if c.Width == nil {
		c.Width = c.Base.Width
	}
	if c.SQL == nil {
		c.SQL = c.Base.SQL
	}
}

// AddReference implements HclResource
func (c *ReportCounter) AddReference(*ResourceReference) {}

// GetMod implements HclResource
func (c *ReportCounter) GetMod() *Mod {
	return c.Mod
}

// GetDeclRange implements HclResource
func (c *ReportCounter) GetDeclRange() *hcl.Range {
	return &c.DeclRange
}

// AddParent implements ModTreeItem
func (c *ReportCounter) AddParent(parent ModTreeItem) error {
	c.parents = append(c.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (c *ReportCounter) GetParents() []ModTreeItem {
	return c.parents
}

// GetChildren implements ModTreeItem
func (c *ReportCounter) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements ModTreeItem
func (c *ReportCounter) GetTitle() string {
	return typehelpers.SafeString(c.Title)
}

// GetDescription implements ModTreeItem
func (c *ReportCounter) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (c *ReportCounter) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (c *ReportCounter) GetPaths() []NodePath {
	// lazy load
	if len(c.Paths) == 0 {
		c.SetPaths()
	}

	return c.Paths
}

// SetPaths implements ModTreeItem
func (c *ReportCounter) SetPaths() {
	for _, parent := range c.parents {
		for _, parentPath := range parent.GetPaths() {
			c.Paths = append(c.Paths, append(parentPath, c.Name()))
		}
	}
}

func (c *ReportCounter) Diff(other *ReportCounter) *ReportTreeItemDiffs {
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

	if !utils.SafeStringsEqual(c.Style, other.Style) {
		res.AddPropertyDiff("Style")
	}

	res.populateChildDiffs(c, other)

	return res
}

// GetWidth implements ReportLeafNode
func (c *ReportCounter) GetWidth() int {
	if c.Width == nil {
		return 0
	}
	return *c.Width
}

// GetUnqualifiedName implements ReportLeafNode, ModTreeItem
func (c *ReportCounter) GetUnqualifiedName() string {
	return c.UnqualifiedName
}

// GetParams implements QueryProvider
func (c *ReportCounter) GetParams() []*ParamDef {
	return c.Params
}

// GetArgs implements QueryProvider
func (c *ReportCounter) GetArgs() *QueryArgs {
	return c.Args
}

// GetSQL implements QueryProvider, ReportLeafNode
func (c *ReportCounter) GetSQL() string {
	return typehelpers.SafeString(c.SQL)
}

// GetQuery implements QueryProvider
func (c *ReportCounter) GetQuery() *Query {
	return c.Query
}

// GetPreparedStatementName implements QueryProvider
func (c *ReportCounter) GetPreparedStatementName() string {
	// lazy load
	if c.PreparedStatementName == "" {
		c.PreparedStatementName = preparedStatementName(c)
	}
	return c.PreparedStatementName
}

// GetModName implements QueryProvider
func (c *ReportCounter) GetModName() string {
	return c.Mod.NameWithVersion()
}

// SetArgs implements QueryProvider
func (c *ReportCounter) SetArgs(args *QueryArgs) {
	// nothing
}

// SetParams implements QueryProvider
func (c *ReportCounter) SetParams(params []*ParamDef) {
	c.Params = params
}
