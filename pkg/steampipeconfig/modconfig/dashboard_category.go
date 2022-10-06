package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

type DashboardCategory struct {
	ShortName       string `hcl:"name,label" json:"-"`
	FullName        string `cty:"name" json:"-"`
	UnqualifiedName string `json:"-"`

	Title *string                `cty:"title" hcl:"title" json:"title,omitempty"`
	Color *string                `cty:"color" hcl:"color" json:"color,omitempty"`
	Depth *int                   `cty:"depth" hcl:"depth" json:"depth,omitempty"`
	Icon  *string                `cty:"icon" hcl:"icon" json:"icon,omitempty"`
	HREF  *string                `cty:"href" hcl:"href" json:"href,omitempty"`
	Fold  *DashboardCategoryFold `cty:"fold" hcl:"fold,block" json:"fold,omitempty"`
	// TODO ask Kai to add proper fields map of field objects
	Fields *string `cty:"fields" hcl:"fields" json:"fields,omitempty"`

	Base       *DashboardCategory   `hcl:"base" json:"-"`
	References []*ResourceReference `json:"-"`
	Mod        *Mod                 `cty:"mod" json:"-"`
	DeclRange  hcl.Range            `json:"-"`
}

func NewDashboardCategory(block *hcl.Block, mod *Mod, shortName string) HclResource {
	c := &DashboardCategory{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}

	return c
}

// Name implements HclResource
// return name in format: '<modname>.control.<shortName>'
func (c *DashboardCategory) Name() string {
	return c.FullName
}

// GetUnqualifiedName implements HclResource
func (c *DashboardCategory) GetUnqualifiedName() string {
	return c.UnqualifiedName
}

// CtyValue implements HclResource
func (c *DashboardCategory) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

// GetDeclRange implements HclResource
func (c *DashboardCategory) GetDeclRange() *hcl.Range {
	return &c.DeclRange
}

// OnDecoded implements HclResource
func (c *DashboardCategory) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	c.setBaseProperties(resourceMapProvider)
	return nil
}

// AddReference implements HclResource
func (c *DashboardCategory) AddReference(ref *ResourceReference) {
	c.References = append(c.References, ref)
}

// GetReferences implements HclResource
func (c *DashboardCategory) GetReferences() []*ResourceReference {
	return c.References
}

func (c DashboardCategory) Equals(other *DashboardCategory) bool {
	if other == nil {
		return false
	}

	var foldEqual bool
	if c.Fold == nil && other == nil {
		foldEqual = true
	} else if c.Fold == nil && other != nil {
		foldEqual = false
	} else {
		foldEqual = c.Fold.Equals(other.Fold)
	}

	return utils.SafeStringsEqual(c.Name, other.Name) &&
		utils.SafeStringsEqual(c.Title, other.Title) &&
		utils.SafeStringsEqual(c.Color, other.Color) &&
		utils.SafeIntEqual(c.Depth, other.Depth) &&
		utils.SafeStringsEqual(c.Icon, other.Icon) &&
		utils.SafeStringsEqual(c.HREF, other.HREF) &&
		utils.SafeStringsEqual(c.Fields, other.Fields) &&
		foldEqual
}

func (c *DashboardCategory) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(c.Base, resourceMapProvider); !resolved {
		return
	} else {
		c.Base = base.(*DashboardCategory)
	}

	if c.Title == nil {
		c.Title = c.Base.Title
	}
	if c.Color == nil {
		c.Color = c.Base.Color
	}
	if c.Depth == nil {
		c.Depth = c.Base.Depth
	}
	if c.Icon == nil {
		c.Icon = c.Base.Icon
	}
	if c.HREF == nil {
		c.HREF = c.Base.HREF
	}
	if c.Fold == nil {
		c.Fold = c.Base.Fold
	}
	if c.Fields == nil {
		c.Fields = c.Base.Fields
	}
}
