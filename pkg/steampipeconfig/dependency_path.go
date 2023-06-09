package steampipeconfig

import "strings"

const pathSeparator = " -> "

// DependencyPathKey is a string representation of a dependency path
//   - a set of mod dependencyPath values separated by '->'
//
// e.g. local -> github.com/kaidaguerre/steampipe-mod-m1@v3.1.1 -> github.com/kaidaguerre/steampipe-mod-m2@v5.1.1
type DependencyPathKey string

func newDependencyPathKey(dependencyPath ...string) DependencyPathKey {
	return DependencyPathKey(strings.Join(dependencyPath, pathSeparator))
}

func (k DependencyPathKey) GetParent() DependencyPathKey {
	elements := strings.Split(string(k), pathSeparator)
	if len(elements) == 1 {
		return ""
	}
	return newDependencyPathKey(elements[:len(elements)-2]...)
}

// how long is the depdency path
func (k DependencyPathKey) PathLength() int {
	return len(strings.Split(string(k), pathSeparator))
}
