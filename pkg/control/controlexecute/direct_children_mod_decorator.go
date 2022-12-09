package controlexecute

import (
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// DirectChildrenModDecorator is a struct used to wrap a Mod but modify the results of GetChildren to only return
// immediate mod children (as opposed to all resources in dependency mods as well)
// This is needed when running 'check all' for a mod which has dependency mopds'
type DirectChildrenModDecorator struct {
	*modconfig.Mod
}

// override GetChildren
func (r DirectChildrenModDecorator) GetChildren() []modconfig.ModTreeItem {
	var res []modconfig.ModTreeItem
	for _, child := range r.Mod.GetChildren() {
		if child.GetMod().ShortName == r.Mod.ShortName {
			res = append(res, child)
		}
	}
	return res
}
