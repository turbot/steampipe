package modconfig

import (
	"fmt"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
)

// DashboardWith is a struct representing a leaf dashboard node
type DashboardWith struct {
	ResourceWithMetadataImpl
	QueryProviderImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`
}

func NewDashboardWith(block *hcl.Block, mod *Mod, shortName string) HclResource {
	// with blocks cannot be anonymous
	c := &DashboardWith{
		QueryProviderImpl: QueryProviderImpl{
			RuntimeDependencyProviderImpl: RuntimeDependencyProviderImpl{
				ModTreeItemImpl: ModTreeItemImpl{
					HclResourceImpl: HclResourceImpl{
						ShortName:       shortName,
						FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
						UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
						DeclRange:       BlockRange(block),
						blockType:       block.Type,
					},
					Mod: mod,
				},
			},
		},
	}

	return c
}

func (w *DashboardWith) Equals(other *DashboardWith) bool {
	diff := w.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource
func (w *DashboardWith) OnDecoded(_ *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	return nil
}

func (w *DashboardWith) Diff(other *DashboardWith) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: w,
		Name: w.Name(),
	}

	res.queryProviderDiff(w, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (*DashboardWith) GetWidth() int {
	return 0
}

// GetDisplay implements DashboardLeafNode
func (*DashboardWith) GetDisplay() string {
	return ""
}

// GetType implements DashboardLeafNode
func (*DashboardWith) GetType() string {
	return ""
}

// CtyValue implements CtyValueProvider
func (w *DashboardWith) CtyValue() (cty.Value, error) {
	return GetCtyValue(w)
}
