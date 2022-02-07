package modconfig

import "fmt"

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

func (b *ReportLeafNodeBase) SetRuntimeDependency(name, value string) error {
	dep, ok := b.runtimeDependencies[name]
	if !ok {
		return fmt.Errorf("runtime dependency % not found", name)
	}
	dep.Value = &value
	return nil

}
