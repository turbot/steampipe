package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
)

// DashboardImage is a struct representing a leaf dashboard node
type DashboardImage struct {
	ResourceWithMetadataBase
	QueryProviderBase
	ModTreeItemBase

	Src *string `cty:"src" hcl:"src" column:"src,text" json:"src,omitempty"`
	Alt *string `cty:"alt" hcl:"alt" column:"alt,text" json:"alt,omitempty"`

	// these properties are JSON serialised by the parent LeafRun
	Width   *int    `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Display *string `cty:"display" hcl:"display" json:"-"`

	Base       *DashboardImage      `hcl:"base" json:"-"`
	References []*ResourceReference `json:"-"`
}

func NewDashboardImage(block *hcl.Block, mod *Mod, shortName string) HclResource {
	i := &DashboardImage{
		QueryProviderBase: QueryProviderBase{
			modNameWithVersion: mod.NameWithVersion(),
			HclResourceBase: HclResourceBase{
				ShortName:       shortName,
				FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
				UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
		},
	}
	i.SetAnonymous(block)
	return i
}

func (i *DashboardImage) Equals(other *DashboardImage) bool {
	diff := i.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource
func (i *DashboardImage) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	i.setBaseProperties(resourceMapProvider)
	return nil
}

// AddReference implements ResourceWithMetadata
func (i *DashboardImage) AddReference(ref *ResourceReference) {
	i.References = append(i.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (i *DashboardImage) GetReferences() []*ResourceReference {
	return i.References
}

func (i *DashboardImage) Diff(other *DashboardImage) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: i,
		Name: i.Name(),
	}
	if !utils.SafeStringsEqual(i.Src, other.Src) {
		res.AddPropertyDiff("Src")
	}

	if !utils.SafeStringsEqual(i.Alt, other.Alt) {
		res.AddPropertyDiff("Alt")
	}

	res.populateChildDiffs(i, other)
	res.queryProviderDiff(i, other)
	res.dashboardLeafNodeDiff(i, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (i *DashboardImage) GetWidth() int {
	if i.Width == nil {
		return 0
	}
	return *i.Width
}

// GetDisplay implements DashboardLeafNode
func (i *DashboardImage) GetDisplay() string {
	return typehelpers.SafeString(i.Display)
}

// GetDocumentation implements DashboardLeafNode, ModTreeItem
func (*DashboardImage) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (*DashboardImage) GetType() string {
	return ""
}

// VerifyQuery implements QueryProvider
func (i *DashboardImage) VerifyQuery(QueryProvider) error {
	// query is optional - nothing to do
	return nil
}

func (i *DashboardImage) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(i.Base, resourceMapProvider); !resolved {
		return
	} else {
		i.Base = base.(*DashboardImage)
	}

	if i.Title == nil {
		i.Title = i.Base.Title
	}

	if i.Src == nil {
		i.Src = i.Base.Src
	}

	if i.Alt == nil {
		i.Alt = i.Base.Alt
	}

	if i.Width == nil {
		i.Width = i.Base.Width
	}

	if i.Display == nil {
		i.Display = i.Base.Display
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
