package modconfig

import (
	"fmt"
	"sort"
	"strings"

	"github.com/turbot/go-kit/types"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type ControlGroupName struct {
	Name string `cty:"name"`
}

func (c ControlGroupName) String() string {
	return c.Name
}

// ControlGroup :: struct representing the control group mod resource
type ControlGroup struct {
	ShortName string
	FullName  string `cty:"name"`

	Description   *string            `cty:"description" hcl:"description" column_type:"text"`
	Documentation *string            `cty:"documentation" hcl:"documentation" column_type:"text"`
	Labels        *[]string          `cty:"labels" hcl:"labels" column_type:"jsonb"`
	Tags          *map[string]string `cty:"tags" hcl:"tags" column_type:"jsonb"`
	ParentName    *ControlGroupName  `cty:"parent" hcl:"parent" column_type:"text"`
	Title         *string            `cty:"title" hcl:"title" column_type:"text"`

	DeclRange hcl.Range

	parent   ControlTreeItem
	children []ControlTreeItem
	metadata *ResourceMetadata
}

func NewControlGroup(block *hcl.Block) *ControlGroup {
	return &ControlGroup{
		ShortName: block.Labels[0],
		FullName:  fmt.Sprintf("control_group.%s", block.Labels[0]),
		DeclRange: block.DefRange,
	}
}

func (c *ControlGroup) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

func (c *ControlGroup) String() string {
	var labels []string
	if c.Labels != nil {
		labels = *c.Labels
	}
	// build list of childrens long names
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
	 Labels: %v
	 Children:
	   %s
	`,
		c.FullName,
		types.SafeString(c.Title),
		types.SafeString(c.Description),
		c.GetParentName(),
		labels, strings.Join(children, "\n    "))
}

// GetChildControls :: return a flat list of controls underneath us in the tree
func (c *ControlGroup) GetChildControls() []*Control {
	var res []*Control
	for _, child := range c.children {
		if control, ok := child.(*Control); ok {
			res = append(res, control)
		} else if controlGroup, ok := child.(*ControlGroup); ok {
			res = append(res, controlGroup.GetChildControls()...)
		}
	}
	return res
}

// AddChild :: implementation of ControlTreeItem
func (c *ControlGroup) AddChild(child ControlTreeItem) error {
	// mod cannot be added as a child
	if _, ok := child.(*Mod); ok {
		return fmt.Errorf("mod cannot be added as a child")
	}

	c.children = append(c.children, child)
	return nil
}

// GetParentName :: implementation of ControlTreeItem
func (c *ControlGroup) GetParentName() string {
	if c.ParentName == nil {
		return ""
	}
	return c.ParentName.Name

}

// SetParent :: implementation of ControlTreeItem
func (c *ControlGroup) SetParent(parent ControlTreeItem) error {
	c.parent = parent
	return nil
}

// Name :: implementation of ControlTreeItem, HclResource
// return name in format: 'control.<shortName>'
func (c *ControlGroup) Name() string {
	return c.FullName
}

// QualifiedName :: name in format: '<modName>.control.<shortName>'
func (c *ControlGroup) QualifiedName() string {
	return fmt.Sprintf("%s.%s", c.metadata.ModShortName, c.FullName)
}

// Path :: implementation of ControlTreeItem
func (c *ControlGroup) Path() []string {
	path := []string{c.FullName}
	if c.parent != nil {
		path = append(c.parent.Path(), path...)
	}
	return path
}

// GetMetadata :: implementation of HclResource
func (c *ControlGroup) GetMetadata() *ResourceMetadata {
	return c.metadata
}

// SetMetadata :: implementation of HclResource
func (c *ControlGroup) SetMetadata(metadata *ResourceMetadata) {
	c.metadata = metadata
}
