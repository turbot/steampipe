package modconfig

import (
	"fmt"
	"sort"
	"strings"

	"github.com/turbot/go-kit/types"

	"github.com/hashicorp/hcl/v2"
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

	ChildNameStrings []string `column:"children,jsonb"`
	DeclRange        hcl.Range

	parent   ControlTreeItem
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
func (c *Benchmark) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

// OnDecoded implements HclResource
func (c *Benchmark) OnDecoded() {
	if c.ChildNames == nil || len(*c.ChildNames) == 0 {
		return
	}

	c.ChildNameStrings = make([]string, len(*c.ChildNames))
	for i, n := range *c.ChildNames {
		c.ChildNameStrings[i] = n.Name
	}
}

func (c *Benchmark) String() string {
	// build list of children's long names
	var children []string
	for _, child := range c.children {
		children = append(children, child.Name())
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
		c.FullName,
		types.SafeString(c.Title),
		types.SafeString(c.Description),
		c.parent.Name(),
		strings.Join(children, "\n    "))
}

// GetChildControls return a flat list of controls underneath us in the tree
func (c *Benchmark) GetChildControls() []*Control {
	var res []*Control
	for _, child := range c.children {
		if control, ok := child.(*Control); ok {
			res = append(res, control)
		} else if benchmark, ok := child.(*Benchmark); ok {
			res = append(res, benchmark.GetChildControls()...)
		}
	}
	return res
}

// AddChild implements ControlTreeItem
func (c *Benchmark) AddChild(child ControlTreeItem) error {
	// mod cannot be added as a child
	if _, ok := child.(*Mod); ok {
		return fmt.Errorf("mod cannot be added as a child")
	}

	c.children = append(c.children, child)
	return nil
}

// SetParent implements ControlTreeItem
func (c *Benchmark) SetParent(parent ControlTreeItem) error {
	c.parent = parent
	return nil
}

// GetMetadata implements ResourceWithMetadata
func (c *Benchmark) GetMetadata() *ResourceMetadata {
	return c.metadata
}

// SetMetadata implements ResourceWithMetadata
func (c *Benchmark) SetMetadata(metadata *ResourceMetadata) {
	c.metadata = metadata
}

// Name implements ControlTreeItem, HclResource, ResourceWithMetadata
// return name in format: 'control.<shortName>'
func (c *Benchmark) Name() string {
	return c.FullName
}

// QualifiedName returns the name in format: '<modName>.control.<shortName>'
func (c *Benchmark) QualifiedName() string {
	return fmt.Sprintf("%s.%s", c.metadata.ModShortName, c.FullName)
}

// Path implements ControlTreeItem
func (c *Benchmark) Path() []string {
	path := []string{c.FullName}
	if c.parent != nil {
		path = append(c.parent.Path(), path...)
	}
	return path
}
