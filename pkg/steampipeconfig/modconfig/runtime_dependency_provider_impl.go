package modconfig

import (
	"github.com/hashicorp/hcl/v2"
)

type RuntimeDependencyProviderImpl struct {
	ModTreeItemImpl
	// required to allow partial decoding
	RuntimeDependencyProviderRemain hcl.Body `hcl:",remain" json:"-"`

	runtimeDependencies map[string]*RuntimeDependency
}

func (b *RuntimeDependencyProviderImpl) AddRuntimeDependencies(dependencies []*RuntimeDependency) {
	if b.runtimeDependencies == nil {
		b.runtimeDependencies = make(map[string]*RuntimeDependency)
	}
	for _, dependency := range dependencies {
		// set the dependency provider (this is used if this resource is inherited via base)
		dependency.Provider = b
		b.runtimeDependencies[dependency.String()] = dependency
	}
}

func (b *RuntimeDependencyProviderImpl) GetRuntimeDependencies() map[string]*RuntimeDependency {
	return b.runtimeDependencies
}
