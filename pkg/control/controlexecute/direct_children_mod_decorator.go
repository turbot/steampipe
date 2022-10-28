package controlexecute

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/zclconf/go-cty/cty"
)

// DirectChildrenModDecorator is a struct used to wrap a Mod but modify the results of GetChildren to only return
// immediate mod children (as opposed to all resources in dependency mods as well)
// This is needed when running 'check all' for a mod which has dependency mopds'
type DirectChildrenModDecorator struct {
	Mod *modconfig.Mod
}

func (r DirectChildrenModDecorator) AddParent(item modconfig.ModTreeItem) error {
	return nil
}
func (r DirectChildrenModDecorator) GetParents() []modconfig.ModTreeItem {
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

// GetDocumentation implements DashboardLeafNode, ModTreeItem
func (r DirectChildrenModDecorator) GetDocumentation() string {
	return r.Mod.GetDocumentation()
}

func (r DirectChildrenModDecorator) GetMod() *modconfig.Mod {
	return r.Mod
}

// BlockType implements HclResource
func (r DirectChildrenModDecorator) BlockType() string {
	return r.Mod.BlockType()

}

// CtyValue implements HclResource
func (r DirectChildrenModDecorator) CtyValue() (cty.Value, error) {
	return r.Mod.CtyValue()
}

// GetDeclRange implements HclResource
func (r DirectChildrenModDecorator) GetDeclRange() *hcl.Range {
	return &r.Mod.DeclRange
}

// OnDecoded implements HclResource
func (r DirectChildrenModDecorator) OnDecoded(block *hcl.Block, resourceMapProvider modconfig.ResourceMapsProvider) hcl.Diagnostics {
	return nil
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
