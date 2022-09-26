package modconfig

import (
	"github.com/turbot/steampipe/pkg/utils"
)

type DashboardCategory struct {
	Name  string  `hcl:"name,label" json:"-"`
	Title *string `cty:"title" hcl:"title" json:"title,omitempty"`
	Color *string `cty:"color" hcl:"color" json:"color,omitempty"`
	Depth *int    `cty:"depth" hcl:"depth" json:"depth,omitempty"`
	Icon  *string `cty:"icon" hcl:"icon" json:"icon,omitempty"`
	HREF  *string `cty:"href" hcl:"href" json:"href,omitempty"`
	// TODO ask Kai to add proper fields map of field objects
	Fields *string                `cty:"fields" hcl:"fields" json:"fields,omitempty"`
	Fold   *DashboardCategoryFold `cty:"fold" hcl:"fold,block" json:"fold,omitempty"`
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
