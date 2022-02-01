package modconfig

type ReportChartSeriesList []*ReportChartSeries

func (s *ReportChartSeriesList) Merge(other ReportChartSeriesList) {
	if other == nil {
		return
	}
	var seriesMap = make(map[string]bool)
	for _, series := range *s {
		seriesMap[series.Name] = true
	}

	for _, otherSeries := range other {
		if !seriesMap[otherSeries.Name] {
			*s = append(*s, otherSeries)
		}
	}
}
