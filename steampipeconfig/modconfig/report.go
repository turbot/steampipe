package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// Report is a struct representing the Report resource
type Report struct {
	FullName  string `cty:"name"`
	ShortName string
	Title     *string

	Reports []*Report //`hcl:"report,block"`
	Panels  []*Panel  //`hcl:"panel,block"`

	Mod *Mod `cty:"mod"`

	DeclRange hcl.Range

	parents         []ModTreeItem
	metadata        *ResourceMetadata
	UnqualifiedName string
}

func NewReport(block *hcl.Block) *Report {
	report := &Report{
		ShortName:       block.Labels[0],
		FullName:        fmt.Sprintf("report.%s", block.Labels[0]),
		UnqualifiedName: fmt.Sprintf("report.%s", block.Labels[0]),
		DeclRange:       block.DefRange,
	}
	return report
}

// CtyValue implements HclResource
func (r *Report) CtyValue() (cty.Value, error) {
	return getCtyValue(r)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'report.<shortName>'
func (r *Report) Name() string {
	return r.FullName
}

// OnDecoded implements HclResource
func (r *Report) OnDecoded(*hcl.Block) hcl.Diagnostics { return nil }

// AddReference implements HclResource
func (r *Report) AddReference(*ResourceReference) {
	// TODO
}

// SetMod implements HclResource
func (r *Report) SetMod(mod *Mod) {
	r.Mod = mod
	r.UnqualifiedName = r.FullName
	r.FullName = fmt.Sprintf("%s.%s", mod.ShortName, r.FullName)
}

// GetMod implements HclResource
func (r *Report) GetMod() *Mod {
	return r.Mod
}

// GetDeclRange implements HclResource
func (r *Report) GetDeclRange() *hcl.Range {
	return &r.DeclRange
}

// AddChild implements ModTreeItem
func (r *Report) AddChild(child ModTreeItem) error {
	switch c := child.(type) {
	case *Panel:
		// avoid duplicates
		if !r.containsPanel(c.Name()) {
			r.Panels = append(r.Panels, c)
		}
	case *Report:
		// avoid duplicates
		if !r.containsReport(c.Name()) {
			r.Reports = append(r.Reports, c)
		}
	}
	return nil
}

// AddParent implements ModTreeItem
func (r *Report) AddParent(parent ModTreeItem) error {
	r.parents = append(r.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (r *Report) GetParents() []ModTreeItem {
	return r.parents
}

// GetChildren implements ModTreeItem
func (r *Report) GetChildren() []ModTreeItem {
	children := make([]ModTreeItem, len(r.Panels)+len(r.Reports))
	idx := 0
	for _, p := range r.Panels {
		children[idx] = p
		idx++
	}
	for _, r := range r.Reports {
		children[idx] = r
		idx++
	}
	return children
}

// GetTitle implements ModTreeItem
func (r *Report) GetTitle() string {
	return typehelpers.SafeString(r.Title)
}

// GetDescription implements ModTreeItem
func (r *Report) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (r *Report) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (r *Report) GetPaths() []NodePath {
	var res []NodePath
	for _, parent := range r.parents {
		for _, parentPath := range parent.GetPaths() {
			res = append(res, append(parentPath, r.Name()))
		}
	}
	return res
}

// GetMetadata implements ResourceWithMetadata
func (r *Report) GetMetadata() *ResourceMetadata {
	return r.metadata
}

// SetMetadata implements ResourceWithMetadata
func (r *Report) SetMetadata(metadata *ResourceMetadata) {
	r.metadata = metadata
}

func (r *Report) Diff(new *Report) *ReportTreeItemDiffs {
	res := &ReportTreeItemDiffs{
		Item: r,
		Name: r.Name(),
	}

	if typehelpers.SafeString(r.Title) != typehelpers.SafeString(new.Title) {
		res.AddPropertyDiff("Title")
	}

	res.populateChildDiffs(r, new)
	return res
}

func (r *Report) containsPanel(name string) bool {
	// does this child already exist
	for _, existingPanel := range r.Panels {
		if existingPanel.Name() == name {
			return true
		}
	}
	return false
}

func (r *Report) containsReport(name string) bool {
	// does this child already exist
	for _, existingReport := range r.Reports {
		if existingReport.Name() == name {
			return true
		}
	}
	return false
}
