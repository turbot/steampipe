package controlexecute

import "github.com/turbot/steampipe/dashboard/dashboardinterfaces"

// ExecutionTreeNode is implemented by all control execution tree nodes
type ExecutionTreeNode interface {
	IsExecutionTreeNode()
	GetChildren() []ExecutionTreeNode
	GetName() string
	AsTreeNode() *dashboardinterfaces.SnapshotTreeNode
}
