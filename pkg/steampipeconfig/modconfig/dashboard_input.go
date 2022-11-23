package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

// DashboardInput is a struct representing a leaf dashboard node
type DashboardInput struct {
	ResourceWithMetadataBase
	QueryProviderBase

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `cty:"unqualified_name" json:"unqualified_name"`
	DashboardName   string `column:"dashboard,text" json:"-"`

	Label       *string                 `cty:"label" hcl:"label" column:"label,text" json:"label,omitempty"`
	Placeholder *string                 `cty:"placeholder" hcl:"placeholder" column:"placeholder,text" json:"placeholder,omitempty"`
	Options     []*DashboardInputOption `cty:"options" hcl:"option,block" json:"options,omitempty"`

	// these properties are JSON serialised by the parent LeafRun
	Title   *string `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width   *int    `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type    *string `cty:"type" hcl:"type" column:"type,text" json:"-"`
	Display *string `cty:"display" hcl:"display" json:"-"`

	// QueryProvider
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"-"`
	Query                 *Query      `hcl:"query" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb"json:"-"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"-"`

	Base       *DashboardInput      `hcl:"base" json:"-"`
	DeclRange  hcl.Range            `json:"-"`
	References []*ResourceReference `json:"-"`
	Mod        *Mod                 `cty:"mod" json:"-"`

	Paths     []NodePath `column:"path,jsonb" json:"-"`
	parents   []ModTreeItem
	dashboard *Dashboard
}

func (i *DashboardInput) Clone() *DashboardInput {
	return &DashboardInput{
		ResourceWithMetadataBase: i.ResourceWithMetadataBase,
		QueryProviderBase:        i.QueryProviderBase,
		FullName:                 i.FullName,
		ShortName:                i.ShortName,
		UnqualifiedName:          i.UnqualifiedName,
		Title:                    i.Title,
		Width:                    i.Width,
		Type:                     i.Type,
		Label:                    i.Label,
		Placeholder:              i.Placeholder,
		Display:                  i.Display,
		Options:                  i.Options,
		SQL:                      i.SQL,
		Query:                    i.Query,
		PreparedStatementName:    i.PreparedStatementName,
		Args:                     i.Args,
		Params:                   i.Params,
		DeclRange:                i.DeclRange,
		Mod:                      i.Mod,
		Paths:                    i.Paths,
		parents:                  i.parents,
		dashboard:                i.dashboard,
	}
}

func NewDashboardInput(block *hcl.Block, mod *Mod, shortName string) HclResource {
	// input cannot be anonymous
	i := &DashboardInput{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}
	return i
}

func (i *DashboardInput) Equals(other *DashboardInput) bool {
	diff := i.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (i *DashboardInput) CtyValue() (cty.Value, error) {
	return getCtyValue(i)
}

// IsSnapshotPanel implements SnapshotPanel
func (*DashboardInput) IsSnapshotPanel() {}

// Name implements HclResource, ModTreeItem
// return name in format: 'chart.<shortName>'
func (i *DashboardInput) Name() string {
	return i.FullName
}

// OnDecoded implements HclResource
func (i *DashboardInput) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	i.setBaseProperties(resourceMapProvider)
	return nil
}

// AddReference implements ResourceWithMetadata
func (i *DashboardInput) AddReference(ref *ResourceReference) {
	i.References = append(i.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (i *DashboardInput) GetReferences() []*ResourceReference {
	return i.References
}

// GetMod implements ModTreeItem
func (i *DashboardInput) GetMod() *Mod {
	return i.Mod
}

// GetDeclRange implements HclResource
func (i *DashboardInput) GetDeclRange() *hcl.Range {
	return &i.DeclRange
}

// BlockType implements HclResource
func (*DashboardInput) BlockType() string {
	return BlockTypeInput
}

// AddParent implements ModTreeItem
func (i *DashboardInput) AddParent(parent ModTreeItem) error {
	i.parents = append(i.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (i *DashboardInput) GetParents() []ModTreeItem {
	return i.parents
}

// GetChildren implements ModTreeItem
func (i *DashboardInput) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements HclResource
func (i *DashboardInput) GetTitle() string {
	return typehelpers.SafeString(i.Title)
}

// GetDescription implements ModTreeItem
func (i *DashboardInput) GetDescription() string {
	return ""
}

// GetTags implements HclResource
func (i *DashboardInput) GetTags() map[string]string {
	return map[string]string{}
}

// GetPaths implements ModTreeItem
func (i *DashboardInput) GetPaths() []NodePath {
	// lazy load
	if len(i.Paths) == 0 {
		i.SetPaths()
	}

	return i.Paths
}

// SetPaths implements ModTreeItem
func (i *DashboardInput) SetPaths() {
	for _, parent := range i.parents {
		for _, parentPath := range parent.GetPaths() {
			i.Paths = append(i.Paths, append(parentPath, i.Name()))
		}
	}
}

func (i *DashboardInput) Diff(other *DashboardInput) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: i,
		Name: i.Name(),
	}

	if !utils.SafeStringsEqual(i.Type, other.Type) {
		res.AddPropertyDiff("Type")
	}

	if !utils.SafeStringsEqual(i.Label, other.Label) {
		res.AddPropertyDiff("Label")
	}

	if !utils.SafeStringsEqual(i.Placeholder, other.Placeholder) {
		res.AddPropertyDiff("Placeholder")
	}

	if len(i.Options) != len(other.Options) {
		res.AddPropertyDiff("Options")
	} else {
		for idx, o := range i.Options {
			if !other.Options[idx].Equals(o) {
				res.AddPropertyDiff("Options")
			}
		}
	}

	res.populateChildDiffs(i, other)
	res.queryProviderDiff(i, other)
	res.dashboardLeafNodeDiff(i, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (i *DashboardInput) GetWidth() int {
	if i.Width == nil {
		return 0
	}
	return *i.Width
}

// GetDisplay implements DashboardLeafNode
func (i *DashboardInput) GetDisplay() string {
	return typehelpers.SafeString(i.Display)
}

// GetDocumentation implements DashboardLeafNode, ModTreeItem
func (*DashboardInput) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (i *DashboardInput) GetType() string {
	return typehelpers.SafeString(i.Type)
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (i *DashboardInput) GetUnqualifiedName() string {
	return i.UnqualifiedName
}

// SetDashboard sets the parent dashboard container
func (i *DashboardInput) SetDashboard(dashboard *Dashboard) {
	i.dashboard = dashboard
	i.DashboardName = dashboard.Name()
}

// GetParams implements QueryProvider
func (i *DashboardInput) GetParams() []*ParamDef {
	return i.Params
}

// GetArgs implements QueryProvider
func (i *DashboardInput) GetArgs() *QueryArgs {
	return i.Args
}

// GetSQL implements QueryProvider
func (i *DashboardInput) GetSQL() *string {
	return i.SQL
}

// GetQuery implements QueryProvider
func (i *DashboardInput) GetQuery() *Query {
	return i.Query
}

// VerifyQuery implements QueryProvider
func (i *DashboardInput) VerifyQuery(QueryProvider) error {
	// query is optional - nothing to do
	return nil
}

// SetArgs implements QueryProvider
func (i *DashboardInput) SetArgs(args *QueryArgs) {
	i.Args = args
}

// SetParams implements QueryProvider
func (i *DashboardInput) SetParams(params []*ParamDef) {
	i.Params = params
}

// GetPreparedStatementName implements QueryProvider
func (i *DashboardInput) GetPreparedStatementName() string {
	if i.PreparedStatementName != "" {
		return i.PreparedStatementName
	}
	i.PreparedStatementName = i.buildPreparedStatementName(i.ShortName, i.Mod.NameWithVersion(), constants.PreparedStatementInputSuffix)
	return i.PreparedStatementName
}

// GetResolvedQuery implements QueryProvider
func (i *DashboardInput) GetResolvedQuery(runtimeArgs *QueryArgs) (*ResolvedQuery, error) {
	// defer to base
	return i.getResolvedQuery(i, runtimeArgs)
}

// DependsOnInput returns whether this input has a runtime dependency on the given input
func (i *DashboardInput) DependsOnInput(changedInputName string) bool {
	for _, r := range i.runtimeDependencies {
		if r.SourceResource.GetUnqualifiedName() == changedInputName {
			return true
		}
	}
	return false
}

func (i *DashboardInput) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(i.Base, resourceMapProvider); !resolved {
		return
	} else {
		i.Base = base.(*DashboardInput)
	}

	if i.Title == nil {
		i.Title = i.Base.Title
	}

	if i.Type == nil {
		i.Type = i.Base.Type
	}

	if i.Display == nil {
		i.Display = i.Base.Display
	}

	if i.Label == nil {
		i.Label = i.Base.Label
	}

	if i.Placeholder == nil {
		i.Placeholder = i.Base.Placeholder
	}

	if i.Width == nil {
		i.Width = i.Base.Width
	}

	if i.SQL == nil {
		i.SQL = i.Base.SQL
	}

	if i.Query == nil {
		i.Query = i.Base.Query
	}

	if i.Args == nil {
		i.Args = i.Base.Args
	}

	if i.Params == nil {
		i.Params = i.Base.Params
	}

	i.MergeRuntimeDependencies(i.Base)
}
