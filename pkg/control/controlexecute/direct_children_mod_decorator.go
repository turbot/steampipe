package controlexecute

import "github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"

// DirectChildrenModDecorator is a struct used to wrap a Mod but modify the results of GetChildren to only return
// immediate mod children (as opposed to all resources in dependency mods as well)
// This is needed when running 'check all' for a mod which has dependency mopds'
type DirectChildrenModDecorator struct {
	Mod *modconfig.Mod
}

func (r DirectChildrenModDecorator) AddParent(item modconfig.ModTreeItem) error {
	return nil
}

func (r DirectChildrenModDecorator) GetChildren() []modconfig.ModTreeItem {
	var res []modconfig.ModTreeItem
	for _, child := range r.Mod.GetChildren() {
		if child.GetMod().ShortName == r.Mod.ShortName {
			res = append(res, child)
		}
	}
	return res
}

func (r DirectChildrenModDecorator) Name() string {
	return r.Mod.Name()
}

func (r DirectChildrenModDecorator) GetUnqualifiedName() string {
	return r.Mod.GetUnqualifiedName()
}

func (r DirectChildrenModDecorator) GetTitle() string {
	return r.Mod.GetTitle()
}

func (r DirectChildrenModDecorator) GetDescription() string {
	return r.Mod.GetDescription()
}

func (r DirectChildrenModDecorator) GetTags() map[string]string {
	return r.Mod.GetTags()
}

func (r DirectChildrenModDecorator) GetPaths() []modconfig.NodePath {
	return r.Mod.GetPaths()
}

func (r DirectChildrenModDecorator) SetPaths() {
	r.Mod.SetPaths()
}

func (r DirectChildrenModDecorator) GetMod() *modconfig.Mod {
	return r.Mod
}
