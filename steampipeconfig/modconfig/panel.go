package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// Panel is a struct representing the Panel resource
type Panel struct {
	FullName        string `cty:"name"`
	ShortName       string
	UnqualifiedName string

	Title      *string           `cty:"title" column:"title,text"`
	Type       *string           `cty:"type" column:"type,text"`
	Width      *int              `cty:"width" column:"width,text"`
	SQL        *string           `cty:"sql" column:"sql,text"`
	Properties map[string]string `cty:"properties" column:"properties,jsonb"`

	DeclRange hcl.Range
	Mod       *Mod `cty:"mod"`

	Base  *Panel
	Paths []NodePath `column:"path,jsonb"`

	parents  []ModTreeItem
	metadata *ResourceMetadata
}

func NewPanel(block *hcl.Block) *Panel {
	name := "anon____________"
	if len(block.Labels) > 0 {
		name = block.Labels[0]
	}
	return &Panel{
		DeclRange:       block.DefRange,
		Properties:      make(map[string]string),
		ShortName:       name,
		FullName:        fmt.Sprintf("%s.%s", block.Type, name),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, name),
	}
}

func (p *Panel) Equals(other *Panel) bool {
	diff := p.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (p *Panel) CtyValue() (cty.Value, error) {
	return getCtyValue(p)
}

// Name implements HclResource, ModTreeItem
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
	for k, v := range p.Base.Properties {
		if _, ok := p.Properties[k]; !ok {
			p.Properties[k] = v
		}
	}
}

// AddReference implements HclResource
func (p *Panel) AddReference(*ResourceReference) {}

// SetMod implements HclResource
func (p *Panel) SetMod(mod *Mod) {
	p.Mod = mod
	// if this is a top level resource, and not a child, the resource names will already be set
	// - we need to update the full name to include the mod
	if p.UnqualifiedName != "" {
		// add mod name to full name
		p.FullName = fmt.Sprintf("%s.%s", p.Mod.ShortName, p.UnqualifiedName)
	}
}

// GetMod implements HclResource
func (p *Panel) GetMod() *Mod {
	return p.Mod
}

// GetDeclRange implements HclResource
func (p *Panel) GetDeclRange() *hcl.Range {
	return &p.DeclRange
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

func (p *Panel) Diff(other *Panel) *ReportTreeItemDiffs {
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
	if typehelpers.SafeString(p.SQL) != typehelpers.SafeString(other.SQL) {
		res.AddPropertyDiff("SQL")
	}

	if p.Width == nil || other.Width == nil {
		if !(p.Width == nil && other.Width == nil) {
			res.AddPropertyDiff("Width")
		}
	} else if *p.Width != *other.Width {
		res.AddPropertyDiff("Width")
	}

	for k, v := range p.Properties {
		if other.Properties[k] != v {
			res.AddPropertyDiff(fmt.Sprintf("Properties.%s", k))
		}
	}

	if typehelpers.SafeString(p.Type) != typehelpers.SafeString(other.Type) {
		res.AddPropertyDiff("Type")
	}

	res.populateChildDiffs(p, other)

	return res
}
