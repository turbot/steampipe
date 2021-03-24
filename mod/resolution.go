package mod

import (
	"fmt"
	"strings"
)

// the result or resolving a set of mods
type Resolution struct {
	Resolved   map[string]*Mod
	Unresolved map[string]*UnresolvedMod
}

func (r *Resolution) String() string {
	var strs []string
	for _, resolved := range r.Resolved {
		strs = append(strs, resolved.String())
	}
	for _, unresolved := range r.Unresolved {
		strs = append(strs, unresolved.String())
	}
	return strings.Join(strs, "\n")
}

// UnresolvedMod :: a struct representing a mod for which the dependencies could NOT be resolved
type UnresolvedMod struct {
	Mod *Mod
	// the reason dependencies could not be resolved
	// TODO replace with 'MissingDependencies'
	Message string
}

func (u *UnresolvedMod) String() string {
	return fmt.Sprintf("%s\n%s", u.Mod.String(), u.Message)
}

func ResolveModDependencies(mods []*Mod) *Resolution {
	res := NewResolution()

	modVersionMap := buildModMap(mods)

	// now try to resolve all mod dependencies
	for _, mod := range mods {
		// we may have already attempted to resolve this mod - if it is someone elses dependency
		_, alreadyResolved := res.Resolved[mod.FullName()]
		_, alreadyUnresolved := res.Unresolved[mod.FullName()]
		if alreadyResolved || alreadyUnresolved {
			continue
		}
		failedDependencies := mod.ResolveModDependencies(modVersionMap, []string{})
		if len(failedDependencies) == 0 {
			res.Resolved[mod.FullName()] = mod
		} else {

			res.Unresolved[mod.FullName()] = &UnresolvedMod{
				Mod:     mod,
				Message: fmt.Sprintf("failed to resolve dependencies: %s", strings.Join(failedDependencies, ",")),
			}

		}
	}
	return res
}

func NewResolution() *Resolution {
	res := &Resolution{
		Resolved:   make(map[string]*Mod),
		Unresolved: make(map[string]*UnresolvedMod),
	}
	return res
}

// build map of mod name to mod version maps
func buildModMap(mods []*Mod) ModMap {
	modMap := ModMap{}
	for _, mod := range mods {
		versionMap, ok := modMap.GetModVersionMap(mod.Manifest.Name)
		if !ok {
			versionMap = make(map[string]*Mod)
		}
		versionMap[mod.Manifest.Version] = mod
		modMap[mod.Manifest.Name] = versionMap
	}
	return modMap
}
