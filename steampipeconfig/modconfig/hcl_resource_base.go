package modconfig

type HclResourceBase struct {
	runtimeDependencies map[string]*ResourceDependency
}

func (b *HclResourceBase) AddRuntimeDependencies(dependency *ResourceDependency) {
	if b.runtimeDependencies == nil {
		b.runtimeDependencies = make(map[string]*ResourceDependency)
	}

	b.runtimeDependencies[dependency.String()] = dependency
}

func (b *HclResourceBase) GetRuntimeDependencies() map[string]*ResourceDependency {
	return b.runtimeDependencies
}
