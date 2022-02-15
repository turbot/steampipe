package modconfig

import "github.com/turbot/steampipe/utils"

type DashboardHierarchyCategory struct {
	Name  string  `hcl:"name,label" json:"-"`
	Title *string `cty:"title" hcl:"title" json:"title,omitempty"`
	Color *string `cty:"color" hcl:"color" json:"color,omitempty"`
}

func (c DashboardHierarchyCategory) Equals(other *DashboardHierarchyCategory) bool {
	if other == nil {
		return false
	}

	return utils.SafeStringsEqual(c.Name, other.Name) &&
		utils.SafeStringsEqual(c.Title, other.Title) &&
		utils.SafeStringsEqual(c.Color, other.Color)
}
