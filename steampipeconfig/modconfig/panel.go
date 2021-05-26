package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
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

	parents  []ControlTreeItem
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

// AddChild implements ControlTreeItem
func (r *Panel) AddChild(child ControlTreeItem) error {
	switch c := child.(type) {
	case *Panel:
		r.Panels = append(r.Panels, c)
	case *Report:
		r.Reports = append(r.Reports, c)
	}
	return nil
}

// AddParent implements ControlTreeItem
func (c *Panel) AddParent(parent ControlTreeItem) error {
	c.parents = append(c.parents, parent)
	return nil
}

// GetParents implements ControlTreeItem
func (c *Panel) GetParents() []ControlTreeItem {
	return c.parents
}

// GetChildren implements ControlTreeItem
func (c *Panel) GetChildren() []ControlTreeItem {
	children := make([]ControlTreeItem, len(c.Panels)+len(c.Reports))
	idx := 0
	for _, p := range c.Panels {
		children[idx] = p
		idx++
	}
	for _, r := range c.Reports {
		children[idx] = r
		idx++
	}
	return children
}

// GetTitle implements ControlTreeItem
func (c *Panel) GetTitle() string {
	return typehelpers.SafeString(c.Title)
}

// GetDescription implements ControlTreeItem
func (c *Panel) GetDescription() string {
	return ""
}

// GetTags implements ControlTreeItem
func (c *Panel) GetTags() map[string]string {
	return nil
}

// Path implements ControlTreeItem
func (c *Panel) Path() []string {
	// TODO update for multiple paths
	path := []string{c.FullName}
	if c.parents != nil {
		path = append(c.parents[0].Path(), path...)
	}
	return path
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
