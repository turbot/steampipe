package mod

import (
	"fmt"
	"strings"

	"github.com/turbot/go-kit/helpers"
)

type ModManifest struct {
	Name       string
	Version    string
	ModDepends []*Mod
}

// Mod: this is used to represent both installed mods, which will always have a version,
// and Mods which are to be installed or are a dependency
// - these may have the version "latest" or no version, which is the same as "latest"
type Mod struct {
	Manifest *ModManifest
	Queries  []Query
}

func (m *Mod) FullName() string {
	if m.Manifest.Version == "" {
		return m.Manifest.Name
	}
	return fmt.Sprintf("%s@%s", m.Manifest.Name, m.Manifest.Version)
}

func (m *Mod) String() string {
	depends := []string{}
	for _, d := range m.Manifest.ModDepends {
		depends = append(depends, d.FullName())
	}
	return fmt.Sprintf("Mod: %s, ModDepends: %s", m.String(), strings.Join(depends, ","))
}

// HasVersion :: if no version is specified, or the version is "latest", this is the latest version
func (m *Mod) HasVersion() bool {
	return !helpers.StringSliceContains([]string{"", "latest"}, m.Manifest.Version)
}

// attempt to resolve mod dependencies:
// for each depencency:
// 		1) look for mod in local mod cache
// 		2) if not found try to install mod from registry
// 		3) if loaded, resolve dependencies of the dependency
func (m *Mod) ResolveModDependencies(modMap ModMap, resolving []string) []string {

	// if there are no depedencies we are done
	if m.Manifest.ModDepends == nil {
		return nil
	}

	// 'resolving' is a list of mods we aree trying to resolve in the current tree
	// if we are already resolving this dependency, there is a circular dependency
	if helpers.StringSliceContains(resolving, m.FullName()) {
		// cannot resolve
		return []string{m.FullName()}
	}

	// add this dependency into the list of ongoing resolutions
	resolving = append(resolving, m.FullName())

	// check if we have any versions of this mod in the local cache
	modVersionMap, ok := modMap.GetModVersionMap(m.Manifest.Name)
	if !ok {
		// this mod has not been downloaded - try to download from the registry
		if !LoadFromRegistry(m) {
			return []string{m.FullName()}
		}
	}

	// so we
	// if mod specifies a version, check this version is in the mod map
	if m.HasVersion() {
		if _, ok := modVersionMap[m.Manifest.Version]; !ok {
			return []string{m.FullName()}
		}
	} else {
		// TODO	check whether a newer version is available
	}

	// now resolve each dependency
	missingDependencies := []string{}
	for _, d := range m.Manifest.ModDepends {
		if missing := d.ResolveModDependencies(modMap, resolving); missing != nil {

			// this dependency is missing - request from registry
			_ := LoadFromRegistry(d)

			missingDependencies = append(missingDependencies, missing...)
		}
	}
	return missingDependencies
}

// LoadFromRegistry :: placeholder for loading
func LoadFromRegistry(d *Mod) bool {

	return false
}
