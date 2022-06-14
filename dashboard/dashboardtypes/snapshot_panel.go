package dashboardtypes

// SnapshotPanel is an interface implemented by all nodes which are to be included in the Snapshot Panels map
type SnapshotPanel interface {
	IsSnapshotPanel()
}
