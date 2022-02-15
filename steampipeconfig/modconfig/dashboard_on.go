package modconfig

import (
	"github.com/turbot/steampipe/utils"
)

// DashboardOn is a struct representing dashboard hook
type DashboardOn struct {
	Name    string  `hcl:"name,label" json:"name"`
	Display *string `cty:"string" hcl:"display" json:"string,omitempty"`
}

func (s DashboardOn) Equals(other *DashboardOn) bool {
	if other == nil {
		return false
	}

	return utils.SafeStringsEqual(s.Name, other.Name) &&
		utils.SafeStringsEqual(s.Display, other.Display)
}
