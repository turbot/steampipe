package modconfig

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/types"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

type NamedItem struct {
	Name string `cty:"name"`
}

func (c NamedItem) String() string {
	return c.Name
}

// Benchmark is a struct representing the Benchmark resource
type Benchmark struct {
	ShortName string
	FullName  string `cty:"name"`

	ChildNames    *[]NamedItem       `cty:"children" hcl:"children"`
	Description   *string            `cty:"description" hcl:"description" column:"description,text"`
	Documentation *string            `cty:"documentation" hcl:"documentation" column:"documentation,text"`
	Tags          *map[string]string `cty:"tags" hcl:"tags" column:"tags,jsonb"`
	Title         *string            `cty:"title" hcl:"title" column:"title,text"`

	// list of all block referenced by the resource
	References []string `column:"refs,jsonb"`

	ChildNameStrings []string `column:"children,jsonb"`
	DeclRange        hcl.Range

	parents  []ControlTreeItem
	children []ControlTreeItem
	metadata *ResourceMetadata
}

func NewBenchmark(block *hcl.Block) *Benchmark {
	return &Benchmark{
		ShortName: block.Labels[0],
		FullName:  fmt.Sprintf("benchmark.%s", block.Labels[0]),
		DeclRange: block.DefRange,
	}
}

// CtyValue implements HclResource
func (b *Benchmark) CtyValue() (cty.Value, error) {
	return getCtyValue(b)
}

// OnDecoded implements HclResource
func (b *Benchmark) OnDecoded(*hcl.Block) {
	if b.ChildNames == nil || len(*b.ChildNames) == 0 {
		return
	}

	b.ChildNameStrings = make([]string, len(*b.ChildNames))
	for i, n := range *b.ChildNames {
		b.ChildNameStrings[i] = n.Name
	}
	b.children = make([]ControlTreeItem, len(b.ChildNameStrings))
}

// AddReference implements HclResource
func (b *Benchmark) AddReference(reference string) {
	b.References = append(b.References, reference)
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

// AddChild implements ControlTreeItem
func (b *Benchmark) AddChild(child ControlTreeItem) error {
	// mod cannot be added as a child
	if _, ok := child.(*Mod); ok {
		return fmt.Errorf("mod cannot be added as a child")
	}

	// now find which positing this child is in the array
	for i, name := range b.ChildNameStrings {
		if name == child.Name() {
			b.children[i] = child
			return nil
		}
	}

	return fmt.Errorf("benchmark '%s' has no child '%s'", b.Name(), child.Name())
}

// AddParent implements ControlTreeItem
func (b *Benchmark) AddParent(parent ControlTreeItem) error {
	b.parents = append(b.parents, parent)
	return nil
}

// GetParents implements ControlTreeItem
func (b *Benchmark) GetParents() []ControlTreeItem {
	return b.parents
}

// GetTitle implements ControlTreeItem
func (b *Benchmark) GetTitle() string {
	return typehelpers.SafeString(b.Title)
}

// GetDescription implements ControlTreeItem
func (b *Benchmark) GetDescription() string {
	return typehelpers.SafeString(b.Description)
}

// GetTags implements ControlTreeItem
func (b *Benchmark) GetTags() map[string]string {
	if b.Tags != nil {
		return *b.Tags
	}
	return map[string]string{}
}

// GetChildren implements ControlTreeItem
func (b *Benchmark) GetChildren() []ControlTreeItem {
	return b.children
}

// Path implements ControlTreeItem
// TODO is this needed - if so it needs to be an array
func (b *Benchmark) Path() []string {
	path := []string{b.FullName}
	if b.parents != nil {
		path = append(b.parents[0].Path(), path...)
	}
	return path
}

// GetMetadata implements ResourceWithMetadata
func (b *Benchmark) GetMetadata() *ResourceMetadata {
	return b.metadata
}

// SetMetadata implements ResourceWithMetadata
func (b *Benchmark) SetMetadata(metadata *ResourceMetadata) {
	b.metadata = metadata
}

// Name implements ControlTreeItem, HclResource, ResourceWithMetadata
// return name in format: 'control.<shortName>'
func (b *Benchmark) Name() string {
	return b.FullName
}

// QualifiedName returns the name in format: '<modName>.control.<shortName>'
func (b *Benchmark) QualifiedName() string {
	return fmt.Sprintf("%s.%s", b.metadata.ModShortName, b.FullName)
}
