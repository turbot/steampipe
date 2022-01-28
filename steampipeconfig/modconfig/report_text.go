package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/utils"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// ReportText is a struct representing a leaf reporting node
type ReportText struct {
	HclResourceBase

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	// these properties are JSON serialised by the parent LeafRun
	Title *string     `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width *int        `cty:"width" hcl:"width" column:"width,text"  json:"-"`
	Type  *string     `cty:"type" hcl:"type" column:"type,text"  json:"type,omitempty"`
	Value *string     `cty:"value" hcl:"value" column:"value,text"  json:"value,omitempty"`
	Base  *ReportText `hcl:"base" json:"-"`

	DeclRange hcl.Range  `json:"-"`
	Mod       *Mod       `cty:"mod" json:"-"`
	Paths     []NodePath `column:"path,jsonb" json:"-"`

	parents  []ModTreeItem
	metadata *ResourceMetadata
}

func NewReportText(block *hcl.Block) *ReportText {
	return &ReportText{
		DeclRange:       block.DefRange,
		ShortName:       block.Labels[0],
		FullName:        fmt.Sprintf("%s.%s", block.Type, block.Labels[0]),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, block.Labels[0]),
	}
}

func (t *ReportText) Equals(other *ReportText) bool {
	diff := t.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (t *ReportText) CtyValue() (cty.Value, error) {
	return getCtyValue(t)
}

// Name implements HclResource, ModTreeItem, ReportLeafNode
// return name in format: 'text.<shortName>'
func (t *ReportText) Name() string {
	return t.FullName
}

// OnDecoded implements HclResource
func (t *ReportText) OnDecoded(*hcl.Block) hcl.Diagnostics {
	t.setBaseProperties()
	return nil
}

func (t *ReportText) setBaseProperties() {
	if t.Base == nil {
		return
	}
	if t.Title == nil {
		t.Title = t.Base.Title
	}
	if t.Type == nil {
		t.Type = t.Base.Type
	}
	if t.Value == nil {
		t.Value = t.Base.Value
	}
	if t.Width == nil {
		t.Width = t.Base.Width
	}
}

// AddReference implements HclResource
func (t *ReportText) AddReference(*ResourceReference) {}

// SetMod implements HclResource
func (t *ReportText) SetMod(mod *Mod) {
	t.Mod = mod
	t.FullName = fmt.Sprintf("%s.%s", t.Mod.ShortName, t.UnqualifiedName)
}

// GetMod implements HclResource
func (t *ReportText) GetMod() *Mod {
	return t.Mod
}

// GetDeclRange implements HclResource
func (t *ReportText) GetDeclRange() *hcl.Range {
	return &t.DeclRange
}

// AddParent implements ModTreeItem
func (t *ReportText) AddParent(parent ModTreeItem) error {
	t.parents = append(t.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (t *ReportText) GetParents() []ModTreeItem {
	return t.parents
}

// GetChildren implements ModTreeItem
func (t *ReportText) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements ModTreeItem
func (t *ReportText) GetTitle() string {
	return typehelpers.SafeString(t.Title)
}

// GetDescription implements ModTreeItem
func (t *ReportText) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (t *ReportText) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (t *ReportText) GetPaths() []NodePath {
	// lazy load
	if len(t.Paths) == 0 {
		t.SetPaths()
	}

	return t.Paths
}

// SetPaths implements ModTreeItem
func (t *ReportText) SetPaths() {
	for _, parent := range t.parents {
		for _, parentPath := range parent.GetPaths() {
			t.Paths = append(t.Paths, append(parentPath, t.Name()))
		}
	}
}

// GetMetadata implements ResourceWithMetadata
func (t *ReportText) GetMetadata() *ResourceMetadata {
	return t.metadata
}

// SetMetadata implements ResourceWithMetadata
func (t *ReportText) SetMetadata(metadata *ResourceMetadata) {
	t.metadata = metadata
}

func (t *ReportText) Diff(other *ReportText) *ReportTreeItemDiffs {
	res := &ReportTreeItemDiffs{
		Item: t,
		Name: t.Name(),
	}

	if !utils.SafeStringsEqual(t.FullName, other.FullName) {
		res.AddPropertyDiff("Title")
	}

	if !utils.SafeStringsEqual(t.Title, other.Title) {
		res.AddPropertyDiff("Title")
	}

	if !utils.SafeIntEqual(t.Width, other.Width) {
		res.AddPropertyDiff("Width")
	}

	if !utils.SafeStringsEqual(t.Type, other.Type) {
		res.AddPropertyDiff("Type")
	}

	if !utils.SafeStringsEqual(t.Value, other.Value) {
		res.AddPropertyDiff("Value")
	}

	res.populateChildDiffs(t, other)

	return res
}

// GetSQL implements ReportLeafNode
func (t *ReportText) GetSQL() string {
	return ""
}

// GetWidth implements ReportLeafNode
func (t *ReportText) GetWidth() int {
	if t.Width == nil {
		return 0
	}
	return *t.Width
}

// GetUnqualifiedName implements ReportLeafNode
func (t *ReportText) GetUnqualifiedName() string {
	return t.UnqualifiedName
}
