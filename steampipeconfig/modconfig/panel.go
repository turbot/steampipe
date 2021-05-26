package modconfig

import (
	"fmt"
	"io/ioutil"
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

	Title   *string `hcl:"title"`
	Width   *int    `hcl:"width"`
	Source  *string `hcl:"source"`
	SQL     *string `hcl:"source"`
	Text    *string `hcl:"text"`
	Reports []*Report
	Panels  []*Panel

	DeclRange hcl.Range

	parents  []ModTreeItem
	metadata *ResourceMetadata
}

func NewPanel(block *hcl.Block) *Panel {
	panel := &Panel{
		ShortName: block.Labels[0],
		FullName:  fmt.Sprintf("panel.%s", block.Labels[0]),
		DeclRange: block.DefRange,
	}
	return panel
}

// PanelFromFile :: factory function
func PanelFromFile(modPath, filePath string) (MappableResource, []byte, error) {
	p := &Panel{}
	return p.InitialiseFromFile(modPath, filePath)
}

// InitialiseFromFile implements MappableResource
func (q *Panel) InitialiseFromFile(modPath, filePath string) (MappableResource, []byte, error) {
	// only valid for sql files
	if filepath.Ext(filePath) != constants.MarkdownExtension {
		return nil, nil, fmt.Errorf("Panel.InitialiseFromFile must be called with markdown files only - filepath: '%s'", filePath)
	}

	markdownBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	markdown := string(markdownBytes)

	// get a sluggified version of the filename
	name, err := PseudoResourceNameFromPath(modPath, filePath)
	if err != nil {
		return nil, nil, err
	}
	q.ShortName = name
	q.FullName = fmt.Sprintf("panel.%s", name)
	q.Text = &markdown
	q.Source = utils.ToStringPointer("steampipe.panel.markdown")
	return q, markdownBytes, nil
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

// QualifiedName returns the name in format: '<modName>.panel.<shortName>'
func (p *Panel) QualifiedName() string {
	return fmt.Sprintf("%s.%s", p.metadata.ModShortName, p.FullName)
}

// OnDecoded implements HclResource
func (p *Panel) OnDecoded(*hcl.Block) {}

// AddReference implements HclResource
func (p *Panel) AddReference(reference string) {
	// TODO
}

// AddChild implements ModTreeItem
func (p *Panel) AddChild(child ModTreeItem) error {
	switch c := child.(type) {
	case *Panel:
		p.Panels = append(p.Panels, c)
	case *Report:
		p.Reports = append(p.Reports, c)
	}
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
	children := make([]ModTreeItem, len(p.Panels)+len(p.Reports))
	idx := 0
	for _, p := range p.Panels {
		children[idx] = p
		idx++
	}
	for _, r := range p.Reports {
		children[idx] = r
		idx++
	}
	return children
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
	var res []NodePath
	for _, parent := range p.parents {
		for _, parentPath := range parent.GetPaths() {
			res = append(res, append(parentPath, p.Name()))
		}
	}
	return res
}

//// AddChild implements ReportTreeItem
//func (p *Panel) AddChild(child ReportTreeItem) {
//	switch c := child.(type) {
//	case *Panel:
//		p.Panels = append(p.Panels, c)
//	case *Report:
//		p.Reports = append(p.Reports, c)
//	}
//}
//

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
