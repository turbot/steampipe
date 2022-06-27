package modconfig

import "github.com/turbot/steampipe/pkg/utils"

// DashboardInputOption is a struct representing dashboard input option
type DashboardInputOption struct {
	Name  string  `hcl:"name,label" json:"name"`
	Label *string `cty:"label" hcl:"label" json:"label,omitempty"`
}

func (o DashboardInputOption) Equals(other *DashboardInputOption) bool {
	return utils.SafeStringsEqual(o.Name, other.Name) && utils.SafeStringsEqual(o.Label, other.Label)
}
