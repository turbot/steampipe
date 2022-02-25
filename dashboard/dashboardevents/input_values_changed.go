package dashboardevents

type InputValuesCleared struct {
	ClearedInputs []string
}

// IsDashboardEvent implements DashboardEvent interface
func (*InputValuesCleared) IsDashboardEvent() {}
