package modconfig

type HclResourceBase struct {
	runtimeDependencies map[string]*RuntimeDependency
}

func (b *HclResourceBase) AddRuntimeDependencies(dependency *RuntimeDependency) {
	if b.runtimeDependencies == nil {
		b.runtimeDependencies = make(map[string]*RuntimeDependency)
	}

	b.runtimeDependencies[dependency.String()] = dependency
}

func (b *HclResourceBase) GetRuntimeDependencies() map[string]*RuntimeDependency {
	return b.runtimeDependencies
}
