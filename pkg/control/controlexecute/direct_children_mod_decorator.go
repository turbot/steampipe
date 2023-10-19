package controlexecute

import (
	"github.com/turbot/pipe-fittings/modconfig"
)

// DirectChildrenModDecorator is a struct used to wrap a Mod but modify the results of GetChildren to only return
// immediate mod children (as opposed to all resources in dependency mods as well)
// This is needed when running 'check all' for a mod which has dependency mopds'
type DirectChildrenModDecorator struct {
	*modconfig.Mod
}

// GetChildren is overridden
func (r DirectChildrenModDecorator) GetChildren() []modconfig.ModTreeItem {
	var res []modconfig.ModTreeItem
	for _, child := range r.Mod.GetChildren() {
		if child.GetMod().ShortName == r.Mod.ShortName {
			res = append(res, child)
		}
	}
	return res
}

// GetDocumentation implements DashboardLeafNode
func (r DirectChildrenModDecorator) GetDocumentation() string {
	return r.Mod.GetDocumentation()
}

// GetDisplay implements DashboardLeafNode
func (r DirectChildrenModDecorator) GetDisplay() string {
	return ""
}

// GetType implements DashboardLeafNode
func (r DirectChildrenModDecorator) GetType() string {
	return ""
}

// GetWidth implements DashboardLeafNode
func (r DirectChildrenModDecorator) GetWidth() int {
	return 0
}
