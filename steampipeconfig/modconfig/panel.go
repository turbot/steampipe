package modconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
	"github.com/zclconf/go-cty/cty"
)

// Panel is a struct representing the Report resource
type Panel struct {
	FullName  string `cty:"name"`
	ShortName string

	Title  *string `cty:"title" column:"title,text"`
	Type   *string `cty:"type" column:"type,text"`
	Width  *int    `cty:"width" column:"width,text"`
	Height *int    `cty:"height" column:"height,text"`
	Source *string `cty:"source" column:"source,text"`
	SQL    *string `cty:"sql" column:"sql,text"`
	Text   *string `cty:"text" column:"text,text"`

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

// PanelFromFile creates a panel from a markdown file
func PanelFromFile(modPath, filePath string) (MappableResource, []byte, error) {
	p := &Panel{}
	return p.InitialiseFromFile(modPath, filePath)
}

// InitialiseFromFile implements MappableResource
func (p *Panel) InitialiseFromFile(modPath, filePath string) (MappableResource, []byte, error) {
	// only valid for sql files
	if filepath.Ext(filePath) != constants.MarkdownExtension {
		return nil, nil, fmt.Errorf("Panel.InitialiseFromFile must be called with markdown files only - filepath: '%s'", filePath)
	}

	markdownBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	markdown := string(markdownBytes)

	// get a sluggified version of the filename
	name, err := PseudoResourceNameFromPath(modPath, filePath)
	if err != nil {
		return nil, nil, err
	}
	p.ShortName = name
	p.FullName = fmt.Sprintf("panel.%s", name)
	p.Text = &markdown
	p.Source = utils.ToStringPointer("steampipe.panel.markdown")
	return p, markdownBytes, nil
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
	if p.Height == nil {
		p.Height = p.Base.Height
	}
	if p.Source == nil {
		p.Source = p.Base.Source
	}
	if p.SQL == nil {
		p.SQL = p.Base.SQL
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
	if typehelpers.SafeString(p.Source) != typehelpers.SafeString(new.Source) {
		res.AddPropertyDiff("Source")
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
	if p.Height == nil || new.Height == nil {
		if !(p.Height == nil && new.Height == nil) {
			res.AddPropertyDiff("Height")
		}
	} else if *p.Height != *new.Height {
		res.AddPropertyDiff("Height")
	}

	res.populateChildDiffs(p, new)

	return res
}
