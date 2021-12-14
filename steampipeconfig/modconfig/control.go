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
	// args
	// arguments may be specified by either a map of named args or as a list of positional args
	// we apply special decode logic to convert the params block into a QueryArgs object
	// with either an args map or list assigned
	Args   *QueryArgs  `cty:"args" column:"args,jsonb"`
	Params []*ParamDef `cty:"params" column:"params,jsonb"`

	// list of all blocks referenced by the resource
	References []*ResourceReference
	Mod        *Mod `cty:"mod"`
	DeclRange  hcl.Range

	parents               []ModTreeItem
	metadata              *ResourceMetadata
	PreparedStatementName string `column:"prepared_statement_name,text"`
}

func NewControl(block *hcl.Block) *Control {
	control := &Control{
		ShortName: block.Labels[0],
		FullName:  fmt.Sprintf("control.%s", block.Labels[0]),
		DeclRange: block.DefRange,
		Args:      NewQueryArgs(),
	}
	return control
}

func (c *Control) Equals(other *Control) bool {
	res := c.ShortName == other.ShortName &&
		c.FullName == other.FullName &&
		typehelpers.SafeString(c.Description) == typehelpers.SafeString(other.Description) &&
		typehelpers.SafeString(c.Documentation) == typehelpers.SafeString(other.Documentation) &&
		typehelpers.SafeString(c.SearchPath) == typehelpers.SafeString(other.SearchPath) &&
		typehelpers.SafeString(c.SearchPathPrefix) == typehelpers.SafeString(other.SearchPathPrefix) &&
		typehelpers.SafeString(c.Severity) == typehelpers.SafeString(other.Severity) &&
		typehelpers.SafeString(c.SQL) == typehelpers.SafeString(other.SQL) &&
		typehelpers.SafeString(c.Title) == typehelpers.SafeString(other.Title)
	if !res {
		return res
	}
	// tags
	if c.Tags == nil {
		if other.Tags != nil {
			return false
		}
	} else {
		// we have tags
		if other.Tags == nil {
			return false
		}
		for k, v := range *c.Tags {
			if otherVal, ok := (*other.Tags)[k]; !ok && v != otherVal {
				return false
			}
		}
	}

	// args
	if c.Args == nil {
		if other.Args != nil {
			return false
		}
	} else {
		// we have args
		if other.Args == nil {
			return false
		}
		if !c.Args.Equals(other.Args) {
			return false
		}
	}

	// query
	if c.Query == nil {
		if other.Query != nil {
			return false
		}
	} else {
		// we have a query
		if other.Query == nil {
			return false
		}
		if !c.Query.Equals(other.Query) {
			return false
		}
	}

	// params
	if len(c.Params) != len(other.Params) {
		return false
	}
	for i, p := range c.Params {
		if !p.Equals(other.Params[i]) {
			return false
		}
	}

	return true
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

	// add param defs if there are any
	if len(c.Params) > 0 {
		var paramDefsStr = make([]string, len(c.Params))
		for i, def := range c.Params {
			paramDefsStr[i] = def.String()
		}
		res += fmt.Sprintf("Params:\n\t%s\n  ", strings.Join(paramDefsStr, "\n\t"))
	}

	// add args
	if c.Args != nil && !c.Args.Empty() {
		res += fmt.Sprintf("Args:\n\t%s\n  ", c.Args)
	}
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
	return fmt.Sprintf("%s.%s", c.metadata.ModName, c.FullName)
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

// AddReference implements HclResource
func (c *Control) AddReference(ref *ResourceReference) {
	c.References = append(c.References, ref)
}

// SetMod implements HclResource
func (c *Control) SetMod(mod *Mod) {
	c.Mod = mod
}

// GetMod implements HclResource
func (c *Control) GetMod() *Mod {
	return c.Mod
}

// GetDeclRange implements HclResource
func (c *Control) GetDeclRange() *hcl.Range {
	return &c.DeclRange
}

// GetMetadata implements ResourceWithMetadata
func (c *Control) GetMetadata() *ResourceMetadata {
	return c.metadata
}

// SetMetadata implements ResourceWithMetadata
func (c *Control) SetMetadata(metadata *ResourceMetadata) {
	c.metadata = metadata
}

// GetParams implements QueryProvider
func (c *Control) GetParams() []*ParamDef {
	return c.Params
}

// GetPreparedStatementName implements QueryProvider
func (c *Control) GetPreparedStatementName() string {
	// lazy load
	if c.PreparedStatementName == "" {
		c.PreparedStatementName = preparedStatementName(c)
	}
	return c.PreparedStatementName
}

// ModName implements QueryProvider
func (c *Control) ModName() string {
	return c.Mod.ShortName
}
