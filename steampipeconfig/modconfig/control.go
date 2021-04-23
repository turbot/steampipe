package modconfig

import (
	"errors"
	"fmt"

	"github.com/turbot/go-kit/types"
)

type Control struct {
	ShortName   *string
	Title       *string   `hcl:"title" column:"title" column_type:"varchar(40)"`
	Description *string   `hcl:"description" column:"description" column_type:"text"`
	Query       *string   `hcl:"query" column:"query" column_type:"text"`
	Labels      *[]string `hcl:"labels" column:"labels" column_type:"varchar(40)[]"`
	Links       *[]string `hcl:"links" column:"links" column_type:"varchar(40)[]"`
	ParentName  *string   `hcl:"parent" column:"parent" column_type:"varchar(40)"`

	// populated when we build tree
	Parent ControlTreeItem

	// reflection data
	ReflectionData *CoreReflectionData
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

//LongName :: name in format: '<modName>.control.<shortName>'
func (c *Control) LongName() string {
	return fmt.Sprintf("%s.%s", types.SafeString(c.ReflectionData.Mod.ShortName), c.Name())
}

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

// GetCoreReflectionData :: implementation of ReflectionDataItem
func (c *Control) GetCoreReflectionData() *CoreReflectionData {
	return c.ReflectionData
}
