package modconfig

import (
	"errors"
	"fmt"

	"github.com/turbot/go-kit/types"
)

type Control struct {
	ShortName   *string
	Title       *string   `hcl:"title"`
	Description *string   `hcl:"description"`
	Query       *string   `hcl:"query"`
	Labels      *[]string `hcl:"labels"`
	Links       *[]string `hcl:"links"`
	ParentName  *string   `hcl:"parent"`

	// populated when we build tree
	Mod    *Mod
	Parent ControlTreeItem

	// reflection data
}

func (c *Control) String() string {
	var labels []string
	if c.Labels != nil {
		labels = *c.Labels
	}
	var links []string
	if c.Links != nil {
		links = *c.Links
	}
	return fmt.Sprintf(`
  -----
  Name: %s
  Title: %s
  Description: %s
  Query: %s
  Parent: %s
  Labels: %v
  Links: %v
`,
		types.SafeString(c.ShortName),
		types.SafeString(c.Title),
		types.SafeString(c.Description),
		types.SafeString(c.Query),
		types.SafeString(c.ParentName),
		labels, links)
}

//func (c *Control) Equals(other *Control) bool {
//	return types.SafeString(c.Name) == types.SafeString(other.Name) &&
//		types.SafeString(c.Title) == types.SafeString(other.Title) &&
//		types.SafeString(c.Description) == types.SafeString(other.Description) &&
//		types.SafeString(c.SQL) == types.SafeString(other.SQL) &&
//		types.SafeString(c.Links) == types.SafeString(other.Links) &&
//		reflect.DeepEqual(c.Tags, other.Tags)
//}

// AddChild  :: implementation of ControlTreeItem - controls cannot have children so just return error
func (c *Control) AddChild(child ControlTreeItem) error {
	return errors.New("cannot add child to a control")
}

// GetParentName :: implementation of ControlTreeItem
func (c *Control) GetParentName() string {
	return types.SafeString(c.ParentName)
}

// SetParent :: implementation of ControlTreeItem
func (c *Control) SetParent(parent ControlTreeItem) error {
	c.Parent = parent
	return nil
}

// Name :: implementation of ControlTreeItem
// return name in format: 'control.<shortName>'
func (c *Control) Name() string {
	return fmt.Sprintf("control.%s", types.SafeString(c.ShortName))
}

// Path :: implementation of ControlTreeItem
func (c *Control) Path() []string {
	path := []string{c.Name()}
	if c.Parent != nil {
		path = append(c.Parent.Path(), path...)
	}
	return path
}
