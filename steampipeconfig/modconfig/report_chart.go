package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/utils"
	"github.com/zclconf/go-cty/cty"
)

// ReportChart is a struct representing a leaf reporting node
type ReportChart struct {
	HclResourceBase
	ResourceWithMetadataBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain"`

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	// these properties are JSON serialised by the parent LeafRun
	Title *string `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width *int    `cty:"width" hcl:"width" column:"width,text"  json:"-"`

	Type       *string                       `cty:"type" hcl:"type" column:"type,text"  json:"type,omitempty"`
	Legend     *ReportChartLegend            `cty:"legend" hcl:"legend,block" column:"legend,jsonb" json:"legend"`
	SeriesList ReportChartSeriesList         `cty:"series_list" hcl:"series,block" column:"series,jsonb" json:"-"`
	Axes       *ReportChartAxes              `cty:"axes" hcl:"axes,block" column:"axes,jsonb" json:"axes"`
	Grouping   *string                       `cty:"grouping" hcl:"grouping" json:"grouping,omitempty"`
	Transform  *string                       `cty:"transform" hcl:"transform" json:"transform,omitempty"`
	Series     map[string]*ReportChartSeries `cty:"series" json:"series"`

	// QueryProvider
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"sql"`
	Query                 *Query      `hcl:"query" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"args"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"params"`

	Base      *ReportChart `hcl:"base" json:"-"`
	DeclRange hcl.Range    `json:"-"`
	Mod       *Mod         `cty:"mod" json:"-"`
	Paths     []NodePath   `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewReportChart(block *hcl.Block, mod *Mod) *ReportChart {
	shortName := GetAnonymousResourceShortName(block, mod)
	c := &ReportChart{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}
	c.SetAnonymous(block)

	return c
}

func (c *ReportChart) Equals(other *ReportChart) bool {
	diff := c.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (c *ReportChart) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'chart.<shortName>'
func (c *ReportChart) Name() string {
	return c.FullName
}

// OnDecoded implements HclResource
func (c *ReportChart) OnDecoded(*hcl.Block) hcl.Diagnostics {
	c.setBaseProperties()
	// populate series map
	if len(c.SeriesList) > 0 {
		c.Series = make(map[string]*ReportChartSeries, len(c.SeriesList))
		for _, s := range c.SeriesList {
			c.Series[s.Name] = s
		}
	}
	return nil
}

func (c *ReportChart) setBaseProperties() {
	if c.Base == nil {
		return
	}
	if c.Title == nil {
		c.Title = c.Base.Title
	}
	if c.Type == nil {
		c.Type = c.Base.Type
	}
	if c.Axes == nil {
		c.Axes = c.Base.Axes
	} else {
		c.Axes.Merge(c.Base.Axes)
	}
	if c.Grouping == nil {
		c.Grouping = c.Base.Grouping
	}
	if c.Transform == nil {
		c.Transform = c.Base.Transform
	}
	if c.Legend == nil {
		c.Legend = c.Base.Legend
	} else {
		c.Legend.Merge(c.Base.Legend)
	}
	if c.SeriesList == nil {
		c.SeriesList = c.Base.SeriesList
	} else {

		c.SeriesList.Merge(c.Base.SeriesList)
	}

	if c.Width == nil {
		c.Width = c.Base.Width
	}
	if c.SQL == nil {
		c.SQL = c.Base.SQL
	}
}

// AddReference implements HclResource
func (c *ReportChart) AddReference(*ResourceReference) {}

// GetMod implements HclResource
func (c *ReportChart) GetMod() *Mod {
	return c.Mod
}

// GetDeclRange implements HclResource
func (c *ReportChart) GetDeclRange() *hcl.Range {
	return &c.DeclRange
}

// AddParent implements ModTreeItem
func (c *ReportChart) AddParent(parent ModTreeItem) error {
	c.parents = append(c.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (c *ReportChart) GetParents() []ModTreeItem {
	return c.parents
}

// GetChildren implements ModTreeItem
func (c *ReportChart) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements ModTreeItem
func (c *ReportChart) GetTitle() string {
	return typehelpers.SafeString(c.Title)
}

// GetDescription implements ModTreeItem
func (c *ReportChart) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (c *ReportChart) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (c *ReportChart) GetPaths() []NodePath {
	// lazy load
	if len(c.Paths) == 0 {
		c.SetPaths()
	}

	return c.Paths
}

// SetPaths implements ModTreeItem
func (c *ReportChart) SetPaths() {
	for _, parent := range c.parents {
		for _, parentPath := range parent.GetPaths() {
			c.Paths = append(c.Paths, append(parentPath, c.Name()))
		}
	}
}

func (c *ReportChart) Diff(other *ReportChart) *ReportTreeItemDiffs {
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

	if !utils.SafeStringsEqual(c.Grouping, other.Grouping) {
		res.AddPropertyDiff("Grouping")
	}

	if !utils.SafeStringsEqual(c.Transform, other.Transform) {
		res.AddPropertyDiff("Transform")
	}

	if len(c.SeriesList) != len(other.SeriesList) {
		res.AddPropertyDiff("Series")
	} else {
		for i, s := range c.Series {
			if !s.Equals(other.Series[i]) {
				res.AddPropertyDiff("Series")
			}
		}
	}

	if c.Legend != nil {
		if !c.Legend.Equals(other.Legend) {
			res.AddPropertyDiff("Legend")
		}
	} else if other.Legend != nil {
		res.AddPropertyDiff("Legend")
	}

	if c.Axes != nil {
		if !c.Axes.Equals(other.Axes) {
			res.AddPropertyDiff("Axes")
		}
	} else if other.Axes != nil {
		res.AddPropertyDiff("Axes")
	}

	res.populateChildDiffs(c, other)

	return res
}

// GetWidth implements ReportLeafNode
func (c *ReportChart) GetWidth() int {
	if c.Width == nil {
		return 0
	}
	return *c.Width
}

// GetUnqualifiedName implements ReportLeafNode, ModTreeItem
func (c *ReportChart) GetUnqualifiedName() string {
	return c.UnqualifiedName
}

// GetParams implements QueryProvider
func (c *ReportChart) GetParams() []*ParamDef {
	return c.Params
}

// GetArgs implements QueryProvider
func (c *ReportChart) GetArgs() *QueryArgs {
	return c.Args
}

// GetSQL implements QueryProvider, ReportLeafNode
func (c *ReportChart) GetSQL() string {
	return typehelpers.SafeString(c.SQL)
}

// GetQuery implements QueryProvider
func (c *ReportChart) GetQuery() *Query {
	return c.Query
}

// GetPreparedStatementName implements QueryProvider
func (c *ReportChart) GetPreparedStatementName() string {
	// lazy load
	if c.PreparedStatementName == "" {
		c.PreparedStatementName = preparedStatementName(c)
	}
	return c.PreparedStatementName
}

// GetModName implements QueryProvider
func (c *ReportChart) GetModName() string {
	return c.Mod.NameWithVersion()
}

// SetArgs implements QueryProvider
func (c *ReportChart) SetArgs(args *QueryArgs) {
	// nothing
}

// SetParams implements QueryProvider
func (c *ReportChart) SetParams(params []*ParamDef) {
	c.Params = params
}
