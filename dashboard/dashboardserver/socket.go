package dashboardserver

type ClientRequestDashboardPayload struct {
	FullName string `json:"full_name"`
}

type ClientRequestPayload struct {
	Dashboard ClientRequestDashboardPayload `json:"dashboard"`
}

type ClientRequest struct {
	Action  string               `json:"action"`
	Payload ClientRequestPayload `json:"payload"`
}

type ModAvailableDashboard struct {
	Title     string `json:"title,omitempty"`
	FullName  string `json:"full_name"`
	ShortName string `json:"short_name"`
}

type AvailableDashboardsPayload struct {
	Action          string                                      `json:"action"`
	DashboardsByMod map[string]map[string]ModAvailableDashboard `json:"dashboards_by_mod"`
}

type ModDashboardMetadata struct {
	Title     string `json:"title,omitempty"`
	FullName  string `json:"full_name"`
	ShortName string `json:"short_name"`
}

type DashboardMetadata struct {
	Mod           ModDashboardMetadata            `json:"mod"`
	InstalledMods map[string]ModDashboardMetadata `json:"installed_mods,omitempty"`
}

type DashboardMetadataPayload struct {
	Action   string            `json:"action"`
	Metadata DashboardMetadata `json:"metadata"`
}
