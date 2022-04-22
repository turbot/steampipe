package modconfig

import (
	"log"
	"sort"
	"strings"
)

type ModVariableMap struct {
	RootVariables       map[string]*Variable
	DependencyVariables map[string]map[string]*Variable
	// a map of top level AND dependency variables
	// used to set variable values from inputVariables
	AllVariables map[string]*Variable

	// a map of promoted dependency variables
	// we use this to ensure the promoted variable and the original variable both get their value set if
	// the value is passed in
	// key - variable short name, value - variable full name
	VariableAliases map[string]string
}

func NewModVariableMap(mod *Mod, dependencyMods ModMap) *ModVariableMap {
	m := &ModVariableMap{
		RootVariables:       make(map[string]*Variable),
		DependencyVariables: make(map[string]map[string]*Variable),
		VariableAliases:     make(map[string]string),
	}

	// add variables into map, modifying th ekey to be the variable short name
	for k, v := range mod.ResourceMaps.Variables {
		m.RootVariables[buildVariableMapKey(k)] = v
	}
	// now add variables from dependency mods
	for _, mod := range dependencyMods {
		// add variables into map, modifying th ekey to be the variable short name
		m.DependencyVariables[mod.ShortName] = make(map[string]*Variable)
		for k, v := range mod.ResourceMaps.Variables {
			m.DependencyVariables[mod.ShortName][buildVariableMapKey(k)] = v
		}
	}
	// add any unique variables of dependency mods into the top level variables
	// this allows users to reference (and set values of) variables in dependency mods without qualifying them
	m.promoteUniqueDependencyVariables()

	// build map of all variables
	m.AllVariables = m.buildCombinedMap()

	return m
}

// promoteUniqueDependencyVariables adds any unique variables of dependency mods into the top level variables
func (m ModVariableMap) promoteUniqueDependencyVariables() {
	// first construct a count of all variable short names
	variableCountMap := make(map[string]int)
	for _, v := range m.ToArray() {
		variableCountMap[v.ShortName]++
	}
	// now for any dependency variable with a count of 1, add to RootVariables
	for _, dep := range m.DependencyVariables {
		for _, v := range dep {
			if variableCountMap[v.ShortName] == 1 {
				log.Printf("[TRACE] variable %s from dependency mod %s is unique in the workspace - adding to Workspace variables",
					v.ShortName, v.ModName)
				m.RootVariables[v.ShortName] = v
				// also add to aliases
				m.VariableAliases[v.ShortName] = v.Name()
			}
		}
	}
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
		var keys = make([]string, len(m.RootVariables)-len(m.VariableAliases))
		idx := 0

		for k := range m.RootVariables {
			// if there is an alias for this variable, that means it is will appear twice in AllVariables,
			// so exclude this copy
			if _, ok := m.VariableAliases[k]; ok {
				continue
			}
			keys[idx] = k
			idx++
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
