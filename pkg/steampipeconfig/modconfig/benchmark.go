package modconfig

import (
	"fmt"
	"github.com/zclconf/go-cty/cty"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/types"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
)

// Benchmark is a struct representing the Benchmark resource
type Benchmark struct {
	ResourceWithMetadataImpl
	ModTreeItemImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	// child names as NamedItem structs - used to allow setting children via the 'children' property
	ChildNames NamedItemList `cty:"child_names" json:"-"`
	// used for introspection tables
	ChildNameStrings []string `cty:"child_name_strings" column:"children,jsonb" json:"-"`

	// dashboard specific properties
	Base    *Benchmark `hcl:"base" json:"-"`
	Width   *int       `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type    *string    `cty:"type" hcl:"type" column:"type,text" json:"-"`
	Display *string    `cty:"display" hcl:"display" json:"-"`
}

func NewBenchmark(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)
	benchmark := &Benchmark{
		ModTreeItemImpl: ModTreeItemImpl{
			HclResourceImpl: HclResourceImpl{
				ShortName:       shortName,
				FullName:        fullName,
				UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
				DeclRange:       BlockRange(block),
				blockType:       block.Type,
			},
			Mod: mod,
		},
	}
	benchmark.SetAnonymous(block)
	return benchmark
}

func (b *Benchmark) Equals(other *Benchmark) bool {
	if other == nil {
		return false
	}

	return !b.Diff(other).HasChanges()
}

// OnDecoded implements HclResource
func (b *Benchmark) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	b.setBaseProperties()
	return nil
}

func (b *Benchmark) String() string {
	// build list of children's names
	var children []string
	for _, child := range b.children {
		children = append(children, child.Name())
	}
	// build list of parents names
	var parents []string
	for _, p := range b.parents {
		parents = append(parents, p.Name())
	}
	sort.Strings(children)
	return fmt.Sprintf(`
	 -----
	 Name: %s
	 Title: %s
	 Description: %s
	 Parent: %s
	 Children:
	   %s
	`,
		b.FullName,
		types.SafeString(b.Title),
		types.SafeString(b.Description),
		strings.Join(parents, "\n    "),
		strings.Join(children, "\n    "))
}

// GetChildControls return a flat list of controls underneath the benchmark in the tree
func (b *Benchmark) GetChildControls() []*Control {
	var res []*Control
	for _, child := range b.children {
		if control, ok := child.(*Control); ok {
			res = append(res, control)
		} else if benchmark, ok := child.(*Benchmark); ok {
			res = append(res, benchmark.GetChildControls()...)
		}
	}
	return res
}

// GetWidth implements DashboardLeafNode
func (b *Benchmark) GetWidth() int {
	if b.Width == nil {
		return 0
	}
	return *b.Width
}

// GetDisplay implements DashboardLeafNode
func (b *Benchmark) GetDisplay() string {
	return typehelpers.SafeString(b.Display)
}

// GetType implements DashboardLeafNode
func (b *Benchmark) GetType() string {
	return typehelpers.SafeString(b.Type)
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (b *Benchmark) GetUnqualifiedName() string {
	return b.UnqualifiedName
}

func (b *Benchmark) Diff(other *Benchmark) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: b,
		Name: b.Name(),
	}

	if !utils.SafeStringsEqual(b.Description, other.Description) {
		res.AddPropertyDiff("Description")
	}
	if !utils.SafeStringsEqual(b.Documentation, other.Documentation) {
		res.AddPropertyDiff("Documentation")
	}
	if !utils.SafeStringsEqual(b.Title, other.Title) {
		res.AddPropertyDiff("Title")
	}
	if len(b.Tags) != len(other.Tags) {
		res.AddPropertyDiff("Tags")
	} else {
		for k, v := range b.Tags {
			if otherVal := other.Tags[k]; v != otherVal {
				res.AddPropertyDiff("Tags")
			}
		}
	}

	if !utils.SafeStringsEqual(b.Type, other.Type) {
		res.AddPropertyDiff("Type")
	}

	if len(b.ChildNameStrings) != len(other.ChildNameStrings) {
		res.AddPropertyDiff("Childen")
	} else {
		myChildNames := b.ChildNameStrings
		sort.Strings(myChildNames)
		otherChildNames := other.ChildNameStrings
		sort.Strings(otherChildNames)
		if strings.Join(myChildNames, ",") != strings.Join(otherChildNames, ",") {
			res.AddPropertyDiff("Childen")
		}
	}

	res.dashboardLeafNodeDiff(b, other)
	return res
}

func (b *Benchmark) WalkResources(resourceFunc func(resource ModTreeItem) (bool, error)) error {
	for _, child := range b.children {
		continueWalking, err := resourceFunc(child)
		if err != nil {
			return err
		}
		if !continueWalking {
			break
		}

		if childContainer, ok := child.(*Benchmark); ok {
			if err := childContainer.WalkResources(resourceFunc); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Benchmark) SetChildren(children []ModTreeItem) {
	b.children = children
}

// CtyValue implements CtyValueProvider
func (b *Benchmark) CtyValue() (cty.Value, error) {
	return GetCtyValue(b)
}

func (b *Benchmark) setBaseProperties() {
	if b.Base == nil {
		return
	}
	// copy base into the HclResourceImpl 'base' property so it is accessible to all nested structs
	b.base = b.Base
	// call into parent nested struct setBaseProperties
	b.ModTreeItemImpl.setBaseProperties()

	if b.Width == nil {
		b.Width = b.Base.Width
	}

	if b.Display == nil {
		b.Display = b.Base.Display
	}

	if len(b.children) == 0 {
		b.children = b.Base.children
		b.ChildNameStrings = b.Base.ChildNameStrings
		b.ChildNames = b.Base.ChildNames
	}
}
