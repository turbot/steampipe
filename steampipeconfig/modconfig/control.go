package modconfig

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/types"
	typehelpers "github.com/turbot/go-kit/types"
)

// Control is a struct representing the Control resource
type Control struct {
	ShortName        string             `cty:"name"`
	FullName         string             `cty:"name"`
	Description      *string            `cty:"description" hcl:"description" column:"description,text"`
	Documentation    *string            `cty:"documentation" hcl:"documentation" column:"documentation,text"`
	SearchPath       *string            `cty:"search_path" hcl:"search_path" column:"search_path,text"`
	SearchPathPrefix *string            `cty:"search_path_prefix" hcl:"search_path_prefix" column:"search_path_prefix,text"`
	Severity         *string            `cty:"severity" hcl:"severity" column:"severity,text"`
	SQL              *string            `cty:"sql" hcl:"sql" column:"sql,text"`
	Tags             *map[string]string `cty:"tags" hcl:"tags" column:"tags,jsonb"`
	Title            *string            `cty:"title" hcl:"title" column:"title,text""`

	// list of all block referenced by the resource
	References []string `column:"refs,jsonb"`

	DeclRange hcl.Range

	parents  []ControlTreeItem
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

func (c *Control) String() string {
	// build list of parents's names
	parents := c.GetParentNames()
	return fmt.Sprintf(`
  -----
  Name: %s
  Title: %s
  Description: %s
  SQL: %s
  Parents: %s
`,
		c.FullName,
		types.SafeString(c.Title),
		types.SafeString(c.Description),
		types.SafeString(c.SQL),
		strings.Join(parents, "\n    "))
}

func (c *Control) GetParentNames() []string {
	var parents []string
	for _, p := range c.parents {
		parents = append(parents, p.Name())
	}
	return parents
}

// AddChild implements ControlTreeItem - controls cannot have children so just return error
func (c *Control) AddChild(child ControlTreeItem) error {
	return errors.New("cannot add child to a control")
}

// AddParent implements ControlTreeItem
func (c *Control) AddParent(parent ControlTreeItem) error {
	c.parents = append(c.parents, parent)
	return nil
}

// GetParents implements ControlTreeItem
func (c *Control) GetParents() []ControlTreeItem {
	return c.parents
}

// GetTitle implements ControlTreeItem
func (c *Control) GetTitle() string {
	return typehelpers.SafeString(c.Title)
}

// GetDescription implements ControlTreeItem
func (c *Control) GetDescription() string {
	return typehelpers.SafeString(c.Description)
}

// GetTags implements ControlTreeItem
func (c *Control) GetTags() map[string]string {
	if c.Tags != nil {
		return *c.Tags
	}
	return map[string]string{}
}

// GetChildren implements ControlTreeItem
func (c *Control) GetChildren() []ControlTreeItem {
	return []ControlTreeItem{}
}

// Name implements ControlTreeItem, HclResource
// return name in format: 'control.<shortName>'
func (c *Control) Name() string {
	return c.FullName
}

// QualifiedName returns the name in format: '<modName>.control.<shortName>'
func (c *Control) QualifiedName() string {
	return fmt.Sprintf("%s.%s", c.metadata.ModShortName, c.FullName)
}

// Path implements ControlTreeItem
func (c *Control) Path() []string {
	// TODO update for multiple paths
	path := []string{c.FullName}
	if c.parents != nil {
		path = append(c.parents[0].Path(), path...)
	}
	return path
}

// CtyValue implements HclResource
func (c *Control) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

// OnDecoded implements HclResource
func (c *Control) OnDecoded(*hcl.Block) {}

// AddReference implements HclResource
func (c *Control) AddReference(reference string) {
	c.References = append(c.References, reference)
}

// GetMetadata implements ResourceWithMetadata
func (c *Control) GetMetadata() *ResourceMetadata {
	return c.metadata
}

// SetMetadata implements ResourceWithMetadata
func (c *Control) SetMetadata(metadata *ResourceMetadata) {
	c.metadata = metadata
}
