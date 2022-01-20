package modconfig

import "github.com/turbot/steampipe/utils"

type ReportChartLegend struct {
	Display  *string `cty:"display" hcl:"display" json:"display,omitempty"`
	Position *string `cty:"position" hcl:"position" json:"position,omitempty"`
}

func (l ReportChartLegend) Equals(other *ReportChartLegend) bool {
	if other == nil {
		return false
	}

	return utils.SafeStringsEqual(l.Display, other.Display) &&
		utils.SafeStringsEqual(l.Position, other.Position)
}
