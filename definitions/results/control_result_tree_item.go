package results

type ControlResultTreeNode interface {
	Children() []ControlResultTreeNode
	Parent() BenchmarkResult
}
