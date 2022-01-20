package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/utils"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// TODO KAI pretty sure we can remove all cty properties from report leaves as they cannot be referred to

// ReportChart is a struct representing a leaf reporting node
type ReportChart struct {
	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	// these properties are JSON serialised by the parent LeafRun
	Title *string `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width *int    `cty:"width" hcl:"width" column:"width,text"  json:"-"`
	SQL   *string `cty:"sql" hcl:"sql" column:"sql,text" json:"-"`

	Type *string      `cty:"type" hcl:"type" column:"type,text"  json:"type,omitempty"`
	Base *ReportChart `hcl:"base" json:"-"`

	Legend     *ReportChartLegend   `cty:"legend" hcl:"legend,block" column:"legend,jsonb" json:"legend"`
	SeriesList []*ReportChartSeries `cty:"series_list" hcl:"series,block" column:"series,jsonb" json:"-"`
	Axes       *ReportChartAxes     `cty:"axes" hcl:"axes,block" column:"axes,jsonb" json:"axes"`

	Series map[string]*ReportChartSeries `cty:"series" json:"series"`

	DeclRange hcl.Range  `json:"-"`
	Mod       *Mod       `cty:"mod" json:"-"`
	Paths     []NodePath `column:"path,jsonb" json:"-"`

	parents   []ModTreeItem
	metadata  *ResourceMetadata
	anonymous bool
}

func NewReportChart(block *hcl.Block) *ReportChart {
	return &ReportChart{
		DeclRange:       block.DefRange,
		ShortName:       block.Labels[0],
		FullName:        fmt.Sprintf("%s.%s", block.Type, block.Labels[0]),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, block.Labels[0]),
	}
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

func (c *ReportChart) SetAnonymous(anonymous bool) {
	c.anonymous = anonymous
}

func (c *ReportChart) IsAnonymous() bool {
	return c.anonymous
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
	// TODO KAI legens,series,axes

	if c.Width == nil {
		c.Width = c.Base.Width
	}
	if c.SQL == nil {
		c.SQL = c.Base.SQL
	}
}

// AddReference implements HclResource
func (c *ReportChart) AddReference(*ResourceReference) {}

// SetMod implements HclResource
func (c *ReportChart) SetMod(mod *Mod) {
	c.Mod = mod
	c.FullName = fmt.Sprintf("%s.%s", c.Mod.ShortName, c.UnqualifiedName)
}

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

// GetMetadata implements ResourceWithMetadata
func (c *ReportChart) GetMetadata() *ResourceMetadata {
	return c.metadata
}

// SetMetadata implements ResourceWithMetadata
func (c *ReportChart) SetMetadata(metadata *ResourceMetadata) {
	c.metadata = metadata
}

func (c *ReportChart) Diff(other *ReportChart) *ReportTreeItemDiffs {
	res := &ReportTreeItemDiffs{
		Item: c,
		Name: c.Name(),
	}

	if utils.SafeStringsEqual(c.FullName, other.FullName) {
		res.AddPropertyDiff("Name")
	}

	if utils.SafeStringsEqual(c.Title, other.Title) {
		res.AddPropertyDiff("Title")
	}

	if utils.SafeStringsEqual(c.SQL, other.SQL) {
		res.AddPropertyDiff("SQL")
	}

	if utils.SafeIntEqual(c.Width, other.Width) {
		res.AddPropertyDiff("Width")
	}

	if utils.SafeStringsEqual(c.Type, other.Type) {
		res.AddPropertyDiff("Type")
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
			res.AddPropertyDiff("Series")
		}
	} else if other.Legend != nil {
		res.AddPropertyDiff("Series")
	}

	if c.Axes != nil {
		if !c.Axes.Equals(other.Axes) {
			res.AddPropertyDiff("Series")
		}
	} else if other.Axes != nil {
		res.AddPropertyDiff("Series")
	}

	res.populateChildDiffs(c, other)

	return res
}

// GetSQL implements ReportLeafNode
func (c *ReportChart) GetSQL() string {
	return typehelpers.SafeString(c.SQL)
}

// GetWidth implements ReportLeafNode
func (c *ReportChart) GetWidth() int {
	if c.Width == nil {
		return 0
	}
	return *c.Width
}

// GetUnqualifiedName implements ReportLeafNode
func (c *ReportChart) GetUnqualifiedName() string {
	return c.UnqualifiedName
}
