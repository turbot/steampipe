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
	ShortName string
	FullName  string `cty:"name"`

	Description   *string            `cty:"description" hcl:"description" column:"description,text"`
	Documentation *string            `cty:"documentation" hcl:"documentation" column:"documentation,text"`
	Links         *[]string          `cty:"links" hcl:"links" column:"links,jsonb"`
	Tags          *map[string]string `cty:"tags" hcl:"tags" column:"tags,jsonb"`
	SQL           *string            `cty:"sql" hcl:"sql" column:"sql,text"`
	Severity      *string            `cty:"severity" hcl:"severity" column:"severity,text"`
	Title         *string            `cty:"title" hcl:"title" column:"title,text"`

	DeclRange hcl.Range

	parent   ControlTreeItem
	metadata *ResourceMetadata
}

func NewControl(block *hcl.Block) *Control {
	control := &Control{
		ShortName: block.Labels[0],
		FullName:  fmt.Sprintf("control.%s", block.Labels[0]),
		DeclRange: block.DefRange,
	}
	return control
}

func (c *Control) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

func (c *Control) String() string {
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
  Links: %v
`,
		c.FullName,
		types.SafeString(c.Title),
		types.SafeString(c.Description),
		types.SafeString(c.SQL),
		c.parent.Name(),
		links)
}

// AddChild  :: implementation of ControlTreeItem - controls cannot have children so just return error
func (c *Control) AddChild(child ControlTreeItem) error {
	return errors.New("cannot add child to a control")
}

// SetParent :: implementation of ControlTreeItem
func (c *Control) SetParent(parent ControlTreeItem) error {
	c.parent = parent
	return nil
}

// Name :: implementation of ControlTreeItem, HclResource
// return name in format: 'control.<shortName>'
func (c *Control) Name() string {
	return c.FullName
}

// QualifiedName :: name in format: '<modName>.control.<shortName>'
func (c *Control) QualifiedName() string {
	return fmt.Sprintf("%s.%s", c.metadata.ModShortName, c.FullName)
}

// Path :: implementation of ControlTreeItem
func (c *Control) Path() []string {
	path := []string{c.FullName}
	if c.parent != nil {
		path = append(c.parent.Path(), path...)
	}
	return path
}

// GetMetadata :: implementation of HclResource
func (c *Control) GetMetadata() *ResourceMetadata {
	return c.metadata
}

// OnDecoded :: implementation of HclResource
func (c *Control) OnDecoded() {}

// SetMetadata :: implementation of HclResource
func (c *Control) SetMetadata(metadata *ResourceMetadata) {
	c.metadata = metadata
}
