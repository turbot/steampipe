package modconfig

import (
	"fmt"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
)

// DashboardWith is a struct representing a leaf dashboard node
type DashboardWith struct {
	ResourceWithMetadataBase
	QueryProviderBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Base       *DashboardWith       `hcl:"base" json:"-"`
	References []*ResourceReference `json:"-"`
}

func NewDashboardWith(block *hcl.Block, mod *Mod, shortName string) HclResource {
	// with blocks cannot be anonymous
	c := &DashboardWith{
		QueryProviderBase: QueryProviderBase{
			RuntimeDependencyProviderBase: RuntimeDependencyProviderBase{
				ModTreeItemBase: ModTreeItemBase{
					HclResourceBase: HclResourceBase{
						ShortName:       shortName,
						FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
						UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
						DeclRange:       block.DefRange,
						blockType:       block.Type,
					},
					Mod: mod,
				},
			},
		},
	}

	return c
}

func (e *DashboardWith) Equals(other *DashboardWith) bool {
	diff := e.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource
func (e *DashboardWith) OnDecoded(_ *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	e.setBaseProperties(resourceMapProvider)

	return nil
}

func (e *DashboardWith) Diff(other *DashboardWith) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: e,
		Name: e.Name(),
	}

	res.queryProviderDiff(e, other)

	return res
}

// IsSnapshotPanel implements SnapshotPanel
func (*DashboardWith) IsSnapshotPanel() {}

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
func (t *DashboardWith) CtyValue() (cty.Value, error) {
	return GetCtyValue(t)
}

func (e *DashboardWith) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(e.Base, resourceMapProvider); !resolved {
		return
	} else {
		e.Base = base.(*DashboardWith)
	}

	if e.Title == nil {
		e.Title = e.Base.Title
	}

	if e.SQL == nil {
		e.SQL = e.Base.SQL
	}

	if e.Query == nil {
		e.Query = e.Base.Query
	}

	if e.Args == nil {
		e.Args = e.Base.Args
	}

	// only inherit params if top level
	if e.Params == nil && e.isTopLevel {
		e.Params = e.Base.Params
	}

	e.MergeRuntimeDependencies(e.Base)
}
