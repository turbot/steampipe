package dashboardtypes

// SnapshotPanel is an interface implemented by all nodes which are to be included in the Snapshot Panels map
// this consists of all 'Run' types - LeafRun, DashboardRun, etc.
type SnapshotPanel interface {
	IsSnapshotPanel()
}
