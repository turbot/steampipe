package versionmap

func (l *WorkspaceLock) GetModList(rootName string) string {

	tree := l.InstallCache.GetDependencyTree(rootName)
	return tree.String()
}
