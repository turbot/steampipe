package modconfig

import "github.com/turbot/steampipe/pkg/utils"

type DashboardGraphCategory struct {
	Name  string  `hcl:"name,label" json:"-"`
	Title *string `cty:"title" hcl:"title" json:"title,omitempty"`
	Color *string `cty:"color" hcl:"color" json:"color,omitempty"`
	Depth *int    `cty:"depth" hcl:"depth" json:"depth,omitempty"`
}

func (c DashboardGraphCategory) Equals(other *DashboardGraphCategory) bool {
	if other == nil {
		return false
	}

	return utils.SafeStringsEqual(c.Name, other.Name) &&
		utils.SafeStringsEqual(c.Title, other.Title) &&
		utils.SafeStringsEqual(c.Color, other.Color) &&
		utils.SafeIntEqual(c.Depth, other.Depth)
}
