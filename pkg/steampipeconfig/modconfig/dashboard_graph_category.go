package modconfig

import (
	"github.com/turbot/steampipe/pkg/utils"
)

type DashboardGraphCategory struct {
	Name  string  `hcl:"name,label" json:"-" :"name"`
	Title *string `cty:"title" hcl:"title" json:"title,omitempty"`
	Color *string `cty:"color" hcl:"color" json:"color,omitempty"`
	Depth *int    `cty:"depth" hcl:"depth" json:"depth,omitempty"`
	Icon  *string `cty:"icon" hcl:"icon" json:"icon,omitempty"`
	HREF  *string `cty:"href" hcl:"href" json:"href,omitempty"`
	// TODO ask Kai to add proper fields map of field objects
	Fields *string `cty:"fields" hcl:"fields" json:"fields,omitempty"`
}

func (c DashboardGraphCategory) Equals(other *DashboardGraphCategory) bool {
	if other == nil {
		return false
	}

	return utils.SafeStringsEqual(c.Name, other.Name) &&
		utils.SafeStringsEqual(c.Title, other.Title) &&
		utils.SafeStringsEqual(c.Color, other.Color) &&
		utils.SafeIntEqual(c.Depth, other.Depth) &&
		utils.SafeStringsEqual(c.Icon, other.Icon) &&
		utils.SafeStringsEqual(c.HREF, other.HREF) &&
		utils.SafeStringsEqual(c.Fields, other.Fields)
}
