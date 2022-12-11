package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"golang.org/x/exp/maps"
)

type RuntimeDependencyProviderBase struct {
	// cty tag required to ensure base struct is serialised in the eval context
	ModTreeItemBase `cty:"mod_tree_item_base"`
	// required to allow partial decoding
	RuntimeDependencyProviderBaseRemain hcl.Body `hcl:",remain" json:"-"`

	// map of withs keyed by unqualified name
	withs               map[string]*DashboardWith
	runtimeDependencies map[string]*RuntimeDependency
}

func (b *RuntimeDependencyProviderBase) AddWith(with *DashboardWith) hcl.Diagnostics {
	if b.withs == nil {
		b.withs = make(map[string]*DashboardWith)
	}
	// if we already have this with, fail
	if _, ok := b.withs[with.UnqualifiedName]; ok {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("duplicate with block '%s'", with.ShortName),
			Subject:  with.GetDeclRange(),
		}}
	}
	b.withs[with.UnqualifiedName] = with
	return nil
}

func (b *RuntimeDependencyProviderBase) GetWiths() []*DashboardWith {
	return maps.Values(b.withs)
}

func (b *RuntimeDependencyProviderBase) AddRuntimeDependencies(dependencies []*RuntimeDependency) {
	if b.runtimeDependencies == nil {
		b.runtimeDependencies = make(map[string]*RuntimeDependency)
	}
	for _, dependency := range dependencies {
		b.runtimeDependencies[dependency.String()] = dependency
	}
}

func (b *RuntimeDependencyProviderBase) MergeRuntimeDependencies(other QueryProvider) {
	dependencies := other.GetRuntimeDependencies()
	if b.runtimeDependencies == nil {
		b.runtimeDependencies = make(map[string]*RuntimeDependency)
	}
	for _, dependency := range dependencies {
		if _, ok := b.runtimeDependencies[dependency.String()]; !ok {
			b.runtimeDependencies[dependency.String()] = dependency
		}
	}
}

func (b *RuntimeDependencyProviderBase) GetRuntimeDependencies() map[string]*RuntimeDependency {
	return b.runtimeDependencies
}
