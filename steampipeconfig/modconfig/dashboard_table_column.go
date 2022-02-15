package modconfig

import "github.com/turbot/steampipe/utils"

type DashboardTableColumn struct {
	Name    string  `hcl:"name,label" json:"-"`
	Display *string `cty:"display" hcl:"display" json:"display,omitempty"`
	Wrap    *string `cty:"wrap" hcl:"wrap" json:"wrap,omitempty"`
}

func (c DashboardTableColumn) Equals(other *DashboardTableColumn) bool {
	if other == nil {
		return false
	}

	return utils.SafeStringsEqual(c.Name, other.Name) &&
		utils.SafeStringsEqual(c.Display, other.Display) &&
		utils.SafeStringsEqual(c.Wrap, other.Wrap)
}
