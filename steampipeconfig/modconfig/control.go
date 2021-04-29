package modconfig

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// Control :: struct representing the control mod resource
type Control struct {
	Name string `cty:"name"`

	Description   *string   `cty:"description" column:"description" column_type:"text"`
	Documentation *string   `cty:"documentation" column:"documentation" column_type:"text"`
	Labels        *[]string `cty:"labels" column:"labels" column_type:"jsonb"`
	Links         *[]string `cty:"links" column:"links" column_type:"jsonb"`
	ParentName    *string   `cty:"parent" column:"parent" column_type:"text"`
	SQL           *string   `cty:"sql" column:"sql" column_type:"text"`
	Severity      *string   `cty:"severity" column:"severity" column_type:"text"`
	Title         *string   `cty:"title" column:"title" column_type:"text"`

	DeclRange hcl.Range

	parent   ControlTreeItem
	metadata *ResourceMetadata
}

// Schema :: hcl schema for control
func (c *Control) Schema() *hcl.BodySchema {
	// todo this could be done automatically if we had a tag for block properties
	var attributes []hcl.AttributeSchema
	for attribute := range GetAttributeDetails(c) {
		attributes = append(attributes, hcl.AttributeSchema{Name: attribute})
	}
	return &hcl.BodySchema{Attributes: attributes}
}

func (c *Control) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
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
  SQL: %s
  Parent: %s
  Labels: %v
  Links: %v
`,
		c.Name,
		types.SafeString(c.Title),
		types.SafeString(c.Description),
		types.SafeString(c.SQL),
		types.SafeString(c.ParentName),
		labels, links)
}

// AddChild  :: implementation of ControlTreeItem - controls cannot have children so just return error
func (c *Control) AddChild(child ControlTreeItem) error {
	return errors.New("cannot add child to a control")
}

// GetParentName :: implementation of ControlTreeItem
func (c *Control) GetParentName() string {
	return getParentName(types.SafeString(c.ParentName))
}

// SetParent :: implementation of ControlTreeItem
func (c *Control) SetParent(parent ControlTreeItem) error {
	c.parent = parent
	return nil
}

// FullName :: implementation of ControlTreeItem, HclResource
// return name in format: 'control.<shortName>'
func (c *Control) FullName() string {
	return fmt.Sprintf("control.%s", c.Name)
}

// QualifiedName :: name in format: '<modName>.control.<shortName>'
func (c *Control) QualifiedName() string {
	return fmt.Sprintf("%s.%s", c.metadata.ModShortName, c.FullName())
}

// Path :: implementation of ControlTreeItem
func (c *Control) Path() []string {
	path := []string{c.FullName()}
	if c.parent != nil {
		path = append(c.parent.Path(), path...)
	}
	return path
}

// GetMetadata :: implementation of HclResource
func (c *Control) GetMetadata() *ResourceMetadata {
	return c.metadata
}

// SetMetadata :: implementation of HclResource
func (c *Control) SetMetadata(metadata *ResourceMetadata) {
	c.metadata = metadata
}
