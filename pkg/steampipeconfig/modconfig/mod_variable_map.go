package modconfig

import (
	"strings"

	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/steampipe/pkg/utils"
)

// ModVariableMap is a struct containins maps of variable definitions
type ModVariableMap struct {
	// which mod have these variables been loaded for?
	Mod *Mod
	// top level variables, keyed by short name
	RootVariables map[string]*Variable
	// map of dependency variable maps, keyed by dependency NAME
	DependencyVariables map[string]*ModVariableMap

	// a list of the pointers to the variables whose values can be changed
	// NOTE: this refers to the SAME variable objects as exist in the RootVariables and DependencyVariables maps,
	// so when we set the value of public variables, we mutate the underlying variable
	PublicVariables map[string]*Variable
}

// NewModVariableMap builds a ModVariableMap using the variables from a mod and its dependencies
func NewModVariableMap(mod *Mod) *ModVariableMap {
	m := &ModVariableMap{
		Mod:                 mod,
		RootVariables:       make(map[string]*Variable),
		DependencyVariables: make(map[string]*ModVariableMap),
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

	// build map of all publicly settable variables
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

// GetPublicVariableValues converts public variables into a map of string variable values
func (m *ModVariableMap) GetPublicVariableValues() (map[string]string, error) {
	res := make(map[string]string, len(m.PublicVariables))
	for k, v := range m.PublicVariables {
		// TODO investigate workspace usage of value string and determine whether we can simply format ValueGo
		valueString, err := hclhelpers.CtyToString(v.Value)
		if err != nil {
			return nil, err
		}
		res[k] = valueString
	}
	return res, nil
}
