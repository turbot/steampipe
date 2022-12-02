package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/constants"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// DashboardNode is a struct representing a leaf dashboard node
type DashboardNode struct {
	ResourceWithMetadataBase
	QueryProviderBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	FullName        string             `cty:"name" json:"name"`
	ShortName       string             `json:"-"`
	UnqualifiedName string             `json:"-"`
	Category        *DashboardCategory `cty:"category" hcl:"category" column:"category,jsonb" json:"category,omitempty"`

	// these properties are JSON serialised by the parent LeafRun
	Title *string `cty:"title" hcl:"title" column:"title,text" json:"-"`
	// QueryProvider
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"-"`
	Query                 *Query      `hcl:"query" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"-"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"-"`

	Base       *DashboardNode       `hcl:"base" json:"-"`
	DeclRange  hcl.Range            `json:"-"`
	References []*ResourceReference `json:"-"`
	Mod        *Mod                 `cty:"mod" json:"-"`
	Paths      []NodePath           `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewDashboardNode(block *hcl.Block, mod *Mod, shortName string) HclResource {
	c := &DashboardNode{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}

	c.SetAnonymous(block)
	return c
}

func (n *DashboardNode) Equals(other *DashboardNode) bool {
	diff := n.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (n *DashboardNode) CtyValue() (cty.Value, error) {
	return getCtyValue(n)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'edge.<shortName>'
func (n *DashboardNode) Name() string {
	return n.FullName
}

// OnDecoded implements HclResource
func (n *DashboardNode) OnDecoded(_ *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	n.setBaseProperties(resourceMapProvider)

	// when we reference resources (i.e. category),
	// not all properties are retrieved as they are no cty serialisable
	// repopulate category from resourceMapProvider
	if n.Category != nil {
		fullCategory, diags := enrichCategory(n.Category, n, resourceMapProvider)
		if diags.HasErrors() {
			return diags
		}
		n.Category = fullCategory
	}
	return nil
}

// AddReference implements ResourceWithMetadata
func (n *DashboardNode) AddReference(ref *ResourceReference) {
	n.References = append(n.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (n *DashboardNode) GetReferences() []*ResourceReference {
	return n.References
}

// GetMod implements ModTreeItem
func (n *DashboardNode) GetMod() *Mod {
	return n.Mod
}

// GetDeclRange implements HclResource
func (n *DashboardNode) GetDeclRange() *hcl.Range {
	return &n.DeclRange
}

// BlockType implements HclResource
func (*DashboardNode) BlockType() string {
	return BlockTypeNode
}

// AddParent implements ModTreeItem
func (n *DashboardNode) AddParent(parent ModTreeItem) error {
	n.parents = append(n.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (n *DashboardNode) GetParents() []ModTreeItem {
	return n.parents
}

// GetChildren implements ModTreeItem
func (n *DashboardNode) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements HclResource
func (n *DashboardNode) GetTitle() string {
	return typehelpers.SafeString(n.Title)
}

// GetDescription implements ModTreeItem
func (n *DashboardNode) GetDescription() string {
	return ""
}

// GetTags implements HclResource
func (n *DashboardNode) GetTags() map[string]string {
	return map[string]string{}
}

// GetPaths implements ModTreeItem
func (n *DashboardNode) GetPaths() []NodePath {
	// lazy load
	if len(n.Paths) == 0 {
		n.SetPaths()
	}

	return n.Paths
}

// SetPaths implements ModTreeItem
func (n *DashboardNode) SetPaths() {
	for _, parent := range n.parents {
		for _, parentPath := range parent.GetPaths() {
			n.Paths = append(n.Paths, append(parentPath, n.Name()))
		}
	}
}

func (n *DashboardNode) Diff(other *DashboardNode) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: n,
		Name: n.Name(),
	}

	if (n.Category == nil) != (other.Category == nil) {
		res.AddPropertyDiff("Category")
	}
	if n.Category != nil && !n.Category.Equals(other.Category) {
		res.AddPropertyDiff("Category")
	}

	res.populateChildDiffs(n, other)
	res.queryProviderDiff(n, other)
	res.dashboardLeafNodeDiff(n, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (n *DashboardNode) GetWidth() int {
	return 0
}

// GetDisplay implements DashboardLeafNode
func (n *DashboardNode) GetDisplay() string {
	return ""
}

// GetDocumentation implements DashboardLeafNode
func (n *DashboardNode) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (n *DashboardNode) GetType() string {
	return ""
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (n *DashboardNode) GetUnqualifiedName() string {
	return n.UnqualifiedName
}

// GetParams implements QueryProvider
func (n *DashboardNode) GetParams() []*ParamDef {
	return n.Params
}

// GetArgs implements QueryProvider
func (n *DashboardNode) GetArgs() *QueryArgs {
	return n.Args
}

// GetSQL implements QueryProvider
func (n *DashboardNode) GetSQL() *string {
	return n.SQL
}

// GetQuery implements QueryProvider
func (n *DashboardNode) GetQuery() *Query {
	return n.Query
}

// SetArgs implements QueryProvider
func (n *DashboardNode) SetArgs(args *QueryArgs) {
	n.Args = args
}

// SetParams implements QueryProvider
func (n *DashboardNode) SetParams(params []*ParamDef) {
	n.Params = params
}

// GetPreparedStatementName implements QueryProvider
func (n *DashboardNode) GetPreparedStatementName() string {
	if n.PreparedStatementName != "" {
		return n.PreparedStatementName
	}
	n.PreparedStatementName = n.buildPreparedStatementName(n.ShortName, n.Mod.NameWithVersion(), constants.PreparedStatementChartSuffix)
	return n.PreparedStatementName
}

// GetResolvedQuery implements QueryProvider
func (n *DashboardNode) GetResolvedQuery(runtimeArgs *QueryArgs) (*ResolvedQuery, error) {
	// defer to base
	return n.getResolvedQuery(n, runtimeArgs)
}

// IsSnapshotPanel implements SnapshotPanel
func (*DashboardNode) IsSnapshotPanel() {}

func (n *DashboardNode) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(n.Base, resourceMapProvider); !resolved {
		return
	} else {
		n.Base = base.(*DashboardNode)
	}

	if n.Title == nil {
		n.Title = n.Base.Title
	}

	if n.SQL == nil {
		n.SQL = n.Base.SQL
	}

	if n.Query == nil {
		n.Query = n.Base.Query
	}

	if n.Args == nil {
		n.Args = n.Base.Args
	}

	if n.Params == nil {
		n.Params = n.Base.Params
	}

	if n.Category == nil {
		n.Category = n.Base.Category
	}

	n.MergeRuntimeDependencies(n.Base)
}
