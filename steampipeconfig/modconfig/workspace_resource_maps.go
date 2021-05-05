package modconfig

// WorkspaceResourceMaps :: maps of all mod resource types
// provided to avoid db needing to reference workspace package
type WorkspaceResourceMaps struct {
	ModMap       map[string]*Mod
	QueryMap     map[string]*Query
	ControlMap   map[string]*Control
	BenchmarkMap map[string]*Benchmark
}
