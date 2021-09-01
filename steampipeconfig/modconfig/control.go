package modconfig

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/types"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// Control is a struct representing the Control resource
type Control struct {
	ShortName        string
	FullName         string             `cty:"name"`
	Description      *string            `cty:"description" column:"description,text"`
	Documentation    *string            `cty:"documentation"  column:"documentation,text"`
	SearchPath       *string            `cty:"search_path"  column:"search_path,text"`
	SearchPathPrefix *string            `cty:"search_path_prefix"  column:"search_path_prefix,text"`
	Severity         *string            `cty:"severity"  column:"severity,text"`
	SQL              *string            `cty:"sql"  column:"sql,text"`
	Tags             *map[string]string `cty:"tags"  column:"tags,jsonb"`
	Title            *string            `cty:"title"  column:"title,text"`
	Query            *Query
	// parameters
	// params may be specified by either a map of named parameters or as a list of positional parameters
	// we apply special decode logic to convert the params block into a QueryParams object
	// with either a param map or list assigned
	// TODO CTY and REFLECTION TABLES?
	Params *QueryParams

	// list of all block referenced by the resource
	References []string `column:"refs,jsonb"`

	DeclRange hcl.Range

	parents  []ModTreeItem
	metadata *ResourceMetadata
}

func NewControl(block *hcl.Block) *Control {
	control := &Control{
		ShortName: block.Labels[0],
		FullName:  fmt.Sprintf("control.%s", block.Labels[0]),
		DeclRange: block.DefRange,
		Params:    NewQueryParams(),
	}
	return control
}

func (c *Control) String() string {
	// build list of parents's names
	parents := c.GetParentNames()
	res := fmt.Sprintf(`
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

	return res
}

func (c *Control) GetParentNames() []string {
	var parents []string
	for _, p := range c.parents {
		parents = append(parents, p.Name())
	}
	return parents
}

// AddChild implements ModTreeItem - controls cannot have children so just return error
func (c *Control) AddChild(child ModTreeItem) error {
	return errors.New("cannot add child to a control")
}

// AddParent implements ModTreeItem
func (c *Control) AddParent(parent ModTreeItem) error {
	c.parents = append(c.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (c *Control) GetParents() []ModTreeItem {
	return c.parents
}

// GetTitle implements ModTreeItem
func (c *Control) GetTitle() string {
	return typehelpers.SafeString(c.Title)
}

// GetDescription implements ModTreeItem
func (c *Control) GetDescription() string {
	return typehelpers.SafeString(c.Description)
}

// GetTags implements ModTreeItem
func (c *Control) GetTags() map[string]string {
	if c.Tags != nil {
		return *c.Tags
	}
	return map[string]string{}
}

// GetChildren implements ModTreeItem
func (c *Control) GetChildren() []ModTreeItem {
	return []ModTreeItem{}
}

// Name implements ModTreeItem, HclResource
// return name in format: 'control.<shortName>'
func (c *Control) Name() string {
	return c.FullName
}

// QualifiedName returns the name in format: '<modName>.control.<shortName>'
func (c *Control) QualifiedName() string {
	return fmt.Sprintf("%s.%s", c.metadata.ModShortName, c.FullName)
}

// GetPaths implements ModTreeItem
func (c *Control) GetPaths() []NodePath {
	var res []NodePath
	for _, parent := range c.parents {
		for _, parentPath := range parent.GetPaths() {
			res = append(res, append(parentPath, c.Name()))
		}
	}
	return res
}

// CtyValue implements HclResource
func (c *Control) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

// OnDecoded implements HclResource
func (c *Control) OnDecoded(*hcl.Block) hcl.Diagnostics { return nil }

// OnDecoded implements HclResource
func (c *Control) SetSQL(sql string, block *hcl.Block) hcl.Diagnostics {
	// if both query and SQL are set, raise an error
	if typehelpers.SafeString(c.SQL) != "" {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s has both 'SQL' and 'query' property set - only 1 of these may be set", c.FullName),
			Subject:  &block.DefRange,
		}}
	}
	return nil
}

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
