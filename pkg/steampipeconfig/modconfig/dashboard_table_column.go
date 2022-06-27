package modconfig

import "github.com/turbot/steampipe/pkg/utils"

type DashboardTableColumn struct {
	Name    string  `hcl:"name,label" json:"name"`
	Display *string `cty:"display" hcl:"display" json:"display,omitempty"`
	Wrap    *string `cty:"wrap" hcl:"wrap" json:"wrap,omitempty"`
	HREF    *string `cty:"href" hcl:"href" json:"href,omitempty"`
}

func (c DashboardTableColumn) Equals(other *DashboardTableColumn) bool {
	if other == nil {
		return false
	}

	return utils.SafeStringsEqual(c.Name, other.Name) &&
		utils.SafeStringsEqual(c.Display, other.Display) &&
		utils.SafeStringsEqual(c.Wrap, other.Wrap) &&
		utils.SafeStringsEqual(c.HREF, other.HREF)
}
