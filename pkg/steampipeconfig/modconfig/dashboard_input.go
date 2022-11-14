package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
)

// DashboardInput is a struct representing a leaf dashboard node
type DashboardInput struct {
	ResourceWithMetadataBase
	QueryProviderBase

	DashboardName string `column:"dashboard,text" json:"-"`

	Label       *string                 `cty:"label" hcl:"label" column:"label,text" json:"label,omitempty"`
	Placeholder *string                 `cty:"placeholder" hcl:"placeholder" column:"placeholder,text" json:"placeholder,omitempty"`
	Options     []*DashboardInputOption `cty:"options" hcl:"option,block" json:"options,omitempty"`

	// these properties are JSON serialised by the parent LeafRun
	Width      *int                 `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type       *string              `cty:"type" hcl:"type" column:"type,text" json:"-"`
	Display    *string              `cty:"display" hcl:"display" json:"-"`
	Base       *DashboardInput      `hcl:"base" json:"-"`
	References []*ResourceReference `json:"-"`

	Paths     []NodePath `column:"path,jsonb" json:"-"`
	parents   []ModTreeItem
	dashboard *Dashboard
}

func NewDashboardInput(block *hcl.Block, mod *Mod, shortName string) HclResource {
	// input cannot be anonymous
	i := &DashboardInput{
		QueryProviderBase: QueryProviderBase{
			Mod: mod,
			HclResourceBase: HclResourceBase{
				ShortName:       shortName,
				FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
				UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
				DeclRange:       block.DefRange,
			},
		},
	}
	return i
}

func (i *DashboardInput) Clone() *DashboardInput {
	return &DashboardInput{
		ResourceWithMetadataBase: i.ResourceWithMetadataBase,
		QueryProviderBase:        i.QueryProviderBase,
		Width:                    i.Width,
		Type:                     i.Type,
		Label:                    i.Label,
		Placeholder:              i.Placeholder,
		Display:                  i.Display,
		Options:                  i.Options,
		Paths:                    i.Paths,
		parents:                  i.parents,
		dashboard:                i.dashboard,
	}
}

func (i *DashboardInput) Equals(other *DashboardInput) bool {
	diff := i.Diff(other)
	return !diff.HasChanges()
}

// IsSnapshotPanel implements SnapshotPanel
func (*DashboardInput) IsSnapshotPanel() {}

// OnDecoded implements HclResource
func (i *DashboardInput) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	i.setBaseProperties(resourceMapProvider)
	return nil
}

// AddReference implements ResourceWithMetadata
func (i *DashboardInput) AddReference(ref *ResourceReference) {
	i.References = append(i.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (i *DashboardInput) GetReferences() []*ResourceReference {
	return i.References
}

// AddParent implements ModTreeItem
func (i *DashboardInput) AddParent(parent ModTreeItem) error {
	i.parents = append(i.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (i *DashboardInput) GetParents() []ModTreeItem {
	return i.parents
}

// GetChildren implements ModTreeItem
func (i *DashboardInput) GetChildren() []ModTreeItem {
	return nil
}

// GetPaths implements ModTreeItem
func (i *DashboardInput) GetPaths() []NodePath {
	// lazy load
	if len(i.Paths) == 0 {
		i.SetPaths()
	}

	return i.Paths
}

// SetPaths implements ModTreeItem
func (i *DashboardInput) SetPaths() {
	for _, parent := range i.parents {
		for _, parentPath := range parent.GetPaths() {
			i.Paths = append(i.Paths, append(parentPath, i.Name()))
		}
	}
}

func (i *DashboardInput) Diff(other *DashboardInput) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: i,
		Name: i.Name(),
	}

	if !utils.SafeStringsEqual(i.Type, other.Type) {
		res.AddPropertyDiff("Type")
	}

	if !utils.SafeStringsEqual(i.Label, other.Label) {
		res.AddPropertyDiff("Label")
	}

	if !utils.SafeStringsEqual(i.Placeholder, other.Placeholder) {
		res.AddPropertyDiff("Placeholder")
	}

	if len(i.Options) != len(other.Options) {
		res.AddPropertyDiff("Options")
	} else {
		for idx, o := range i.Options {
			if !other.Options[idx].Equals(o) {
				res.AddPropertyDiff("Options")
			}
		}
	}

	res.populateChildDiffs(i, other)
	res.queryProviderDiff(i, other)
	res.dashboardLeafNodeDiff(i, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (i *DashboardInput) GetWidth() int {
	if i.Width == nil {
		return 0
	}
	return *i.Width
}

// GetDisplay implements DashboardLeafNode
func (i *DashboardInput) GetDisplay() string {
	return typehelpers.SafeString(i.Display)
}

// GetDocumentation implements DashboardLeafNode, ModTreeItem
func (*DashboardInput) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (i *DashboardInput) GetType() string {
	return typehelpers.SafeString(i.Type)
}

// SetDashboard sets the parent dashboard container
func (i *DashboardInput) SetDashboard(dashboard *Dashboard) {
	i.dashboard = dashboard
	i.DashboardName = dashboard.Name()
}

// VerifyQuery implements QueryProvider
func (i *DashboardInput) VerifyQuery(QueryProvider) error {
	// query is optional - nothing to do
	return nil
}

// DependsOnInput returns whether this input has a runtime dependency on the given input
func (i *DashboardInput) DependsOnInput(changedInputName string) bool {
	for _, r := range i.runtimeDependencies {
		if r.SourceResource.GetUnqualifiedName() == changedInputName {
			return true
		}
	}
	return false
}

func (i *DashboardInput) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(i.Base, resourceMapProvider); !resolved {
		return
	} else {
		i.Base = base.(*DashboardInput)
	}

	if i.Title == nil {
		i.Title = i.Base.Title
	}

	if i.Type == nil {
		i.Type = i.Base.Type
	}

	if i.Display == nil {
		i.Display = i.Base.Display
	}

	if i.Label == nil {
		i.Label = i.Base.Label
	}

	if i.Placeholder == nil {
		i.Placeholder = i.Base.Placeholder
	}

	if i.Width == nil {
		i.Width = i.Base.Width
	}

	if i.SQL == nil {
		i.SQL = i.Base.SQL
	}

	if i.Query == nil {
		i.Query = i.Base.Query
	}

	if i.Args == nil {
		i.Args = i.Base.Args
	}

	if i.Params == nil {
		i.Params = i.Base.Params
	}

	i.MergeRuntimeDependencies(i.Base)
}
