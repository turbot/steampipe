package modconfig

import (
	"github.com/turbot/steampipe/pkg/utils"
	"strings"
)

// ModVariableMap is a struct containins maps of variable definitions
type ModVariableMap struct {
	ModInstallCacheKey string
	RootVariables      map[string]*Variable
	// map of dependency variable maps, keyed by dependency name
	DependencyVariables map[string]*ModVariableMap

	PublicVariables      map[string]*Variable
	PublicVariableValues map[string]string
}

// NewModVariableMap builds a ModVariableMap using the variables from a mod and its dependencies
func NewModVariableMap(mod *Mod) *ModVariableMap {
	m := &ModVariableMap{
		ModInstallCacheKey:   mod.GetInstallCacheKey(),
		RootVariables:        make(map[string]*Variable),
		DependencyVariables:  make(map[string]*ModVariableMap),
		PublicVariableValues: make(map[string]string),
	}

	// add variables into map, modifying the key to be the variable short name
	for k, v := range mod.ResourceMaps.Variables {
		m.RootVariables[buildVariableMapKey(k)] = v
	}

	// now traverse all dependency mods
	for _, depMod := range mod.ResourceMaps.Mods {
		// todo for some reason the mod appears in its own resource maps?
		if depMod.Name() != mod.Name() {
			m.DependencyVariables[depMod.DependencyName] = NewModVariableMap(depMod)
		}
	}

	// build map of all publicy settable variables
	m.PopulatePublicVariables()

	return m
}

func (m *ModVariableMap) ToArray() []*Variable {
	var res []*Variable

	keys := utils.SortedMapKeys(m.RootVariables)
	for _, k := range keys {
		res = append(res, m.RootVariables[k])
	}

	for _, depVariables := range m.DependencyVariables {

		keys := utils.SortedMapKeys(depVariables.RootVariables)
		for _, k := range keys {
			res = append(res, depVariables.RootVariables[k])
		}
	}

	return res
}

// build map key fopr root variables - they are keyed by short name
// to allow the user to set their value using the short name
func buildVariableMapKey(k string) string {
	name := strings.TrimPrefix(k, "var.")
	return name
}

// PopulatePublicVariables builds a map of top level and dependency variables
// (dependency variables are keyed by full (qualified) name
func (m *ModVariableMap) PopulatePublicVariables() {
	res := make(map[string]*Variable)
	for k, v := range m.RootVariables {
		// add top level vars keyed by short name
		res[k] = v
	}
	// copy ROOT variables for each top level dependency
	for _, depVars := range m.DependencyVariables {
		for _, v := range depVars.RootVariables {
			// add dependency vars keyed by full name
			res[v.FullName] = v
		}
	}
	m.PublicVariables = res
}
