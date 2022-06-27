package modconfig

import "github.com/turbot/steampipe/pkg/utils"

type DashboardChartSeries struct {
	Name       string                                `hcl:"name,label" json:"name"`
	Title      *string                               `cty:"title" hcl:"title" json:"title,omitempty"`
	Color      *string                               `cty:"color" hcl:"color" json:"color,omitempty"`
	Points     map[string]*DashboardChartSeriesPoint `cty:"points" json:"points,omitempty"`
	PointsList []*DashboardChartSeriesPoint          `hcl:"point,block" json:"-"`
}

func (s DashboardChartSeries) Equals(other *DashboardChartSeries) bool {
	if other == nil {
		return false
	}

	if len(s.PointsList) != len(other.PointsList) {
		return false
	}
	for i, p := range s.PointsList {
		if !p.Equals(other.PointsList[i]) {
			return false
		}
	}

	return utils.SafeStringsEqual(s.Name, other.Name) &&
		utils.SafeStringsEqual(s.Title, other.Title) &&
		utils.SafeStringsEqual(s.Color, other.Color)
}

func (s *DashboardChartSeries) OnDecoded() {
	if len(s.PointsList) > 0 {
		s.Points = make(map[string]*DashboardChartSeriesPoint, len(s.PointsList))
		for _, p := range s.PointsList {
			s.Points[p.Name] = p
		}
	}
}
