package modconfig

import (
	"errors"
	"fmt"

	"github.com/turbot/go-kit/types"
)

type Control struct {
	ShortName *string

	Description   *string   `hcl:"description" column:"description" column_type:"text"`
	Documentation *string   `hcl:"documentation" column:"documentation" column_type:"text"`
	Labels        *[]string `hcl:"labels" column:"labels" column_type:"text[]"`
	Links         *[]string `hcl:"links" column:"links" column_type:"text[]"`
	ParentName    *string   `hcl:"parent" column:"parent" column_type:"text"`
	Query         *string   `hcl:"query" column:"query" column_type:"text"`
	Severity      *string   `hcl:"severity" column:"severity" column_type:"text"`
	Title         *string   `hcl:"title" column:"title" column_type:"text"`

	// populated when we build tree
	Parent ControlTreeItem

	// resource metadata
	Metadata *ResourceMetadata
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
	return fmt.Sprintf("%s.%s", c.Metadata.ModShortName, c.Name())
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

// GetMetadata :: implementation of ResourceWithMetadata
func (c *Control) GetMetadata() *ResourceMetadata {
	return c.Metadata
}
