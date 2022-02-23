package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/constants"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/utils"
	"github.com/zclconf/go-cty/cty"
)

// DashboardImage is a struct representing a leaf dashboard node
type DashboardImage struct {
	ResourceWithMetadataBase
	QueryProviderBase

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	// these properties are JSON serialised by the parent LeafRun
	Title   *string        `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width   *int           `cty:"width" hcl:"width" column:"width,text"  json:"-"`
	Src     *string        `cty:"src" hcl:"src" column:"src,text"  json:"src,omitempty"`
	Alt     *string        `cty:"alt" hcl:"alt" column:"alt,text"  json:"alt,omitempty"`
	Display *string        `cty:"display" hcl:"display" json:"display,omitempty"`
	OnHooks []*DashboardOn `cty:"on" hcl:"on,block" json:"on,omitempty"`

	// QueryProvider
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"-"`
	Query                 *Query      `hcl:"query" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"args"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"params"`

	Base *DashboardImage `hcl:"base" json:"-"`

	DeclRange  hcl.Range `json:"-"`
	References []*ResourceReference
	Mod        *Mod       `cty:"mod" json:"-"`
	Paths      []NodePath `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewDashboardImage(block *hcl.Block, mod *Mod, shortName string) *DashboardImage {
	i := &DashboardImage{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}
	i.SetAnonymous(block)
	return i
}

func (i *DashboardImage) Equals(other *DashboardImage) bool {
	diff := i.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (i *DashboardImage) CtyValue() (cty.Value, error) {
	return getCtyValue(i)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'image.<shortName>'
func (i *DashboardImage) Name() string {
	return i.FullName
}

// OnDecoded implements HclResource
func (i *DashboardImage) OnDecoded(*hcl.Block) hcl.Diagnostics {
	i.setBaseProperties()
	return nil
}

func (i *DashboardImage) setBaseProperties() {
	if i.Base == nil {
		return
	}
	if i.Title == nil {
		i.Title = i.Base.Title
	}
	if i.Src == nil {
		i.Src = i.Base.Src
	}
	if i.Alt == nil {
		i.Alt = i.Base.Alt
	}
	if i.Width == nil {
		i.Width = i.Base.Width
	}
}

// AddReference implements HclResource
func (i *DashboardImage) AddReference(ref *ResourceReference) {
	i.References = append(i.References, ref)
}

// GetReferences implements HclResource
func (i *DashboardImage) GetReferences() []*ResourceReference {
	return i.References
}

// GetMod implements HclResource
func (i *DashboardImage) GetMod() *Mod {
	return i.Mod
}

// GetDeclRange implements HclResource
func (i *DashboardImage) GetDeclRange() *hcl.Range {
	return &i.DeclRange
}

// AddParent implements ModTreeItem
func (i *DashboardImage) AddParent(parent ModTreeItem) error {
	i.parents = append(i.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (i *DashboardImage) GetParents() []ModTreeItem {
	return i.parents
}

// GetChildren implements ModTreeItem
func (i *DashboardImage) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements ModTreeItem
func (i *DashboardImage) GetTitle() string {
	return typehelpers.SafeString(i.Title)
}

// GetDescription implements ModTreeItem
func (i *DashboardImage) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (i *DashboardImage) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (i *DashboardImage) GetPaths() []NodePath {
	// lazy load
	if len(i.Paths) == 0 {
		i.SetPaths()
	}

	return i.Paths
}

// SetPaths implements ModTreeItem
func (i *DashboardImage) SetPaths() {
	for _, parent := range i.parents {
		for _, parentPath := range parent.GetPaths() {
			i.Paths = append(i.Paths, append(parentPath, i.Name()))
		}
	}
}

func (i *DashboardImage) Diff(other *DashboardImage) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: i,
		Name: i.Name(),
	}
	if !utils.SafeStringsEqual(i.FullName, other.FullName) {
		res.AddPropertyDiff("Name")
	}

	if !utils.SafeStringsEqual(i.Title, other.Title) {
		res.AddPropertyDiff("Title")
	}

	if !utils.SafeIntEqual(i.Width, other.Width) {
		res.AddPropertyDiff("Width")
	}

	if !utils.SafeStringsEqual(i.Src, other.Src) {
		res.AddPropertyDiff("Src")
	}

	if !utils.SafeStringsEqual(i.Alt, other.Alt) {
		res.AddPropertyDiff("Alt")
	}

	res.populateChildDiffs(i, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (i *DashboardImage) GetWidth() int {
	if i.Width == nil {
		return 0
	}
	return *i.Width
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (i *DashboardImage) GetUnqualifiedName() string {
	return i.UnqualifiedName
}

// GetParams implements QueryProvider
func (i *DashboardImage) GetParams() []*ParamDef {
	return i.Params
}

// GetArgs implements QueryProvider
func (i *DashboardImage) GetArgs() *QueryArgs {
	return i.Args

}

// GetSQL implements QueryProvider
func (i *DashboardImage) GetSQL() *string {
	return i.SQL
}

// GetQuery implements QueryProvider
func (i *DashboardImage) GetQuery() *Query {
	return i.Query
}

// SetArgs implements QueryProvider
func (i *DashboardImage) SetArgs(args *QueryArgs) {
	i.Args = args
}

// SetParams implements QueryProvider
func (i *DashboardImage) SetParams(params []*ParamDef) {
	i.Params = params
}

// GetPreparedStatementName implements QueryProvider
func (i *DashboardImage) GetPreparedStatementName() string {
	if i.PreparedStatementName != "" {
		return i.PreparedStatementName
	}
	i.PreparedStatementName = i.buildPreparedStatementName(i.ShortName, i.Mod.NameWithVersion(), constants.PreparedStatementImageSuffix)
	return i.PreparedStatementName
}

// GetPreparedStatementExecuteSQL implements QueryProvider
func (i *DashboardImage) GetPreparedStatementExecuteSQL(args *QueryArgs) (string, error) {
	// defer to base
	return i.getPreparedStatementExecuteSQL(i, args)
}
