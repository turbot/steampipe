package dashboardinterfaces

// SnapshotLeafNode is an interface implemented by all nodes which are to be included in the Snapshot LeafNodes map
type SnapshotLeafNode interface {
	IsSnapshotLeafNode()
}
