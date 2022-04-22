package modconfig

import (
	"log"
	"strings"
)

type ModVariableMap struct {
	Variables           map[string]*Variable
	DependencyVariables map[string]map[string]*Variable
}

func (m ModVariableMap) AddVariables(variables map[string]*Variable) {
	m.Variables = variables
}

func (m ModVariableMap) AddDependencyVariables(dependencyModName string, variables map[string]*Variable) {
	m.DependencyVariables[dependencyModName] = make(map[string]*Variable)
	for k, v := range variables {
		m.DependencyVariables[dependencyModName][buildVariableMapKey(k)] = v
	}
}

// PromoteUniqueDependencyVariables adds any unique variables of dependency mods into the top level variables
func (m ModVariableMap) PromoteUniqueDependencyVariables() {
	// first construct a count of all variable short names
	variableCountMap := make(map[string]int)
	for _, v := range m.ToArray() {
		variableCountMap[v.ShortName]++
	}
	// now for any dependency varioable with a count of 1, add to top level Variables map
	for _, dep := range m.DependencyVariables {
		for _, v := range dep {
			if variableCountMap[v.ShortName] == 1 {
				log.Printf("[TRACE] variable %s from dependency mod %s is unique in the workspace - adding to Workspace variables",
					v.ShortName, v.ModName)
				m.Variables[v.ShortName] = v
			}
		}
	}
}

func (m ModVariableMap) ToArray() []*Variable {
	var res []*Variable

	for _, v := range m.Variables {
		res = append(res, v)
	}
	for _, dep := range m.DependencyVariables {
		for _, v := range dep {
			res = append(res, v)
		}
	}
	return res
}

// CombinedMap returns a map of top level and dependency variables
// (dependency variables are keyed by full (qualified) name
func (m ModVariableMap) CombinedMap() map[string]*Variable {
	var res = make(map[string]*Variable)
	for k, v := range m.Variables {
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

func NewModVariableMap() *ModVariableMap {
	return &ModVariableMap{
		Variables:           make(map[string]*Variable),
		DependencyVariables: make(map[string]map[string]*Variable),
	}
}

// as the tf derived code builds a map keyed by the short variable name, do the same
// (i.e. strip the "var." from the start
func buildVariableMapKey(k string) string {
	name := strings.TrimPrefix(k, "var.")
	return name
}
