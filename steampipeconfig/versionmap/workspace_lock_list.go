package versionmap

func (l *WorkspaceLock) GetModList(rootName string) string {
	if len(l.installCache) == 0 {
		return "No mods installed"
	}

	tree := l.installCache.GetDependencyTree(rootName)
	return tree.String()
}
