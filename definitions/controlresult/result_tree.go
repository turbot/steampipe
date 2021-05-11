package controlresult

import (
	"fmt"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// ResultTree is a structure representing the control result hierarchy
type ResultTree struct {
	Groups map[string]*ResultGroup
	Root   *ResultGroup
}

// NewResultTree creates a result group from a ControlTreeItem
func NewResultTree(rootItem modconfig.ControlTreeItem) *ResultTree {
	res := &ResultTree{
		Groups: make(map[string]*ResultGroup),
		Root:   NewResultGroup(rootItem),
	}

	res.Root.GetGroupMap(res.Groups)

	return res
}

func (t ResultTree) AddResult(result *Result) error {
	// find parent group
	parents := result.Control.GetParents()
	for _, parent := range parents {
		// find result group with name of parent
		group, ok := t.Groups[parent.Name()]
		// this parent group must exist in the tree
		if ok {
			group.Results = append(group.Results, result)
			return nil
		}
	}
	return fmt.Errorf("could not find result group for any parents of Control %s", result.Control.Name())
}
