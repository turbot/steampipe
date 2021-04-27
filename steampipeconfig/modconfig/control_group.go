package modconfig

import (
	"fmt"
	"sort"
	"strings"

	"github.com/turbot/go-kit/types"
)

type ControlGroup struct {
	ShortName *string

	Description   *string   `hcl:"description" column:"description" column_type:"text"`
	Documentation *string   `hcl:"documentation" column:"documentation" column_type:"text"`
	Labels        *[]string `hcl:"labels" column:"labels" column_type:"jsonb"`
	ParentName    *string   `hcl:"parent" column:"parent" column_type:"text"`
	Title         *string   `hcl:"title" column:"title" column_type:"text"`

	// populated when we build tree
	Parent   ControlTreeItem
	Children []ControlTreeItem

	// resource metadata
	Metadata *ResourceMetadata
}

func (c *ControlGroup) String() string {
	var labels []string
	if c.Labels != nil {
		labels = *c.Labels
	}
	// build list of childrens long names
	var children []string
	for _, child := range c.Children {
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
		types.SafeString(c.Name),
		types.SafeString(c.Title),
		types.SafeString(c.Description),
		types.SafeString(c.ParentName),
		labels, strings.Join(children, "\n    "))
}

// GetChildControls :: return a flat list of controls underneath us in the tree
func (c *ControlGroup) GetChildControls() []*Control {
	var res []*Control
	for _, child := range c.Children {
		if control, ok := child.(*Control); ok {
			res = append(res, control)
		} else if controlGroup, ok := child.(*ControlGroup); ok {
			res = append(res, controlGroup.GetChildControls()...)
		}
	}
	return res
}

// LongName :: name in format: '<modName>.control.<shortName>'
func (c *ControlGroup) LongName() string {
	return fmt.Sprintf("%s.%s", c.Metadata.ModShortName, c.Name())
}

// AddChild :: implementation of ControlTreeItem
func (c *ControlGroup) AddChild(child ControlTreeItem) error {
	// mod cannot be added as a child
	if _, ok := child.(*Mod); ok {
		return fmt.Errorf("mod cannot be added as a child")
	}

	c.Children = append(c.Children, child)
	return nil
}

// GetParentName :: implementation of ControlTreeItem
func (c *ControlGroup) GetParentName() string {
	return types.SafeString(c.ParentName)
}

// SetParent :: implementation of ControlTreeItem
func (c *ControlGroup) SetParent(parent ControlTreeItem) error {
	c.Parent = parent
	return nil
}

// Name :: implementation of ControlTreeItem
// return name in format: 'control.<shortName>'
func (c *ControlGroup) Name() string {
	return fmt.Sprintf("control_group.%s", types.SafeString(c.ShortName))
}

// Path :: implementation of ControlTreeItem
func (c *ControlGroup) Path() []string {
	path := []string{c.Name()}
	if c.Parent != nil {
		path = append(c.Parent.Path(), path...)
	}
	return path
}

// GetMetadata :: implementation of ResourceWithMetadata
func (c *ControlGroup) GetMetadata() *ResourceMetadata {
	return c.Metadata
}
