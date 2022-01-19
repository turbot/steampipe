package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// ReportTable is a struct representing a leaf reporting node
type ReportTable struct {
	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	Title *string      `cty:"title" hcl:"title" column:"title,text" json:"title,omitempty"`
	Type  *string      `cty:"type" hcl:"type" column:"type,text"  json:"type,omitempty"`
	Width *int         `cty:"width" hcl:"width" column:"width,text"  json:"width,omitempty"`
	SQL   *string      `cty:"sql" hcl:"sql" column:"sql,text" json:"sql"`
	Base  *ReportTable `hcl:"base" json:"-"`

	DeclRange hcl.Range `json:"-"`
	Mod       *Mod      `cty:"mod" json:"-"`

	Paths []NodePath `column:"path,jsonb" json:"-"`

	parents   []ModTreeItem
	metadata  *ResourceMetadata
	anonymous bool
}

func NewReportTable(block *hcl.Block) *ReportTable {
	return &ReportTable{
		DeclRange:       block.DefRange,
		ShortName:       block.Labels[0],
		FullName:        fmt.Sprintf("%s.%s", block.Type, block.Labels[0]),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, block.Labels[0]),
	}
}

func (t *ReportTable) Equals(other *ReportTable) bool {
	diff := t.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (t *ReportTable) CtyValue() (cty.Value, error) {
	return getCtyValue(t)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'table.<shortName>'
func (t *ReportTable) Name() string {
	return t.FullName
}

func (t *ReportTable) SetAnonymous(anonymous bool) {
	t.anonymous = anonymous
}

func (t *ReportTable) IsAnonymous() bool {
	return t.anonymous
}

// OnDecoded implements HclResource
func (t *ReportTable) OnDecoded(*hcl.Block) hcl.Diagnostics {
	t.setBaseProperties()
	return nil
}

func (t *ReportTable) setBaseProperties() {
	if t.Base == nil {
		return
	}
	if t.Title == nil {
		t.Title = t.Base.Title
	}
	if t.Type == nil {
		t.Type = t.Base.Type
	}

	if t.Width == nil {
		t.Width = t.Base.Width
	}
	if t.SQL == nil {
		t.SQL = t.Base.SQL
	}
}

// AddReference implements HclResource
func (t *ReportTable) AddReference(*ResourceReference) {}

// SetMod implements HclResource
func (t *ReportTable) SetMod(mod *Mod) {
	t.Mod = mod
	// if this resource has a name, update to include the mod
	// TODO kai is this conditional needed?
	if t.UnqualifiedName != "" {
		t.FullName = fmt.Sprintf("%s.%s", t.Mod.ShortName, t.UnqualifiedName)
	}
}

// GetMod implements HclResource
func (t *ReportTable) GetMod() *Mod {
	return t.Mod
}

// GetDeclRange implements HclResource
func (t *ReportTable) GetDeclRange() *hcl.Range {
	return &t.DeclRange
}

// AddParent implements ModTreeItem
func (t *ReportTable) AddParent(parent ModTreeItem) error {
	t.parents = append(t.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (t *ReportTable) GetParents() []ModTreeItem {
	return t.parents
}

// GetChildren implements ModTreeItem
func (t *ReportTable) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements ModTreeItem
func (t *ReportTable) GetTitle() string {
	return typehelpers.SafeString(t.Title)
}

// GetDescription implements ModTreeItem
func (t *ReportTable) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (t *ReportTable) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (t *ReportTable) GetPaths() []NodePath {
	// lazy load
	if len(t.Paths) == 0 {
		t.SetPaths()
	}

	return t.Paths
}

// SetPaths implements ModTreeItem
func (t *ReportTable) SetPaths() {
	for _, parent := range t.parents {
		for _, parentPath := range parent.GetPaths() {
			t.Paths = append(t.Paths, append(parentPath, t.Name()))
		}
	}
}

// GetMetadata implements ResourceWithMetadata
func (t *ReportTable) GetMetadata() *ResourceMetadata {
	return t.metadata
}

// SetMetadata implements ResourceWithMetadata
func (t *ReportTable) SetMetadata(metadata *ResourceMetadata) {
	t.metadata = metadata
}

func (t *ReportTable) Diff(other *ReportTable) *ReportTreeItemDiffs {
	res := &ReportTreeItemDiffs{
		Item: t,
		Name: t.Name(),
	}
	if t.FullName != other.FullName {
		res.AddPropertyDiff("Name")
	}
	if typehelpers.SafeString(t.Title) != typehelpers.SafeString(other.Title) {
		res.AddPropertyDiff("Title")
	}
	if typehelpers.SafeString(t.SQL) != typehelpers.SafeString(other.SQL) {
		res.AddPropertyDiff("SQL")
	}

	if t.Width == nil || other.Width == nil {
		if !(t.Width == nil && other.Width == nil) {
			res.AddPropertyDiff("Width")
		}
	} else if *t.Width != *other.Width {
		res.AddPropertyDiff("Width")
	}

	if typehelpers.SafeString(t.Type) != typehelpers.SafeString(other.Type) {
		res.AddPropertyDiff("Type")
	}

	res.populateChildDiffs(t, other)

	return res
}

// GetSQL implements ReportLeafNode
func (t *ReportTable) GetSQL() *string {
	return t.SQL
}
