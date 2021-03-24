package mod

// ModMap :: map of mod name to mod-version map
type ModMap map[string]ModVersionMap

func (m ModMap) GetModVersionMap(modName string) (versionMap map[string]*Mod, exists bool) {
	versionMap, exists = m[modName]
	return

}
