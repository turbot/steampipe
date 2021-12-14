package version_map

import (
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/xlab/treeprint"
)

func (l *WorkspaceLock) GetModList(rootName string) string {
	if len(l.InstallCache) == 0 {
		return "No mods installed"
	}

	tree := treeprint.NewWithRoot(rootName)
	l.buildTree(rootName, tree)
	return tree.String()
}

func (l *WorkspaceLock) buildTree(name string, tree treeprint.Tree) {
	deps := l.InstallCache[name]
	for name, version := range deps {
		fullName := modconfig.ModVersionFullName(name, version.Version)
		child := tree.AddBranch(fullName)
		// if there are children add them
		l.buildTree(fullName, child)
	}
}
