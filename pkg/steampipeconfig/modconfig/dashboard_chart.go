package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
)

// DashboardChart is a struct representing a leaf dashboard node
type DashboardChart struct {
	ResourceWithMetadataBase
	QueryProviderBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Width      *int                             `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type       *string                          `cty:"type" hcl:"type" column:"type,text" json:"-"`
	Display    *string                          `cty:"display" hcl:"display" json:"-"`
	Legend     *DashboardChartLegend            `cty:"legend" hcl:"legend,block" column:"legend,jsonb" json:"legend,omitempty"`
	SeriesList DashboardChartSeriesList         `cty:"series_list" hcl:"series,block" column:"series,jsonb" json:"-"`
	Axes       *DashboardChartAxes              `cty:"axes" hcl:"axes,block" column:"axes,jsonb" json:"axes,omitempty"`
	Grouping   *string                          `cty:"grouping" hcl:"grouping" json:"grouping,omitempty"`
	Transform  *string                          `cty:"transform" hcl:"transform" json:"transform,omitempty"`
	Series     map[string]*DashboardChartSeries `cty:"series" json:"series,omitempty"`

	Base       *DashboardChart      `hcl:"base" json:"-"`
	References []*ResourceReference `json:"-"`
	Paths      []NodePath           `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewDashboardChart(block *hcl.Block, mod *Mod, shortName string) HclResource {
	c := &DashboardChart{
		QueryProviderBase: QueryProviderBase{
			Mod: mod,
			HclResourceBase: HclResourceBase{
				ShortName:       shortName,
				FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
				UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
		},
	}

	c.SetAnonymous(block)
	return c
}

func (c *DashboardChart) Equals(other *DashboardChart) bool {
	diff := c.Diff(other)
	return !diff.HasChanges()
}

// AddReference implements ResourceWithMetadata
func (c *DashboardChart) AddReference(ref *ResourceReference) {
	c.References = append(c.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (c *DashboardChart) GetReferences() []*ResourceReference {
	return c.References
}

// AddParent implements ModTreeItem
func (c *DashboardChart) AddParent(parent ModTreeItem) error {
	c.parents = append(c.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (c *DashboardChart) GetParents() []ModTreeItem {
	return c.parents
}

// GetChildren implements ModTreeItem
func (c *DashboardChart) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements HclResource
func (c *DashboardChart) GetTitle() string {
	return typehelpers.SafeString(c.Title)
}

// GetPaths implements ModTreeItem
func (c *DashboardChart) GetPaths() []NodePath {
	// lazy load
	if len(c.Paths) == 0 {
		c.SetPaths()
	}

	return c.Paths
}

// SetPaths implements ModTreeItem
func (c *DashboardChart) SetPaths() {
	for _, parent := range c.parents {
		for _, parentPath := range parent.GetPaths() {
			c.Paths = append(c.Paths, append(parentPath, c.Name()))
		}
	}
}

func (c *DashboardChart) Diff(other *DashboardChart) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: c,
		Name: c.Name(),
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
	res.queryProviderDiff(c, other)
	res.dashboardLeafNodeDiff(c, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (c *DashboardChart) GetWidth() int {
	if c.Width == nil {
		return 0
	}
	return *c.Width
}

// GetDisplay implements DashboardLeafNode
func (c *DashboardChart) GetDisplay() string {
	return typehelpers.SafeString(c.Display)
}

// GetDocumentation implements DashboardLeafNode, ModTreeItem
func (c *DashboardChart) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (c *DashboardChart) GetType() string {
	return typehelpers.SafeString(c.Type)
}

func (c *DashboardChart) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(c.Base, resourceMapProvider); !resolved {
		return
	} else {
		c.Base = base.(*DashboardChart)
	}

	if c.Title == nil {
		c.Title = c.Base.Title
	}

	if c.Type == nil {
		c.Type = c.Base.Type
	}

	if c.Display == nil {
		c.Display = c.Base.Display
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

	if c.Query == nil {
		c.Query = c.Base.Query
	}

	if c.Args == nil {
		c.Args = c.Base.Args
	}

	if c.Params == nil {
		c.Params = c.Base.Params
	}

	c.MergeRuntimeDependencies(c.Base)
}
