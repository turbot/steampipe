package modconfig

// WorkspaceResourceMaps is a struct containing maps of all mod resource types
// This is provided to avoid db needing to reference workspace package
type WorkspaceResourceMaps struct {
	ModMap       map[string]*Mod
	QueryMap     map[string]*Query
	ControlMap   map[string]*Control
	BenchmarkMap map[string]*Benchmark
}

func (m WorkspaceResourceMaps) Equals(other *WorkspaceResourceMaps) bool {
	for name, mod := range m.ModMap {
		if otherMod, ok := other.ModMap[name]; !ok {
			return false
		} else if !mod.Equals(otherMod) {
			return false
		}
	}
	for name := range other.ModMap {
		if _, ok := m.ModMap[name]; !ok {
			return false
		}
	}
	for name, control := range m.ControlMap {
		if otherControl, ok := other.ControlMap[name]; !ok {
			return false
		} else if !control.Equals(otherControl) {
			return false
		}
	}
	for name := range other.ControlMap {
		if _, ok := m.ControlMap[name]; !ok {
			return false
		}
	}
	for name, query := range m.QueryMap {
		if otherQuery, ok := other.QueryMap[name]; !ok {
			return false
		} else if !query.Equals(otherQuery) {
			return false
		}
	}

	for name := range other.QueryMap {
		if _, ok := m.QueryMap[name]; !ok {
			return false
		}
	}
	return true
}
