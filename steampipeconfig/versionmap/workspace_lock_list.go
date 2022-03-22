package versionmap

func (l *WorkspaceLock) GetModList(rootName string) string {
	if len(l.InstallCache) == 0 {
		return "No mods installed"
	}

	tree := l.InstallCache.GetDependencyTree(rootName)
	return tree.String()
}
