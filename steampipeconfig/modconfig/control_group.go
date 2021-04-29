package modconfig

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// ControlGroup :: struct representing the control group mod resource
type ControlGroup struct {
	ShortName string `json:"name"`

	Description   *string   `json:"description" column:"description" column_type:"text"`
	Documentation *string   `json:"documentation" column:"documentation" column_type:"text"`
	Labels        *[]string `json:"labels" column:"labels" column_type:"jsonb"`
	ParentName    *string   `json:"parent" column:"parent" column_type:"text"`
	Title         *string   `json:"title" column:"title" column_type:"text"`

	DeclRange hcl.Range `json:"-"`

	parent   ControlTreeItem
	children []ControlTreeItem
	metadata *ResourceMetadata
}

// Schema :: hcl schema for control
func (c *ControlGroup) Schema() *hcl.BodySchema {
	var attributes []hcl.AttributeSchema
	for attribute := range GetAttributeDetails(c) {
		attributes = append(attributes, hcl.AttributeSchema{Name: attribute})
	}
	return &hcl.BodySchema{Attributes: attributes}
}

func (q *ControlGroup) CtyValue() (cty.Value, error) {
	return getCtyValue(q)
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
		types.SafeString(c.Name),
		types.SafeString(c.Title),
		types.SafeString(c.Description),
		types.SafeString(c.ParentName),
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

// LongName :: name in format: '<modName>.control.<shortName>'
func (c *ControlGroup) LongName() string {
	return fmt.Sprintf("%s.%s", c.metadata.ModShortName, c.Name())
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
	return types.SafeString(c.ParentName)
}

// SetParent :: implementation of ControlTreeItem
func (c *ControlGroup) SetParent(parent ControlTreeItem) error {
	c.parent = parent
	return nil
}

// Name :: implementation of ControlTreeItem, HclResource
// return name in format: 'control.<shortName>'
func (c *ControlGroup) Name() string {
	return fmt.Sprintf("control_group.%s", c.ShortName)
}

// Path :: implementation of ControlTreeItem
func (c *ControlGroup) Path() []string {
	path := []string{c.Name()}
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
