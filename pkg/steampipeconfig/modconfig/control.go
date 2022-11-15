package modconfig

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/types"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/utils"
)

// Control is a struct representing the Control resource
type Control struct {
	ResourceWithMetadataBase
	QueryProviderBase
	ModTreeItemBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	SearchPath       *string           `cty:"search_path" hcl:"search_path"  column:"search_path,text" json:"search_path,omitempty"`
	SearchPathPrefix *string           `cty:"search_path_prefix" hcl:"search_path_prefix"  column:"search_path_prefix,text" json:"search_path_prefix,omitempty"`
	Severity         *string           `cty:"severity" hcl:"severity"  column:"severity,text" json:"severity,omitempty"`
	Tags             map[string]string `cty:"tags" hcl:"tags,optional"  column:"tags,jsonb" json:"-"`
	Title            *string           `cty:"title" hcl:"title"  column:"title,text" json:"-"`

	// QueryProvider
	PreparedStatementName string               `column:"prepared_statement_name,text" json:"-"`
	References            []*ResourceReference ` json:"-"`
	Paths                 []NodePath           `json:"-"`

	// dashboard specific properties
	Base    *Control `hcl:"base" json:"-"`
	Width   *int     `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type    *string  `cty:"type" hcl:"type" column:"type,text" json:"-"`
	Display *string  `cty:"display" hcl:"display" json:"-"`

	parents []ModTreeItem
}

func NewControl(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)

	control := &Control{
		QueryProviderBase: QueryProviderBase{
			Args:               NewQueryArgs(),
			modNameWithVersion: mod.NameWithVersion(),
			HclResourceBase: HclResourceBase{
				FullName:        fullName,
				UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
				ShortName:       shortName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
		},
		ModTreeItemBase: ModTreeItemBase{
			Mod:      mod,
			fullName: fullName,
		},
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

// QualifiedNameWithVersion returns the name in format: '<modName>@version.control.<shortName>'
func (c *Control) QualifiedNameWithVersion() string {
	return fmt.Sprintf("%s.%s", c.Mod.NameWithVersion(), c.FullName)
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

// GetType implements DashboardLeafNode
func (c *Control) GetType() string {
	return typehelpers.SafeString(c.Type)
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
