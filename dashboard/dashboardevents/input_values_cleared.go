package dashboardevents

type InputValuesCleared struct {
	ClearedInputs []string
	Session       string
}

// IsDashboardEvent implements DashboardEvent interface
func (*InputValuesCleared) IsDashboardEvent() {}
