package modconfig

import (
	"fmt"
	"sort"
	"strings"

	"github.com/turbot/go-kit/types"
)

type ControlGroup struct {
	ShortName   *string
	Title       *string   `hcl:"title"`
	Description *string   `hcl:"description"`
	Labels      *[]string `hcl:"labels"`
	ParentName  *string   `hcl:"parent"`

	// populated when we build tree
	Parent   ControlTreeItem
	Children []ControlTreeItem

	// reflection data
	ReflectionData *ReflectionData
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

//func (c *ControlGroup) Equals(other *Control) bool {
//	return types.SafeString(c.Name) == types.SafeString(other.Name) &&
//		types.SafeString(c.Title) == types.SafeString(other.Title) &&
//		types.SafeString(c.Description) == types.SafeString(other.Description) &&
//		reflect.DeepEqual(c.Labels, other.Labels) &&
//		c.Parent == other.Parent z
//
//
//}

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
	return fmt.Sprintf("control.%s", types.SafeString(c.ShortName))
}

// Path :: implementation of ControlTreeItem
func (c *ControlGroup) Path() []string {
	path := []string{c.Name()}
	if c.Parent != nil {
		path = append(c.Parent.Path(), path...)
	}
	return path
}
