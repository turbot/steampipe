package dashboardtypes

// SnapshotTreeNode is a struct used to store the dashboard structure in the snapshot
type SnapshotTreeNode struct {
	Name     string              `json:"name"`
	Children []*SnapshotTreeNode `json:"children,omitempty"`
	NodeType string              `json:"panel_type"`
}
