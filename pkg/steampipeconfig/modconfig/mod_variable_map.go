package modconfig

import (
	"sort"
	"strings"
)

// ModVariableMap is a struct containins maps of variable definitions
type ModVariableMap struct {
	RootVariables       map[string]*Variable
	DependencyVariables map[string]map[string]*Variable
	// a map of top level AND dependency variables
	// used to set variable values from inputVariables
	AllVariables map[string]*Variable
	// the input variables evaluated in the parse
	VariableValues map[string]string
}

// NewModVariableMap builds a ModVariableMap using the variables from a mod and its dependencies
func NewModVariableMap(mod *Mod, dependencyMods ModMap) *ModVariableMap {
	m := &ModVariableMap{
		RootVariables:       make(map[string]*Variable),
		DependencyVariables: make(map[string]map[string]*Variable),
		VariableValues:      make(map[string]string),
	}

	// add variables into map, modifying the key to be the variable short name
	for k, v := range mod.ResourceMaps.Variables {
		m.RootVariables[buildVariableMapKey(k)] = v
	}
	// now add variables from dependency mods
	for _, mod := range dependencyMods {
		// add variables into map, modifying the key to be the variable short name
		m.DependencyVariables[mod.ShortName] = make(map[string]*Variable)
		for k, v := range mod.ResourceMaps.Variables {
			m.DependencyVariables[mod.ShortName][buildVariableMapKey(k)] = v
		}
	}
	// build map of all variables
	m.AllVariables = m.buildCombinedMap()

	return m
}

// build a map of top level and dependency variables
// (dependency variables are keyed by full (qualified) name
func (m ModVariableMap) buildCombinedMap() map[string]*Variable {
	res := make(map[string]*Variable)
	for k, v := range m.RootVariables {
		// add top level vars keyed by short name
		res[k] = v
	}
	for _, dep := range m.DependencyVariables {
		for _, v := range dep {
			// add dependency vars keyed by full name
			res[v.FullName] = v
		}
	}
	return res
}

func (m ModVariableMap) ToArray() []*Variable {
	var res []*Variable

	if len(m.AllVariables) > 0 {
		var keys []string

		for k := range m.RootVariables {
			keys = append(keys, k)
		}
		// sort keys
		sort.Strings(keys)
		for _, k := range keys {
			res = append(res, m.RootVariables[k])
		}
	}

	for _, depVariables := range m.DependencyVariables {
		if len(depVariables) == 0 {
			continue
		}
		keys := make([]string, len(depVariables))
		idx := 0

		for k := range depVariables {
			keys[idx] = k
			idx++
		}
		// sort keys
		sort.Strings(keys)
		for _, k := range keys {
			res = append(res, depVariables[k])
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
