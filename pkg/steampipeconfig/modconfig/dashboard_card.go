package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/constants"

	"github.com/turbot/steampipe/pkg/utils"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// DashboardCard is a struct representing a leaf dashboard node
type DashboardCard struct {
	ResourceWithMetadataBase
	QueryProviderBase

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	Label *string `cty:"label" hcl:"label" column:"label,text" json:"label,omitempty"`
	Value *string `cty:"value" hcl:"value" column:"value,text" json:"value,omitempty"`
	Icon  *string `cty:"icon" hcl:"icon" column:"icon,text" json:"icon,omitempty"`
	HREF  *string `cty:"href" hcl:"href" json:"href,omitempty"`

	// these properties are JSON serialised by the parent LeafRun
	Title   *string `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width   *int    `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type    *string `cty:"type" hcl:"type" column:"type,text" json:"-"`
	Display *string `cty:"display" hcl:"display" json:"-"`

	// QueryProvider
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"-"`
	Query                 *Query      `hcl:"query" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"-"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"-"`

	Base       *DashboardCard       `hcl:"base" json:"-"`
	DeclRange  hcl.Range            `json:"-"`
	References []*ResourceReference `json:"-"`
	Mod        *Mod                 `cty:"mod" json:"-"`
	Paths      []NodePath           `column:"path,jsonb" json:"-"`

	parents  []ModTreeItem
	metadata *ResourceMetadata
}

func NewDashboardCard(block *hcl.Block, mod *Mod, shortName string) HclResource {
	c := &DashboardCard{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}

	c.SetAnonymous(block)
	return c
}

func (c *DashboardCard) Equals(other *DashboardCard) bool {
	diff := c.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (c *DashboardCard) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'card.<shortName>'
func (c *DashboardCard) Name() string {
	return c.FullName
}

// OnDecoded implements HclResource
func (c *DashboardCard) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	c.setBaseProperties(resourceMapProvider)
	return nil
}

// AddReference implements ResourceWithMetadata
func (c *DashboardCard) AddReference(ref *ResourceReference) {
	c.References = append(c.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (c *DashboardCard) GetReferences() []*ResourceReference {
	return c.References
}

// GetMod implements ModTreeItem
func (c *DashboardCard) GetMod() *Mod {
	return c.Mod
}

// GetDeclRange implements HclResource
func (c *DashboardCard) GetDeclRange() *hcl.Range {
	return &c.DeclRange
}

// BlockType implements HclResource
func (*DashboardCard) BlockType() string {
	return BlockTypeCard
}

// AddParent implements ModTreeItem
func (c *DashboardCard) AddParent(parent ModTreeItem) error {
	c.parents = append(c.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (c *DashboardCard) GetParents() []ModTreeItem {
	return c.parents
}

// GetChildren implements ModTreeItem
func (c *DashboardCard) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements HclResource
func (c *DashboardCard) GetTitle() string {
	return typehelpers.SafeString(c.Title)
}

// GetDescription implements ModTreeItem
func (c *DashboardCard) GetDescription() string {
	return ""
}

// GetTags implements HclResource
func (c *DashboardCard) GetTags() map[string]string {
	return map[string]string{}
}

// GetPaths implements ModTreeItem
func (c *DashboardCard) GetPaths() []NodePath {
	// lazy load
	if len(c.Paths) == 0 {
		c.SetPaths()
	}

	return c.Paths
}

// SetPaths implements ModTreeItem
func (c *DashboardCard) SetPaths() {
	for _, parent := range c.parents {
		for _, parentPath := range parent.GetPaths() {
			c.Paths = append(c.Paths, append(parentPath, c.Name()))
		}
	}
}

func (c *DashboardCard) Diff(other *DashboardCard) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: c,
		Name: c.Name(),
	}

	if !utils.SafeStringsEqual(c.Label, other.Label) {
		res.AddPropertyDiff("Label")
	}

	if !utils.SafeStringsEqual(c.Value, other.Value) {
		res.AddPropertyDiff("Value")
	}

	if !utils.SafeStringsEqual(c.Type, other.Type) {
		res.AddPropertyDiff("Type")
	}

	if !utils.SafeStringsEqual(c.Icon, other.Icon) {
		res.AddPropertyDiff("Icon")
	}

	if !utils.SafeStringsEqual(c.HREF, other.HREF) {
		res.AddPropertyDiff("HREF")
	}

	res.populateChildDiffs(c, other)
	res.queryProviderDiff(c, other)
	res.dashboardLeafNodeDiff(c, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (c *DashboardCard) GetWidth() int {
	if c.Width == nil {
		return 0
	}
	return *c.Width
}

// GetDisplay implements DashboardLeafNode
func (c *DashboardCard) GetDisplay() string {
	return typehelpers.SafeString(c.Display)
}

// GetDocumentation implements DashboardLeafNode, ModTreeItem
func (c *DashboardCard) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (c *DashboardCard) GetType() string {
	return typehelpers.SafeString(c.Type)
}

// GetUnqualifiedName implements DashboardLeafNode
func (c *DashboardCard) GetUnqualifiedName() string {
	return c.UnqualifiedName
}

// GetParams implements QueryProvider
func (c *DashboardCard) GetParams() []*ParamDef {
	return c.Params
}

// GetArgs implements QueryProvider
func (c *DashboardCard) GetArgs() *QueryArgs {
	return c.Args
}

// GetSQL implements QueryProvider
func (c *DashboardCard) GetSQL() *string {
	return c.SQL
}

// GetQuery implements QueryProvider
func (c *DashboardCard) GetQuery() *Query {
	return c.Query
}

// VerifyQuery implements QueryProvider
func (c *DashboardCard) VerifyQuery(QueryProvider) error {
	// query is optional - nothing to do
	return nil
}

// SetArgs implements QueryProvider
func (c *DashboardCard) SetArgs(args *QueryArgs) {
	c.Args = args
}

// SetParams implements QueryProvider
func (c *DashboardCard) SetParams(params []*ParamDef) {
	c.Params = params
}

// GetPreparedStatementName implements QueryProvider
func (c *DashboardCard) GetPreparedStatementName() string {
	if c.PreparedStatementName != "" {
		return c.PreparedStatementName
	}
	c.PreparedStatementName = c.buildPreparedStatementName(c.ShortName, c.Mod.NameWithVersion(), constants.PreparedStatementCardSuffix)
	return c.PreparedStatementName
}

// GetResolvedQuery implements QueryProvider
func (c *DashboardCard) GetResolvedQuery(runtimeArgs *QueryArgs) (*ResolvedQuery, error) {
	// defer to base
	return c.getResolvedQuery(c, runtimeArgs)
}

func (c *DashboardCard) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(c.Base, resourceMapProvider); !resolved {
		return
	} else {
		c.Base = base.(*DashboardCard)
	}

	if c.Title == nil {
		c.Title = c.Base.Title
	}

	if c.Label == nil {
		c.Label = c.Base.Label
	}

	if c.Value == nil {
		c.Value = c.Base.Value
	}

	if c.Type == nil {
		c.Type = c.Base.Type
	}

	if c.Display == nil {
		c.Display = c.Base.Display
	}

	if c.Icon == nil {
		c.Icon = c.Base.Icon
	}

	if c.HREF == nil {
		c.HREF = c.Base.HREF
	}

	if c.Width == nil {
		c.Width = c.Base.Width
	}

	if c.SQL == nil {
		c.SQL = c.Base.SQL
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

	c.MergeRuntimeDependencies(c.Base)
}
