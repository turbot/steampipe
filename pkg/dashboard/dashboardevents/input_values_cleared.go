package dashboardevents

type InputValuesCleared struct {
	ClearedInputs []string
	Session       string
	ExecutionId   string
}

// IsDashboardEvent implements DashboardEvent interface
func (*InputValuesCleared) IsDashboardEvent() {}
