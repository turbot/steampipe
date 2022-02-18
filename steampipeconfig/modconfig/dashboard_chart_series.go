package modconfig

import "github.com/turbot/steampipe/utils"

// TODO [reports] PointList and PointMap

type DashboardChartSeries struct {
	Name   string                       `hcl:"name,label" json:"name"`
	Title  *string                      `cty:"title" hcl:"title" json:"title,omitempty"`
	Color  *string                      `cty:"color" hcl:"color" json:"color,omitempty"`
	Points []*DashboardChartSeriesPoint `cty:"points" hcl:"point,block" json:"points,omitempty"`
}

func (s DashboardChartSeries) Equals(other *DashboardChartSeries) bool {
	if other == nil {
		return false
	}

	if len(s.Points) != len(other.Points) {
		return false
	}
	for i, p := range s.Points {
		if !p.Equals(other.Points[i]) {
			return false
		}
	}

	return utils.SafeStringsEqual(s.Name, other.Name) &&
		utils.SafeStringsEqual(s.Title, other.Title) &&
		utils.SafeStringsEqual(s.Color, other.Color)
}
