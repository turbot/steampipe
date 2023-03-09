package modconfig

import "github.com/turbot/steampipe/pkg/utils"

type DashboardChartLegend struct {
	Display  *string `cty:"display" hcl:"display" json:"display,omitempty"`
	Position *string `cty:"position" hcl:"position" json:"position,omitempty"`
}

func (l *DashboardChartLegend) Equals(other *DashboardChartLegend) bool {
	if other == nil {
		return false
	}

	return utils.SafeStringsEqual(l.Display, other.Display) &&
		utils.SafeStringsEqual(l.Position, other.Position)
}

func (l *DashboardChartLegend) Merge(other *DashboardChartLegend) {
	if l.Display == nil {
		l.Display = other.Display
	}
	if l.Position == nil {
		l.Position = other.Position
	}
}
