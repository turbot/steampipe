package modconfig

import (
	"strings"
)

// ModMap is a map of mod name to mod
type ModMap map[string]*Mod

func (m ModMap) String() string {
	var modStrings []string
	for _, mod := range m {
		modStrings = append(modStrings, mod.String())
	}
	return strings.Join(modStrings, "\n")
}
