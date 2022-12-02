package modconfig

import (
	"github.com/turbot/steampipe/pkg/utils"
)

type DashboardCategoryProperty struct {
	ShortName string  `hcl:"name,label" json:"name"`
	Display   *string `cty:"display" hcl:"display" json:"display,omitempty"`
	Wrap      *string `cty:"wrap" hcl:"wrap" json:"wrap,omitempty"`
	HREF      *string `cty:"href" hcl:"href" json:"href,omitempty"`
}

func (c DashboardCategoryProperty) Equals(other *DashboardCategoryProperty) bool {
	if other == nil {
		return false
	}

	return utils.SafeStringsEqual(c.ShortName, other.ShortName) &&
		utils.SafeStringsEqual(c.Display, other.Display) &&
		utils.SafeStringsEqual(c.Wrap, other.Wrap) &&
		utils.SafeStringsEqual(c.HREF, other.HREF)
}
