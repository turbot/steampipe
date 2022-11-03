package modconfig

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/types"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

// Benchmark is a struct representing the Benchmark resource
type Benchmark struct {
	ResourceWithMetadataBase

	ShortName       string `json:"-"`
	FullName        string `cty:"name" json:"-"`
	UnqualifiedName string `json:"-"`

	// child names as NamedItem structs - used to allow setting children via the 'children' property
	ChildNames NamedItemList `cty:"child_names" json:"-"`
	// used for introspection tables
	ChildNameStrings []string `cty:"child_name_strings" column:"children,jsonb" json:"-"`
	// the actual children
	Children      []ModTreeItem     `json:"-"`
	Description   *string           `cty:"description" hcl:"description" column:"description,text" json:"-"`
	Documentation *string           `cty:"documentation" hcl:"documentation" column:"documentation,text" json:"-"`
	Tags          map[string]string `cty:"tags" hcl:"tags,optional" column:"tags,jsonb" json:"-"`
	Title         *string           `cty:"title" hcl:"title" column:"title,text" json:"-"`

	// dashboard specific properties
	Base    *Benchmark `hcl:"base" json:"-"`
	Width   *int       `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type    *string    `cty:"type" hcl:"type" column:"type,text" json:"-"`
	Display *string    `cty:"display" hcl:"display" json:"-"`

	References []*ResourceReference `json:"-"`
	Mod        *Mod                 `cty:"mod" json:"-"`
	DeclRange  hcl.Range            `json:"-"`
	Paths      []NodePath           `column:"path,jsonb" json:"-"`
	Parents    []ModTreeItem        `json:"-"`
}

func NewBenchmark(block *hcl.Block, mod *Mod, shortName string) HclResource {
	benchmark := &Benchmark{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.benchmark.%s", mod.ShortName, shortName),
		UnqualifiedName: fmt.Sprintf("benchmark.%s", shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
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

// CtyValue implements HclResource
func (b *Benchmark) CtyValue() (cty.Value, error) {
	return getCtyValue(b)
}

// GetDeclRange implements HclResource
func (b *Benchmark) GetDeclRange() *hcl.Range {
	return &b.DeclRange
}

// BlockType implements HclResource
func (*Benchmark) BlockType() string {
	return BlockTypeBenchmark
}

// OnDecoded implements HclResource
func (b *Benchmark) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	b.setBaseProperties(resourceMapProvider)
	return nil
}

// AddReference implements ResourceWithMetadata
func (b *Benchmark) AddReference(ref *ResourceReference) {
	b.References = append(b.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (b *Benchmark) GetReferences() []*ResourceReference {
	return b.References
}

// GetMod implements ModTreeItem
func (b *Benchmark) GetMod() *Mod {
	return b.Mod
}

func (b *Benchmark) String() string {
	// build list of children's names
	var children []string
	for _, child := range b.Children {
		children = append(children, child.Name())
	}
	// build list of parents names
	var parents []string
	for _, p := range b.Parents {
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
	for _, child := range b.Children {
		if control, ok := child.(*Control); ok {
			res = append(res, control)
		} else if benchmark, ok := child.(*Benchmark); ok {
			res = append(res, benchmark.GetChildControls()...)
		}
	}
	return res
}

// AddParent implements ModTreeItem
func (b *Benchmark) AddParent(parent ModTreeItem) error {
	b.Parents = append(b.Parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (b *Benchmark) GetParents() []ModTreeItem {
	return b.Parents
}

// GetTitle implements HclResource
func (b *Benchmark) GetTitle() string {
	return typehelpers.SafeString(b.Title)
}

// GetDescription implements ModTreeItem, DashboardLeafNode
func (b *Benchmark) GetDescription() string {
	return typehelpers.SafeString(b.Description)
}

// GetTags implements HclResource, DashboardLeafNode
func (b *Benchmark) GetTags() map[string]string {
	if b.Tags != nil {
		return b.Tags
	}
	return map[string]string{}
}

// GetChildren implements ModTreeItem
func (b *Benchmark) GetChildren() []ModTreeItem {
	return b.Children
}

// GetPaths implements ModTreeItem
func (b *Benchmark) GetPaths() []NodePath {
	// lazy load
	if len(b.Paths) == 0 {
		b.SetPaths()
	}

	return b.Paths
}

// SetPaths implements ModTreeItem
func (b *Benchmark) SetPaths() {
	for _, parent := range b.Parents {
		for _, parentPath := range parent.GetPaths() {
			b.Paths = append(b.Paths, append(parentPath, b.Name()))
		}
	}
}

// Name implements ModTreeItem, HclResource, ResourceWithMetadata
// return name in format: '<modname>.control.<shortName>'
func (b *Benchmark) Name() string {
	return b.FullName
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

// GetDocumentation implements DashboardLeafNode, ModTreeItem
func (b *Benchmark) GetDocumentation() string {
	return typehelpers.SafeString(b.Documentation)
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
	for _, child := range b.Children {
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

func (b *Benchmark) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(b.Base, resourceMapProvider); !resolved {
		return
	} else {
		b.Base = base.(*Benchmark)
	}

	if b.Description == nil {
		b.Description = b.Base.Description
	}

	if b.Documentation == nil {
		b.Documentation = b.Base.Documentation
	}

	if b.Type == nil {
		b.Type = b.Base.Type
	}

	if b.Display == nil {
		b.Display = b.Base.Display
	}

	b.Tags = utils.MergeMaps(b.Tags, b.Base.Tags)
	if b.Title == nil {
		b.Title = b.Base.Title
	}

	if len(b.Children) == 0 {
		b.Children = b.Base.Children
		b.ChildNameStrings = b.Base.ChildNameStrings
		b.ChildNames = b.Base.ChildNames
	}
}
