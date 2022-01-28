package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/utils"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// ReportHierarchy is a struct representing a leaf reporting node
type ReportHierarchy struct {
	HclResourceBase

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	// these properties are JSON serialised by the parent LeafRun
	Title        *string                             `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width        *int                                `cty:"width" hcl:"width" column:"width,text"  json:"-"`
	Type         *string                             `cty:"type" hcl:"type" column:"type,text"  json:"type,omitempty"`
	CategoryList ReportHierarchyCategoryList         `cty:"category_list" hcl:"category,block" column:"category,jsonb" json:"-"`
	Categories   map[string]*ReportHierarchyCategory `cty:"categories" json:"categories"`

	// QueryProvider
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"sql"`
	Query                 *Query      `hcl:"query" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"args"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"params"`

	Base *ReportHierarchy `hcl:"base" json:"-"`

	DeclRange hcl.Range  `json:"-"`
	Mod       *Mod       `cty:"mod" json:"-"`
	Paths     []NodePath `column:"path,jsonb" json:"-"`

	parents  []ModTreeItem
	metadata *ResourceMetadata
}

func NewReportHierarchy(block *hcl.Block) *ReportHierarchy {
	return &ReportHierarchy{
		DeclRange:       block.DefRange,
		ShortName:       block.Labels[0],
		FullName:        fmt.Sprintf("%s.%s", block.Type, block.Labels[0]),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, block.Labels[0]),
	}
}

func (h *ReportHierarchy) Equals(other *ReportHierarchy) bool {
	diff := h.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (h *ReportHierarchy) CtyValue() (cty.Value, error) {
	return getCtyValue(h)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'chart.<shortName>'
func (h *ReportHierarchy) Name() string {
	return h.FullName
}

// OnDecoded implements HclResource
func (h *ReportHierarchy) OnDecoded(*hcl.Block) hcl.Diagnostics {
	h.setBaseProperties()
	// populate categories map
	if len(h.CategoryList) > 0 {
		h.Categories = make(map[string]*ReportHierarchyCategory, len(h.CategoryList))
		for _, c := range h.CategoryList {
			h.Categories[c.Name] = c
		}
	}
	return nil
}

func (h *ReportHierarchy) setBaseProperties() {
	if h.Base == nil {
		return
	}
	if h.Title == nil {
		h.Title = h.Base.Title
	}
	if h.Type == nil {
		h.Type = h.Base.Type
	}

	if h.Width == nil {
		h.Width = h.Base.Width
	}
	if h.SQL == nil {
		h.SQL = h.Base.SQL
	}
	if h.CategoryList == nil {
		h.CategoryList = h.Base.CategoryList
	} else {

		h.CategoryList.Merge(h.Base.CategoryList)
	}
}

// AddReference implements HclResource
func (h *ReportHierarchy) AddReference(*ResourceReference) {}

// SetMod implements HclResource
func (h *ReportHierarchy) SetMod(mod *Mod) {
	h.Mod = mod
	h.FullName = fmt.Sprintf("%s.%s", h.Mod.ShortName, h.UnqualifiedName)
}

// GetMod implements HclResource
func (h *ReportHierarchy) GetMod() *Mod {
	return h.Mod
}

// GetDeclRange implements HclResource
func (h *ReportHierarchy) GetDeclRange() *hcl.Range {
	return &h.DeclRange
}

// AddParent implements ModTreeItem
func (h *ReportHierarchy) AddParent(parent ModTreeItem) error {
	h.parents = append(h.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (h *ReportHierarchy) GetParents() []ModTreeItem {
	return h.parents
}

// GetChildren implements ModTreeItem
func (h *ReportHierarchy) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements ModTreeItem
func (h *ReportHierarchy) GetTitle() string {
	return typehelpers.SafeString(h.Title)
}

// GetDescription implements ModTreeItem
func (h *ReportHierarchy) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (h *ReportHierarchy) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (h *ReportHierarchy) GetPaths() []NodePath {
	// lazy load
	if len(h.Paths) == 0 {
		h.SetPaths()
	}

	return h.Paths
}

// SetPaths implements ModTreeItem
func (h *ReportHierarchy) SetPaths() {
	for _, parent := range h.parents {
		for _, parentPath := range parent.GetPaths() {
			h.Paths = append(h.Paths, append(parentPath, h.Name()))
		}
	}
}

// GetMetadata implements ResourceWithMetadata
func (h *ReportHierarchy) GetMetadata() *ResourceMetadata {
	return h.metadata
}

// SetMetadata implements ResourceWithMetadata
func (h *ReportHierarchy) SetMetadata(metadata *ResourceMetadata) {
	h.metadata = metadata
}

func (h *ReportHierarchy) Diff(other *ReportHierarchy) *ReportTreeItemDiffs {
	res := &ReportTreeItemDiffs{
		Item: h,
		Name: h.Name(),
	}

	if !utils.SafeStringsEqual(h.FullName, other.FullName) {
		res.AddPropertyDiff("Name")
	}

	if !utils.SafeStringsEqual(h.Title, other.Title) {
		res.AddPropertyDiff("Title")
	}

	if !utils.SafeStringsEqual(h.SQL, other.SQL) {
		res.AddPropertyDiff("SQL")
	}

	if !utils.SafeIntEqual(h.Width, other.Width) {
		res.AddPropertyDiff("Width")
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

	return res
}

// GetWidth implements ReportLeafNode
func (h *ReportHierarchy) GetWidth() int {
	if h.Width == nil {
		return 0
	}
	return *h.Width
}

// GetUnqualifiedName implements ReportLeafNode
func (h *ReportHierarchy) GetUnqualifiedName() string {
	return h.UnqualifiedName
}

// GetParams implements QueryProvider
func (h *ReportHierarchy) GetParams() []*ParamDef {
	return h.Params
}

// GetSQL implements QueryProvider, ReportingLeafNode
func (h *ReportHierarchy) GetSQL() string {
	return typehelpers.SafeString(h.SQL)
}

// GetQuery implements QueryProvider
func (h *ReportHierarchy) GetQuery() *Query {
	return h.Query
}

// GetPreparedStatementName implements QueryProvider
func (h *ReportHierarchy) GetPreparedStatementName() string {
	// lazy load
	if h.PreparedStatementName == "" {
		h.PreparedStatementName = preparedStatementName(h)
	}
	return h.PreparedStatementName
}

// GetModName implements QueryProvider
func (h *ReportHierarchy) GetModName() string {
	return h.Mod.NameWithVersion()
}

// SetArgs implements QueryProvider
func (h *ReportHierarchy) SetArgs(args *QueryArgs) {
	// nothing
}

// SetParams implements QueryProvider
func (h *ReportHierarchy) SetParams(params []*ParamDef) {
	h.Params = params
}
