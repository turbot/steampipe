package mod

import (
	"strings"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// ModMap :: map of mod name to mod-version map
type ModMap map[string]*modconfig.Mod

func (m ModMap) String() string {
	var modStrings []string
	for _, mod := range m {
		modStrings = append(modStrings, mod.String())
	}
	return strings.Join(modStrings, "\n")
}

//
//// attempt to resolve mod dependencies:
//// for each depencency:
//// 		1) look for mod in local mod cache
//// 		2) if not found try to install mod from registry
//// 		3) if loaded, resolve dependencies of the dependency
//func (modMap ModMap) ResolveModDependencies(m *modconfig.Mod, resolving []string) []string {
//
//	// if there are no depedencies we are done
//	if m.ModDepends == nil {
//		return nil
//	}
//
//	// 'resolving' is a list of mods we aree trying to resolve in the current tree
//	// if we are already resolving this dependency, there is a circular dependency
//	if helpers.StringSliceContains(resolving, m.FullName()) {
//		// cannot resolve
//		return []string{m.FullName()}
//	}
//
//	// add this dependency into the list of ongoing resolutions
//	resolving = append(resolving, m.FullName())
//
//	// check if we have any versions of this mod in the local cache
//	modVersionMap, ok := modMap.GetModVersionMap(m.Manifest.Name)
//	if !ok {
//		// this mod has not been downloaded - try to download from the registry
//		if !LoadFromRegistry(m) {
//			return []string{m.FullName()}
//		}
//	}
//
//	// so we
//	// if mod specifies a version, check this version is in the mod map
//	if m.HasVersion() {
//		if _, ok := modVersionMap[m.Manifest.Version]; !ok {
//			return []string{m.FullName()}
//		}
//	} else {
//		// TODO	check whether a newer version is available
//	}
//
//	// now resolve each dependency
//	missingDependencies := []string{}
//	for _, d := range m.Manifest.ModDepends {
//		if missing := modMap.ResolveModDependencies(d, resolving); missing != nil {
//
//			// this dependency is missing - request from registry
//			mod := LoadFromRegistry(d)
//
//			missingDependencies = append(missingDependencies, missing...)
//		}
//	}
//	return missingDependencies
//}

// LoadFromRegistry :: placeholder for loading
func LoadFromRegistry(d *modconfig.Mod) bool {

	return false
}
