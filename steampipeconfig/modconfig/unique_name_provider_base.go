package modconfig

import "fmt"

type UniqueNameProviderBase struct {
	nameMap map[string]struct{}
}

// GetUniqueName returns a name unique within the scope of this execution tree
func (p *UniqueNameProviderBase) GetUniqueName(name string) string {
	if p.nameMap == nil {
		p.nameMap = make(map[string]struct{})
	}
	// keep adding a suffix until we get a unique name
	uniqueName := name
	suffix := 0
	for {
		if _, ok := p.nameMap[uniqueName]; !ok {
			p.nameMap[uniqueName] = struct{}{}
			return uniqueName
		}
		suffix++
		uniqueName = fmt.Sprintf("%s_%d", name, suffix)
	}
	return uniqueName
}
