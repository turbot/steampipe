package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// Panel is a struct representing the Panel resource
type Panel struct {
	FullName  string `cty:"name"`
	ShortName string

	Title *string `cty:"title" column:"title,text"`
	Type  *string `cty:"type" column:"type,text"`
	Width *int    `cty:"width" column:"width,text"`
	SQL   *string `cty:"sql" column:"sql,text"`
	Text  *string `cty:"text" column:"text,text"`

	DeclRange hcl.Range
	Mod       *Mod `cty:"mod"`

	Base  *Panel
	Paths []NodePath `column:"path,jsonb"`

	parents         []ModTreeItem
	metadata        *ResourceMetadata
	UnqualifiedName string
}

func NewPanel(block *hcl.Block) *Panel {
	panel := &Panel{
		ShortName:       block.Labels[0],
		FullName:        fmt.Sprintf("panel.%s", block.Labels[0]),
		UnqualifiedName: fmt.Sprintf("panel.%s", block.Labels[0]),
		DeclRange:       block.DefRange,
	}
	return panel
}

// CtyValue implements HclResource
func (p *Panel) CtyValue() (cty.Value, error) {
	return getCtyValue(p)
}

// Name implements HclResource
// return name in format: 'panel.<shortName>'
func (p *Panel) Name() string {
	return p.FullName
}

// OnDecoded implements HclResource
func (p *Panel) OnDecoded(*hcl.Block) hcl.Diagnostics {
	p.setBaseProperties()
	return nil
}

func (p *Panel) setBaseProperties() {
	if p.Base == nil {
		return
	}
	if p.Title == nil {
		p.Title = p.Base.Title
	}
	if p.Type == nil {
		p.Type = p.Base.Type
	}
	if p.Width == nil {
		p.Width = p.Base.Width
	}
	if p.SQL == nil {
		p.SQL = p.Base.SQL
	}
	if p.Text == nil {
		p.Text = p.Base.Text
	}
}

// AddReference implements HclResource
func (p *Panel) AddReference(*ResourceReference) {}

// SetMod implements HclResource
func (p *Panel) SetMod(mod *Mod) {
	p.Mod = mod
	p.UnqualifiedName = p.FullName
	p.FullName = fmt.Sprintf("%s.%s", mod.ShortName, p.FullName)
}

// GetMod implements HclResource
func (p *Panel) GetMod() *Mod {
	return p.Mod
}

// GetDeclRange implements HclResource
func (p *Panel) GetDeclRange() *hcl.Range {
	return &p.DeclRange
}

// AddChild implements ModTreeItem
func (p *Panel) AddChild(ModTreeItem) error {
	return nil
}

// AddParent implements ModTreeItem
func (p *Panel) AddParent(parent ModTreeItem) error {
	p.parents = append(p.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (p *Panel) GetParents() []ModTreeItem {
	return p.parents
}

// GetChildren implements ModTreeItem
func (p *Panel) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements ModTreeItem
func (p *Panel) GetTitle() string {
	return typehelpers.SafeString(p.Title)
}

// GetDescription implements ModTreeItem
func (p *Panel) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (p *Panel) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (p *Panel) GetPaths() []NodePath {
	// lazy load
	if len(p.Paths) == 0 {
		p.SetPaths()
	}

	return p.Paths
}

// SetPaths implements ModTreeItem
func (p *Panel) SetPaths() {
	for _, parent := range p.parents {
		for _, parentPath := range parent.GetPaths() {
			p.Paths = append(p.Paths, append(parentPath, p.Name()))
		}
	}
}

// GetMetadata implements ResourceWithMetadata
func (p *Panel) GetMetadata() *ResourceMetadata {
	return p.metadata
}

// SetMetadata implements ResourceWithMetadata
func (p *Panel) SetMetadata(metadata *ResourceMetadata) {
	p.metadata = metadata
}

func (p *Panel) Diff(new *Panel) *ReportTreeItemDiffs {
	res := &ReportTreeItemDiffs{
		Item: p,
		Name: p.Name(),
	}
	if typehelpers.SafeString(p.Title) != typehelpers.SafeString(new.Title) {
		res.AddPropertyDiff("Title")
	}
	if typehelpers.SafeString(p.SQL) != typehelpers.SafeString(new.SQL) {
		res.AddPropertyDiff("SQL")
	}
	if typehelpers.SafeString(p.Text) != typehelpers.SafeString(new.Text) {
		res.AddPropertyDiff("Text")
	}
	if typehelpers.SafeString(p.Type) != typehelpers.SafeString(new.Type) {
		res.AddPropertyDiff("Type")
	}
	if p.Width == nil || new.Width == nil {
		if !(p.Width == nil && new.Width == nil) {
			res.AddPropertyDiff("Width")
		}
	} else if *p.Width != *new.Width {
		res.AddPropertyDiff("Width")
	}

	res.populateChildDiffs(p, new)

	return res
}
