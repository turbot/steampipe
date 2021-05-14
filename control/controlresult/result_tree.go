package controlresult

import (
	"fmt"

	"github.com/turbot/steampipe/workspace"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// ResultTree is a structure representing the control result hierarchy
type ResultTree struct {
	Groups map[string]*ResultGroup
	Root   *ResultGroup
}

// NewResultTree creates a result group from a ControlTreeItem
func NewResultTree(includeControlPredicate func(string) bool, workspace *workspace.Workspace, rootItems ...modconfig.ControlTreeItem) *ResultTree {
	// build tree of result groups, starting with a synthetic 'root' node
	root := NewRootResultGroup(includeControlPredicate, workspace, rootItems...)

	// now populate the ResultTree
	res := &ResultTree{
		Groups: make(map[string]*ResultGroup),
		Root:   root,
	}
	// now populate the map of result groups
	// NOTE: this mutates res.Groups
	res.Root.PopulateGroupMap(res.Groups)

	return res
}

func (t ResultTree) AddResult(result *ControlRun) error {
	// find parent group
	// TODO what if the same control is run by 2 parents?? we need result to know parent
	parents := result.Control.GetParents()
	for _, parent := range parents {
		// find result group with name of parent
		group, ok := t.Groups[parent.Name()]
		// this parent group must exist in the tree
		if ok {
			group.AddResult(result)
			return nil
		}
	}
	return fmt.Errorf("could not find result group for any parents of Control %s", result.Control.Name())
}
