package modconfig

// WorkspaceResourceMaps is a struct containing maps of all mod resource types
// This is provided to avoid db needing to reference workspace package
type WorkspaceResourceMaps struct {
	ModMap       map[string]*Mod
	QueryMap     map[string]*Query
	ControlMap   map[string]*Control
	BenchmarkMap map[string]*Benchmark
}
