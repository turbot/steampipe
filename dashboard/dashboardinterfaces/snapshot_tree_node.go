package dashboardinterfaces

// SnapshotTreeNode is a struct used to store the dashboard structure in the snapshot
type SnapshotTreeNode struct {
	Name     string              `json:"name"`
	Children []*SnapshotTreeNode `json:"children,omitempty"`
	NodeType string              `json:"node_type"`
	Display  string              `json:"display,omitempty"`
	Width    int                 `json:"width,omitempty"`
	Title    string              `json:"title,omitempty"`
}
