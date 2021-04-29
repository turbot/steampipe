package modconfig

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// Control :: struct representing the control mod resource
type Control struct {
	ShortName string `json:"name"`

	Description   *string   `json:"description" column:"description" column_type:"text"`
	Documentation *string   `json:"documentation" column:"documentation" column_type:"text"`
	Labels        *[]string `json:"labels" column:"labels" column_type:"jsonb"`
	Links         *[]string `json:"links" column:"links" column_type:"jsonb"`
	ParentName    *string   `json:"parent" column:"parent" column_type:"text"`
	SQL           *string   `json:"sql" column:"sql" column_type:"text"`
	Severity      *string   `json:"severity" column:"severity" column_type:"text"`
	Title         *string   `json:"title" column:"title" column_type:"text"`

	DeclRange hcl.Range `json:"-"`

	parent   ControlTreeItem
	metadata *ResourceMetadata
}

// Schema :: hcl schema for control
func (c *Control) Schema() *hcl.BodySchema {
	// todo this could be done automatically if we had a tag for block properties
	var attributes []hcl.AttributeSchema
	for attribute := range HclProperties(c) {
		attributes = append(attributes, hcl.AttributeSchema{Name: attribute})
	}
	return &hcl.BodySchema{Attributes: attributes}
}

func (c *Control) CtyValue() (cty.Value, error) {
	return getCtyValue(c, controlBlock)
}

// controlBlock :: return the block schema of a hydrated Control
// used to convert a control into a cty type for block evaluation
// TODO autogenerate from Control struct by reflection?
var controlBlock = configschema.Block{
	Attributes: map[string]*configschema.Attribute{
		"name":          {Optional: true, Type: cty.String},
		"description":   {Optional: true, Type: cty.String},
		"documentation": {Optional: true, Type: cty.String},
		"labels":        {Optional: true, Type: cty.List(cty.String)},
		"links":         {Optional: true, Type: cty.List(cty.String)},
		"parent":        {Optional: true, Type: cty.String},
		"sql":           {Optional: true, Type: cty.String},
		"severity":      {Optional: true, Type: cty.String},
		"title":         {Optional: true, Type: cty.String},
	},
}

func controlCtyType() cty.Type {
	spec := controlBlock.DecoderSpec()
	return hcldec.ImpliedType(spec)
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
		c.ShortName,
		types.SafeString(c.Title),
		types.SafeString(c.Description),
		types.SafeString(c.SQL),
		types.SafeString(c.ParentName),
		labels, links)
}

//LongName :: name in format: '<modName>.control.<shortName>'
func (c *Control) LongName() string {
	return fmt.Sprintf("%s.%s", c.metadata.ModShortName, c.Name())
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
	c.parent = parent
	return nil
}

// Name :: implementation of ControlTreeItem, HclResource
// return name in format: 'control.<shortName>'
func (c *Control) Name() string {
	return fmt.Sprintf("control.%s", c.ShortName)
}

// Path :: implementation of ControlTreeItem
func (c *Control) Path() []string {
	path := []string{c.Name()}
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
