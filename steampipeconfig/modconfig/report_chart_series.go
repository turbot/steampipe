package modconfig

import "github.com/turbot/steampipe/utils"

type ReportChartSeries struct {
	Name  string  `hcl:"name,label"`
	Title *string `cty:"title" hcl:"title" json:"title,omitempty"`
	Color *string `cty:"color" hcl:"color" json:"color,omitempty"`
}

func (s ReportChartSeries) Equals(other *ReportChartSeries) bool {
	if other == nil {
		return false
	}

	return utils.SafeStringsEqual(s.Name, other.Name) &&
		utils.SafeStringsEqual(s.Title, other.Title) &&
		utils.SafeStringsEqual(s.Color, other.Color)
}
