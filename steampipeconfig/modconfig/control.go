package modconfig

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/utils"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/types"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// Control is a struct representing the Control resource
type Control struct {
	ShortName        string            `json:"-"`
	FullName         string            `cty:"name" json:"-"`
	Description      *string           `cty:"description" column:"description,text" json:"-"`
	Documentation    *string           `cty:"documentation"  column:"documentation,text" json:"-"`
	SearchPath       *string           `cty:"search_path"  column:"search_path,text" json:"-"`
	SearchPathPrefix *string           `cty:"search_path_prefix"  column:"search_path_prefix,text" json:"-"`
	Severity         *string           `cty:"severity"  column:"severity,text" json:"-"`
	SQL              *string           `cty:"sql"  column:"sql,text" json:"-"`
	Tags             map[string]string `cty:"tags"  column:"tags,jsonb" json:"-"`
	Title            *string           `cty:"title"  column:"title,text" json:"-"`
	Query            *Query
	// args
	// arguments may be specified by either a map of named args or as a list of positional args
	// we apply special decode logic to convert the params block into a QueryArgs object
	// with either an args map or list assigned
	Args                  *QueryArgs           `cty:"args" column:"args,jsonb" json:"-"`
	Params                []*ParamDef          `cty:"params" column:"params,jsonb" json:"-"`
	References            []*ResourceReference `json:"-"`
	Mod                   *Mod                 `cty:"mod" json:"-"`
	DeclRange             hcl.Range            `json:"-"`
	PreparedStatementName string               `column:"prepared_statement_name,text" json:"-"`
	UnqualifiedName       string               `json:"-"`

	// report specific properties
	Base  *Control   `hcl:"base" json:"-"`
	Width *int       `cty:"width" hcl:"width" column:"width,text"  json:"-"`
	Paths []NodePath `json:"-"`

	parents  []ModTreeItem
	metadata *ResourceMetadata
}

func NewControl(block *hcl.Block) *Control {
	control := &Control{
		ShortName:       block.Labels[0],
		FullName:        fmt.Sprintf("control.%s", block.Labels[0]),
		UnqualifiedName: fmt.Sprintf("control.%s", block.Labels[0]),
		DeclRange:       block.DefRange,
		Args:            NewQueryArgs(),
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
	if len(c.Tags) != len(other.Tags) {
		return false
	}
	for k, v := range c.Tags {
		if otherVal := other.Tags[k]; v != otherVal {
			return false
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
		return c.Tags
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

// QualifiedNameWithVersion returns the name in format: '<modName>@version.control.<shortName>'
func (c *Control) QualifiedNameWithVersion() string {
	return fmt.Sprintf("%s.%s", c.Mod.NameWithVersion(), c.FullName)
}

// GetPaths implements ModTreeItem
func (c *Control) GetPaths() []NodePath {
	// lazy load
	if len(c.Paths) == 0 {
		c.SetPaths()
	}

	return c.Paths
}

// SetPaths implements ModTreeItem
func (c *Control) SetPaths() {
	for _, parent := range c.parents {
		for _, parentPath := range parent.GetPaths() {
			c.Paths = append(c.Paths, append(parentPath, c.Name()))
		}
	}
}

// CtyValue implements HclResource
func (c *Control) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

// OnDecoded implements HclResource
func (c *Control) OnDecoded(*hcl.Block) hcl.Diagnostics {
	c.setBaseProperties()
	// verify the control has either a query or a sql attribute
	if c.Query == nil && c.SQL == nil {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s must define either a 'sql' property or a 'query' property", c.FullName),
			Subject:  &c.DeclRange,
		}}
	}

	return nil
}

// AddReference implements HclResource
func (c *Control) AddReference(ref *ResourceReference) {
	c.References = append(c.References, ref)
}

// SetMod implements HclResource
func (c *Control) SetMod(mod *Mod) {
	c.Mod = mod
	// add mod name to full name
	c.FullName = fmt.Sprintf("%s.%s", mod.ShortName, c.FullName)
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
	return c.Mod.NameWithVersion()
}

// GetSQL implements ReportLeafNode
func (c *Control) GetSQL() string {
	return typehelpers.SafeString(c.SQL)
}

// GetWidth implements ReportLeafNode
func (c *Control) GetWidth() int {
	if c.Width == nil {
		return 0
	}
	return *c.Width
}

// GetUnqualifiedName implements ReportLeafNode
func (c *Control) GetUnqualifiedName() string {
	return c.UnqualifiedName
}

func (c *Control) Diff(other *Control) *ReportTreeItemDiffs {
	res := &ReportTreeItemDiffs{
		Item: c,
		Name: c.Name(),
	}

	if !utils.SafeStringsEqual(c.FullName, other.FullName) {
		res.AddPropertyDiff("Name")
	}
	if !utils.SafeStringsEqual(c.Description, other.Description) {
		res.AddPropertyDiff("Description")
	}
	if !utils.SafeStringsEqual(c.Documentation, other.Documentation) {
		res.AddPropertyDiff("Documentation")
	}
	if !utils.SafeStringsEqual(c.SearchPath, other.SearchPath) {
		res.AddPropertyDiff("SearchPath")
	}
	if !utils.SafeStringsEqual(c.SearchPathPrefix, other.SearchPathPrefix) {
		res.AddPropertyDiff("SearchPathPrefix")
	}
	if !utils.SafeStringsEqual(c.Severity, other.Severity) {
		res.AddPropertyDiff("Severity")
	}
	if !utils.SafeStringsEqual(c.Title, other.Title) {
		res.AddPropertyDiff("Title")
	}
	if !utils.SafeStringsEqual(c.SQL, other.SQL) {
		res.AddPropertyDiff("SQL")
	}
	if !utils.SafeStringsEqual(c.Title, other.Title) {
		res.AddPropertyDiff("Title")
	}
	if !utils.SafeIntEqual(c.Width, other.Width) {
		res.AddPropertyDiff("Width")
	}
	if len(c.Tags) != len(other.Tags) {
		res.AddPropertyDiff("Tags")
	} else {
		for k, v := range c.Tags {
			if otherVal := other.Tags[k]; v != otherVal {
				res.AddPropertyDiff("Tags")
			}
		}
	}

	// args
	if c.Args == nil {
		if other.Args != nil {
			res.AddPropertyDiff("Args")
		}
	} else {
		// we have args
		if other.Args == nil {
			res.AddPropertyDiff("Args")
		} else {
			if !c.Args.Equals(other.Args) {
				res.AddPropertyDiff("Args")
			}
		}
	}

	// query
	if c.Query == nil {
		if other.Query != nil {
			res.AddPropertyDiff("Query")
		}
	} else {
		// we have query
		if other.Query == nil {
			res.AddPropertyDiff("Query")
		} else {
			if !c.Query.Equals(other.Query) {
				res.AddPropertyDiff("Query")
			}
		}
	}

	// params
	if len(c.Params) != len(other.Params) {
		res.AddPropertyDiff("Params")
	} else {
		for i, p := range c.Params {
			if !p.Equals(other.Params[i]) {
				res.AddPropertyDiff("Params")
			}
		}
	}
	return res
}

func (c *Control) setBaseProperties() {
	if c.Base == nil {
		return
	}

	if c.Description == nil {
		c.Description = c.Base.Description
	}
	if c.Documentation == nil {
		c.Documentation = c.Base.Documentation
	}
	if c.SearchPath == nil {
		c.SearchPath = c.Base.SearchPath
	}
	if c.SearchPathPrefix == nil {
		c.SearchPathPrefix = c.Base.SearchPathPrefix
	}
	if c.Severity == nil {
		c.Severity = c.Base.Severity
	}
	if c.SQL == nil {
		c.SQL = c.Base.SQL
	}
	if c.Tags == nil {
		c.Tags = c.Base.Tags
	}
	if c.Title == nil {
		c.Title = c.Base.Title
	}
	if c.Query == nil {
		c.Query = c.Base.Query
	}
	if c.Args == nil {
		c.Args = c.Base.Args
	}
	if c.Params == nil {
		c.Params = c.Base.Params
	}
}
