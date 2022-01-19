package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// ReportText is a struct representing a leaf reporting node
type ReportText struct {
	FullName        string `cty:"name"`
	ShortName       string
	UnqualifiedName string

	Title *string     `cty:"title" hcl:"title" column:"title,text"`
	Type  *string     `cty:"type" hcl:"type" column:"type,text"`
	Value *string     `cty:"value" hcl:"value" column:"value,text"`
	Width *int        `cty:"width" hcl:"width" column:"width,text"`
	Base  *ReportText `hcl:"base"`

	DeclRange hcl.Range
	Mod       *Mod `cty:"mod"`

	Paths []NodePath `column:"path,jsonb"`

	parents   []ModTreeItem
	metadata  *ResourceMetadata
	anonymous bool
}

func NewReportText(block *hcl.Block) *ReportText {
	return &ReportText{
		DeclRange:       block.DefRange,
		ShortName:       block.Labels[0],
		FullName:        fmt.Sprintf("%s.%s", block.Type, block.Labels[0]),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, block.Labels[0]),
	}
}

func (p *ReportText) Equals(other *ReportText) bool {
	diff := p.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (p *ReportText) CtyValue() (cty.Value, error) {
	return getCtyValue(p)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'text.<shortName>'
func (p *ReportText) Name() string {
	return p.FullName
}

func (p *ReportText) SetAnonymous(anonymous bool) {
	p.anonymous = anonymous
}

func (p *ReportText) IsAnonymous() bool {
	return p.anonymous
}

// OnDecoded implements HclResource
func (p *ReportText) OnDecoded(*hcl.Block) hcl.Diagnostics {
	p.setBaseProperties()
	return nil
}

func (p *ReportText) setBaseProperties() {
	if p.Base == nil {
		return
	}
	if p.Title == nil {
		p.Title = p.Base.Title
	}
	if p.Type == nil {
		p.Type = p.Base.Type
	}
	if p.Value == nil {
		p.Value = p.Base.Value
	}

	if p.Width == nil {
		p.Width = p.Base.Width
	}
}

// AddReference implements HclResource
func (p *ReportText) AddReference(*ResourceReference) {}

// SetMod implements HclResource
func (p *ReportText) SetMod(mod *Mod) {
	p.Mod = mod
	// if this resource has a name, update to include the mod
	// TODO kai is this conditional needed?
	if p.UnqualifiedName != "" {
		p.FullName = fmt.Sprintf("%s.%s", p.Mod.ShortName, p.UnqualifiedName)
	}
}

// GetMod implements HclResource
func (p *ReportText) GetMod() *Mod {
	return p.Mod
}

// GetDeclRange implements HclResource
func (p *ReportText) GetDeclRange() *hcl.Range {
	return &p.DeclRange
}

// AddParent implements ModTreeItem
func (p *ReportText) AddParent(parent ModTreeItem) error {
	p.parents = append(p.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (p *ReportText) GetParents() []ModTreeItem {
	return p.parents
}

// GetChildren implements ModTreeItem
func (p *ReportText) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements ModTreeItem
func (p *ReportText) GetTitle() string {
	return typehelpers.SafeString(p.Title)
}

// GetDescription implements ModTreeItem
func (p *ReportText) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (p *ReportText) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (p *ReportText) GetPaths() []NodePath {
	// lazy load
	if len(p.Paths) == 0 {
		p.SetPaths()
	}

	return p.Paths
}

// SetPaths implements ModTreeItem
func (p *ReportText) SetPaths() {
	for _, parent := range p.parents {
		for _, parentPath := range parent.GetPaths() {
			p.Paths = append(p.Paths, append(parentPath, p.Name()))
		}
	}
}

// GetMetadata implements ResourceWithMetadata
func (p *ReportText) GetMetadata() *ResourceMetadata {
	return p.metadata
}

// SetMetadata implements ResourceWithMetadata
func (p *ReportText) SetMetadata(metadata *ResourceMetadata) {
	p.metadata = metadata
}

func (p *ReportText) Diff(other *ReportText) *ReportTreeItemDiffs {
	res := &ReportTreeItemDiffs{
		Item: p,
		Name: p.Name(),
	}
	if p.FullName != other.FullName {
		res.AddPropertyDiff("Name")
	}
	if typehelpers.SafeString(p.Title) != typehelpers.SafeString(other.Title) {
		res.AddPropertyDiff("Title")
	}

	if p.Width == nil || other.Width == nil {
		if !(p.Width == nil && other.Width == nil) {
			res.AddPropertyDiff("Width")
		}
	} else if *p.Width != *other.Width {
		res.AddPropertyDiff("Width")
	}

	if typehelpers.SafeString(p.Type) != typehelpers.SafeString(other.Type) {
		res.AddPropertyDiff("Type")
	}

	if typehelpers.SafeString(p.Value) != typehelpers.SafeString(other.Value) {
		res.AddPropertyDiff("Style")
	}

	res.populateChildDiffs(p, other)

	return res
}
