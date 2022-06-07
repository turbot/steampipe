package controlexecute

// empty interface implemented by execution tree nodes
type ExecutionTreeNode interface {
	IsExecutionTreeNode() bool
}

func (*ControlRun) IsExecutionTreeNode() bool  { return true }
func (*ResultGroup) IsExecutionTreeNode() bool { return true }
