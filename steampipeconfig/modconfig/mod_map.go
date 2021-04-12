package modconfig

import (
	"strings"
)

// ModMap :: map of mod name to mod-version map
type ModMap map[string]*Mod

func (m ModMap) String() string {
	var modStrings []string
	for _, mod := range m {
		modStrings = append(modStrings, mod.String())
	}
	return strings.Join(modStrings, "\n")
}
