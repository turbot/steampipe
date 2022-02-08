package modconfig

type ReportLeafNodeBase struct {
	runtimeDependencies map[string]*RuntimeDependency
}

func (b *ReportLeafNodeBase) AddRuntimeDependencies(dependency *RuntimeDependency) {
	if b.runtimeDependencies == nil {
		b.runtimeDependencies = make(map[string]*RuntimeDependency)
	}

	b.runtimeDependencies[dependency.String()] = dependency
}

func (b *ReportLeafNodeBase) GetRuntimeDependencies() map[string]*RuntimeDependency {
	return b.runtimeDependencies
}
