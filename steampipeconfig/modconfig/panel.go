package modconfig

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
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

// GetMetadata implements HclResource
func (p *Panel) GetMetadata() *ResourceMetadata {
	// TODO
	return nil
}

// OnDecoded implements HclResource
func (p *Panel) OnDecoded(*hcl.Block) {}

// AddReference implements HclResource
func (p *Panel) AddReference(reference string) {
	// TODO
}

// AddChild implements ReportTreeItem
func (p *Panel) AddChild(child ReportTreeItem) {
	switch c := child.(type) {
	case *Panel:
		p.Panels = append(p.Panels, c)
	case *Report:
		p.Reports = append(p.Reports, c)
	}
}
