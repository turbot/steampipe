package modconfig

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
)

// Report is a struct representing the Report resource
type Report struct {
	FullName  string `cty:"name"`
	ShortName string `json:"short_name"`

	Reports []*Report //`hcl:"report,block"`
	Panels  []*Panel  //`hcl:"panel,block"`

	DeclRange hcl.Range
}

func NewReport(block *hcl.Block) *Report {
	report := &Report{
		ShortName: block.Labels[0],
		FullName:  fmt.Sprintf("report.%s", block.Labels[0]),
		DeclRange: block.DefRange,
	}
	return report
}

// CtyValue implements HclResource
func (r *Report) CtyValue() (cty.Value, error) {
	return getCtyValue(r)
}

// Name implements HclResource
// return name in format: 'panel.<shortName>'
func (r *Report) Name() string {
	return r.FullName
}

// GetMetadata implements HclResource
func (r *Report) GetMetadata() *ResourceMetadata {
	// TODO
	return nil
}

// OnDecoded implements HclResource
func (r *Report) OnDecoded(*hcl.Block) {}

// AddReference implements HclResource
func (r *Report) AddReference(reference string) {
	// TODO
}

// SetPanels implements ReportTreeItem
func (r *Report) SetPanels(panels []*Panel) {
	r.Panels = panels
}

// SetReports implements ReportTreeItem
func (r *Report) SetReports(reports []*Report) {
	r.Reports = reports
}
