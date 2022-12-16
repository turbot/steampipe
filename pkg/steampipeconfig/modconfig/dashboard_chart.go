package modconfig

import (
	"fmt"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
)

// DashboardChart is a struct representing a leaf dashboard node
type DashboardChart struct {
	ResourceWithMetadataImpl
	QueryProviderImpl

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
	Base       *DashboardChart                  `hcl:"base" json:"-"`
	References []*ResourceReference             `json:"-"`
	Paths      []NodePath                       `column:"path,jsonb" json:"-"`
}

func NewDashboardChart(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)

	c := &DashboardChart{
		QueryProviderImpl: QueryProviderImpl{
			RuntimeDependencyProviderImpl: RuntimeDependencyProviderImpl{
				ModTreeItemImpl: ModTreeItemImpl{
					HclResourceImpl: HclResourceImpl{
						ShortName:       shortName,
						FullName:        fullName,
						UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
						DeclRange:       block.DefRange,
						blockType:       block.Type,
					},
					Mod: mod,
				},
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

// OnDecoded implements HclResource
func (c *DashboardChart) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	c.setBaseProperties(resourceMapProvider)
	// populate series map
	if len(c.SeriesList) > 0 {
		c.Series = make(map[string]*DashboardChartSeries, len(c.SeriesList))
		for _, s := range c.SeriesList {
			s.OnDecoded()
			c.Series[s.Name] = s
		}
	}
	return nil
}

// AddReference implements ResourceWithMetadata
func (c *DashboardChart) AddReference(ref *ResourceReference) {
	c.References = append(c.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (c *DashboardChart) GetReferences() []*ResourceReference {
	return c.References
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

// GetType implements DashboardLeafNode
func (c *DashboardChart) GetType() string {
	return typehelpers.SafeString(c.Type)
}

// CtyValue implements CtyValueProvider
func (c *DashboardChart) CtyValue() (cty.Value, error) {
	return GetCtyValue(c)
}

func (c *DashboardChart) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	base, resolved := resolveBase(c.Base, resourceMapProvider)
	if !resolved {
		return
	}
	c.base = base
	c.QueryProviderImpl.setBaseProperties()
	baseChart := base.(*DashboardChart)

	if c.Type == nil {
		c.Type = baseChart.Type
	}

	if c.Display == nil {
		c.Display = baseChart.Display
	}

	if c.Axes == nil {
		c.Axes = baseChart.Axes
	} else {
		c.Axes.Merge(baseChart.Axes)
	}

	if c.Grouping == nil {
		c.Grouping = baseChart.Grouping
	}

	if c.Transform == nil {
		c.Transform = baseChart.Transform
	}

	if c.Legend == nil {
		c.Legend = baseChart.Legend
	} else {
		c.Legend.Merge(baseChart.Legend)
	}

	if c.SeriesList == nil {
		c.SeriesList = baseChart.SeriesList
	} else {
		c.SeriesList.Merge(baseChart.SeriesList)
	}

	if c.Width == nil {
		c.Width = baseChart.Width
	}

	c.MergeRuntimeDependencies(baseChart)
}
