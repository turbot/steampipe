package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/constants"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/utils"
	"github.com/zclconf/go-cty/cty"
)

// DashboardHierarchy is a struct representing a leaf dashboard node
type DashboardHierarchy struct {
	ResourceWithMetadataBase
	QueryProviderBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	// these properties are JSON serialised by the parent LeafRun
	Title        *string                                `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width        *int                                   `cty:"width" hcl:"width" column:"width,text"  json:"-"`
	Type         *string                                `cty:"type" hcl:"type" column:"type,text"  json:"type,omitempty"`
	CategoryList DashboardHierarchyCategoryList         `cty:"category_list" hcl:"category,block" column:"category,jsonb" json:"-"`
	Categories   map[string]*DashboardHierarchyCategory `cty:"categories" json:"categories"`
	Display      *string                                `cty:"display" hcl:"display" json:"display,omitempty"`
	OnHooks      []*DashboardOn                         `cty:"on" hcl:"on,block" json:"on,omitempty"`

	// QueryProvider
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"-"`
	Query                 *Query      `hcl:"query" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb"json:"-"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"params,omitempty"`

	Base       *DashboardHierarchy  `hcl:"base" json:"-"`
	DeclRange  hcl.Range            `json:"-"`
	References []*ResourceReference `json:"-"`
	Mod        *Mod                 `cty:"mod" json:"-"`
	Paths      []NodePath           `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewDashboardHierarchy(block *hcl.Block, mod *Mod, shortName string) *DashboardHierarchy {
	h := &DashboardHierarchy{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}
	h.SetAnonymous(block)
	return h
}

func (h *DashboardHierarchy) Equals(other *DashboardHierarchy) bool {
	diff := h.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (h *DashboardHierarchy) CtyValue() (cty.Value, error) {
	return getCtyValue(h)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'chart.<shortName>'
func (h *DashboardHierarchy) Name() string {
	return h.FullName
}

// OnDecoded implements HclResource
func (h *DashboardHierarchy) OnDecoded(block *hcl.Block, resourceMapProvider ModResourcesProvider) hcl.Diagnostics {
	h.setBaseProperties(resourceMapProvider)
	// populate categories map
	if len(h.CategoryList) > 0 {
		h.Categories = make(map[string]*DashboardHierarchyCategory, len(h.CategoryList))
		for _, c := range h.CategoryList {
			h.Categories[c.Name] = c
		}
	}
	return nil
}

// AddReference implements HclResource
func (h *DashboardHierarchy) AddReference(ref *ResourceReference) {
	h.References = append(h.References, ref)
}

// GetReferences implements HclResource
func (h *DashboardHierarchy) GetReferences() []*ResourceReference {
	return h.References
}

// GetMod implements HclResource
func (h *DashboardHierarchy) GetMod() *Mod {
	return h.Mod
}

// GetDeclRange implements HclResource
func (h *DashboardHierarchy) GetDeclRange() *hcl.Range {
	return &h.DeclRange
}

// AddParent implements ModTreeItem
func (h *DashboardHierarchy) AddParent(parent ModTreeItem) error {
	h.parents = append(h.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (h *DashboardHierarchy) GetParents() []ModTreeItem {
	return h.parents
}

// GetChildren implements ModTreeItem
func (h *DashboardHierarchy) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements ModTreeItem
func (h *DashboardHierarchy) GetTitle() string {
	return typehelpers.SafeString(h.Title)
}

// GetDescription implements ModTreeItem
func (h *DashboardHierarchy) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (h *DashboardHierarchy) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (h *DashboardHierarchy) GetPaths() []NodePath {
	// lazy load
	if len(h.Paths) == 0 {
		h.SetPaths()
	}

	return h.Paths
}

// SetPaths implements ModTreeItem
func (h *DashboardHierarchy) SetPaths() {
	for _, parent := range h.parents {
		for _, parentPath := range parent.GetPaths() {
			h.Paths = append(h.Paths, append(parentPath, h.Name()))
		}
	}
}

func (h *DashboardHierarchy) Diff(other *DashboardHierarchy) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: h,
		Name: h.Name(),
	}

	if !utils.SafeStringsEqual(h.Type, other.Type) {
		res.AddPropertyDiff("Type")
	}

	if len(h.CategoryList) != len(other.CategoryList) {
		res.AddPropertyDiff("Categories")
	} else {
		for i, c := range h.Categories {
			if !c.Equals(other.Categories[i]) {
				res.AddPropertyDiff("Categories")
			}
		}
	}

	res.populateChildDiffs(h, other)
	res.queryProviderDiff(h, other)
	res.dashboardLeafNodeDiff(h, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (h *DashboardHierarchy) GetWidth() int {
	if h.Width == nil {
		return 0
	}
	return *h.Width
}

// GetDisplay implements DashboardLeafNode
func (h *DashboardHierarchy) GetDisplay() *string {
	return h.Display
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (h *DashboardHierarchy) GetUnqualifiedName() string {
	return h.UnqualifiedName
}

// GetParams implements QueryProvider
func (h *DashboardHierarchy) GetParams() []*ParamDef {
	return h.Params
}

// GetArgs implements QueryProvider
func (h *DashboardHierarchy) GetArgs() *QueryArgs {
	return h.Args
}

// GetSQL implements QueryProvider
func (h *DashboardHierarchy) GetSQL() *string {
	return h.SQL
}

// GetQuery implements QueryProvider
func (h *DashboardHierarchy) GetQuery() *Query {
	return h.Query
}

// SetArgs implements QueryProvider
func (h *DashboardHierarchy) SetArgs(args *QueryArgs) {
	h.Args = args
}

// SetParams implements QueryProvider
func (h *DashboardHierarchy) SetParams(params []*ParamDef) {
	h.Params = params
}

// GetPreparedStatementName implements QueryProvider
func (h *DashboardHierarchy) GetPreparedStatementName() string {
	if h.PreparedStatementName != "" {
		return h.PreparedStatementName
	}
	h.PreparedStatementName = h.buildPreparedStatementName(h.ShortName, h.Mod.NameWithVersion(), constants.PreparedStatementHierarchySuffix)
	return h.PreparedStatementName
}

// GetPreparedStatementExecuteSQL implements QueryProvider
func (h *DashboardHierarchy) GetPreparedStatementExecuteSQL(runtimeArgs *QueryArgs) (*ResolvedQuery, error) {
	// defer to base
	return h.getPreparedStatementExecuteSQL(h, runtimeArgs)
}

func (h *DashboardHierarchy) setBaseProperties(resourceMapProvider ModResourcesProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(h.Base, resourceMapProvider); !resolved {
		return
	} else {
		h.Base = base.(*DashboardHierarchy)
	}

	if h.Title == nil {
		h.Title = h.Base.Title
	}

	if h.Type == nil {
		h.Type = h.Base.Type
	}

	if h.Display == nil {
		h.Display = h.Base.Display
	}

	if h.Width == nil {
		h.Width = h.Base.Width
	}

	if h.SQL == nil {
		h.SQL = h.Base.SQL
	}

	if h.Query == nil {
		h.Query = h.Base.Query
	}

	if h.Args == nil {
		h.Args = h.Base.Args
	}

	if h.Params == nil {
		h.Params = h.Base.Params
	}

	if h.CategoryList == nil {
		h.CategoryList = h.Base.CategoryList
	} else {
		h.CategoryList.Merge(h.Base.CategoryList)
	}

	h.MergeRuntimeDependencies(h.Base)
}
