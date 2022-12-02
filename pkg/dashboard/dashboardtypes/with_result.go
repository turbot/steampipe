package dashboardtypes

type WithResult struct {
	*LeafData
	Error error
}

type ResolvedRuntimeDependencyValue struct {
	Value any
	Error error
}
