package modconfig

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/types"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

// Control is a struct representing the Control resource
type Control struct {
	ResourceWithMetadataBase
	QueryProviderBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	ShortName        string            `json:"-"`
	FullName         string            `cty:"name" json:"-"`
	Description      *string           `cty:"description" hcl:"description" column:"description,text" json:"-"`
	Documentation    *string           `cty:"documentation" hcl:"documentation"  column:"documentation,text" json:"-"`
	SearchPath       *string           `cty:"search_path" hcl:"search_path"  column:"search_path,text" json:"search_path,omitempty"`
	SearchPathPrefix *string           `cty:"search_path_prefix" hcl:"search_path_prefix"  column:"search_path_prefix,text" json:"search_path_prefix,omitempty"`
	Severity         *string           `cty:"severity" hcl:"severity"  column:"severity,text" json:"severity,omitempty"`
	Tags             map[string]string `cty:"tags" hcl:"tags,optional"  column:"tags,jsonb" json:"-"`
	Title            *string           `cty:"title" hcl:"title"  column:"title,text" json:"-"`

	// QueryProvider
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"sql,omitempty"`
	Query                 *Query      `hcl:"query" json:"query,omitempty"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"-"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"-"`

	References      []*ResourceReference ` json:"-"`
	Mod             *Mod                 `cty:"mod" json:"-"`
	DeclRange       hcl.Range            `json:"-"`
	UnqualifiedName string               `json:"-"`
	Paths           []NodePath           `json:"-"`

	// dashboard specific properties
	Base    *Control `hcl:"base" json:"-"`
	Width   *int     `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type    *string  `cty:"type" hcl:"type" column:"type,text" json:"-"`
	Display *string  `cty:"display" hcl:"display" json:"-"`

	parents []ModTreeItem
}

func NewControl(block *hcl.Block, mod *Mod, shortName string) HclResource {
	control := &Control{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.control.%s", mod.ShortName, shortName),
		UnqualifiedName: fmt.Sprintf("control.%s", shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
		Args:            NewQueryArgs(),
	}

	control.SetAnonymous(block)
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

// GetTitle implements HclResource
func (c *Control) GetTitle() string {
	return typehelpers.SafeString(c.Title)
}

// GetDescription implements ModTreeItem
func (c *Control) GetDescription() string {
	return typehelpers.SafeString(c.Description)
}

// GetTags implements HclResource
func (c *Control) GetTags() map[string]string {
	if c.Tags != nil {
		return c.Tags
	}
	return map[string]string{}
}

// GetChildren implements ModTreeItem
func (c *Control) GetChildren() []ModTreeItem {
	return nil
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
func (c *Control) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	c.setBaseProperties(resourceMapProvider)
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

// AddReference implements ResourceWithMetadata
func (c *Control) AddReference(ref *ResourceReference) {
	c.References = append(c.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (c *Control) GetReferences() []*ResourceReference {
	return c.References
}

// GetMod implements ModTreeItem
func (c *Control) GetMod() *Mod {
	return c.Mod
}

// GetDeclRange implements HclResource
func (c *Control) GetDeclRange() *hcl.Range {
	return &c.DeclRange
}

// BlockType implements HclResource
func (*Control) BlockType() string {
	return BlockTypeControl
}

// GetParams implements QueryProvider
func (c *Control) GetParams() []*ParamDef {
	return c.Params
}

// GetQuery implements QueryProvider
func (c *Control) GetQuery() *Query {
	return c.Query
}

// GetArgs implements QueryProvider
func (c *Control) GetArgs() *QueryArgs {
	return c.Args
}

// GetSQL implements QueryProvider
func (c *Control) GetSQL() *string {
	return c.SQL
}

// SetArgs implements QueryProvider
func (c *Control) SetArgs(args *QueryArgs) {
	c.Args = args
}

// SetParams implements QueryProvider
func (c *Control) SetParams(params []*ParamDef) {
	c.Params = params
}

// GetPreparedStatementName implements QueryProvider
func (c *Control) GetPreparedStatementName() string {
	if c.PreparedStatementName != "" {
		return c.PreparedStatementName
	}
	c.PreparedStatementName = c.buildPreparedStatementName(c.ShortName, c.Mod.NameWithVersion(), constants.PreparedStatementControlSuffix)
	return c.PreparedStatementName
}

// GetResolvedQuery implements QueryProvider
func (c *Control) GetResolvedQuery(runtimeArgs *QueryArgs) (*ResolvedQuery, error) {
	// defer to base
	return c.getResolvedQuery(c, runtimeArgs)
}

// GetWidth implements DashboardLeafNode
func (c *Control) GetWidth() int {
	if c.Width == nil {
		return 0
	}
	return *c.Width
}

// GetDisplay implements DashboardLeafNode
func (c *Control) GetDisplay() string {
	return ""
}

// GetDocumentation implements DashboardLeafNode, ModTreeItem
func (c *Control) GetDocumentation() string {
	return typehelpers.SafeString(c.Documentation)
}

// GetType implements DashboardLeafNode
func (c *Control) GetType() string {
	return typehelpers.SafeString(c.Type)
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (c *Control) GetUnqualifiedName() string {
	return c.UnqualifiedName
}

func (c *Control) Diff(other *Control) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: c,
		Name: c.Name(),
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
	if len(c.Tags) != len(other.Tags) {
		res.AddPropertyDiff("Tags")
	} else {
		for k, v := range c.Tags {
			if otherVal := other.Tags[k]; v != otherVal {
				res.AddPropertyDiff("Tags")
			}
		}
	}

	res.dashboardLeafNodeDiff(c, other)
	res.queryProviderDiff(c, other)

	return res
}

func (c *Control) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(c.Base, resourceMapProvider); !resolved {
		return
	} else {
		c.Base = base.(*Control)
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
	c.Tags = utils.MergeMaps(c.Tags, c.Base.Tags)
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
	if c.Width == nil {
		c.Width = c.Base.Width
	}
	if c.Type == nil {
		c.Type = c.Base.Type
	}
	if c.Display == nil {
		c.Display = c.Base.Display
	}
	c.MergeRuntimeDependencies(c.Base)
}
