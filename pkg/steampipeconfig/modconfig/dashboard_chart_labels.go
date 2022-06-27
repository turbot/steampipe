package modconfig

import "github.com/turbot/steampipe/pkg/utils"

type DashboardChartLabels struct {
	Display *string `cty:"display" hcl:"display" json:"display,omitempty"`
	Format  *string `cty:"format" hcl:"format" json:"format,omitempty"`
}

func (l *DashboardChartLabels) Equals(other *DashboardChartLabels) bool {
	if other == nil {
		return false
	}

	return utils.SafeStringsEqual(l.Display, other.Display) &&
		utils.SafeStringsEqual(l.Format, other.Format)
}

func (l *DashboardChartLabels) Merge(other *DashboardChartLabels) {
	if l.Display == nil {
		l.Display = other.Display
	}
	if l.Format == nil {
		l.Format = other.Format
	}
}
