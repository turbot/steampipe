package controlexecute

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// DirectChildrenModDecorator is a struct used to wrap a Mod but modify the results of GetChildren to only return
// immediate mod children (as opposed to all resources in dependency mods as well)
// This is needed when running 'check all' for a mod which has dependency mopds'
type DirectChildrenModDecorator struct {
	modconfig.ModTreeItemBase
	modconfig.HclResourceBase
	Mod *modconfig.Mod
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

// GetMetadata implements ResourceWithMetadata
func (r DirectChildrenModDecorator) GetMetadata() *modconfig.ResourceMetadata {
	return r.Mod.GetMetadata()
}

// SetMetadata implements ResourceWithMetadata
func (r DirectChildrenModDecorator) SetMetadata(metadata *modconfig.ResourceMetadata) {
	r.Mod.SetMetadata(metadata)
}

// SetAnonymous implements SetAnonymous
func (r DirectChildrenModDecorator) SetAnonymous(block *hcl.Block) {
	r.Mod.SetAnonymous(block)
}

// IsAnonymous implements ResourceWithMetadata
func (r DirectChildrenModDecorator) IsAnonymous() bool {
	return r.Mod.IsAnonymous()
}

// AddReference implements ResourceWithMetadata
func (r DirectChildrenModDecorator) AddReference(ref *modconfig.ResourceReference) {
	r.Mod.AddReference(ref)
}

// GetReferences implements ResourceWithMetadata
func (r DirectChildrenModDecorator) GetReferences() []*modconfig.ResourceReference {
	return r.Mod.GetReferences()
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
