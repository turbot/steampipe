package modconfig

type DashboardLeafNodeBase struct {
	runtimeDependencies map[string]*RuntimeDependency
}

func (b *DashboardLeafNodeBase) AddRuntimeDependencies(dependency *RuntimeDependency) {
	if b.runtimeDependencies == nil {
		b.runtimeDependencies = make(map[string]*RuntimeDependency)
	}

	b.runtimeDependencies[dependency.String()] = dependency
}

func (b *DashboardLeafNodeBase) GetRuntimeDependencies() map[string]*RuntimeDependency {
	return b.runtimeDependencies
}
