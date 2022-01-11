package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// ReportContainer is a struct representing the Report and Container resource
type ReportContainer struct {
	ShortName       string
	FullName        string `cty:"name"`
	UnqualifiedName string

	ChildNames       []NamedItem `cty:"children" `
	ChildNameStrings []string    `column:"children,jsonb"`

	Title *string `cty:"title" column:"title,text"`
	Width *int    `cty:"width"  column:"width,text"`

	Mod       *Mod `cty:"mod"`
	DeclRange hcl.Range
	Paths     []NodePath `column:"path,jsonb"`

	parents  []ModTreeItem
	children []ModTreeItem
	metadata *ResourceMetadata

	hclType string
}

func NewReportContainer(block *hcl.Block) *ReportContainer {
	report := &ReportContainer{
		ShortName:       block.Labels[0],
		FullName:        fmt.Sprintf("report.%s", block.Labels[0]),
		UnqualifiedName: fmt.Sprintf("report.%s", block.Labels[0]),
		DeclRange:       block.DefRange,
		hclType:         block.Type,
	}
	return report
}

// CtyValue implements HclResource
func (r *ReportContainer) CtyValue() (cty.Value, error) {
	return getCtyValue(r)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'report.<shortName>'
func (r *ReportContainer) Name() string {
	return r.FullName
}

// OnDecoded implements HclResource
func (r *ReportContainer) OnDecoded(block *hcl.Block) hcl.Diagnostics {
	var res hcl.Diagnostics
	if len(r.ChildNames) == 0 {
		return nil
	}

	// validate each child name appears only once
	nameMap := make(map[string]bool)
	r.ChildNameStrings = make([]string, len(r.ChildNames))
	for i, n := range r.ChildNames {
		if nameMap[n.Name] {
			res = append(res, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("container '%s' has duplicate child name '%s'", r.FullName, n.Name),
				Subject:  &block.DefRange})

			continue
		}
		r.ChildNameStrings[i] = n.Name
		nameMap[n.Name] = true
	}

	// in order to populate the children in the order specified, we create an empty array and populate by index in AddChild
	r.children = make([]ModTreeItem, len(r.ChildNameStrings))
	return res
}

// AddReference implements HclResource
func (r *ReportContainer) AddReference(*ResourceReference) {
	// TODO
}

// SetMod implements HclResource
func (r *ReportContainer) SetMod(mod *Mod) {
	r.Mod = mod
	r.UnqualifiedName = r.FullName
	r.FullName = fmt.Sprintf("%s.%s", mod.ShortName, r.FullName)
}

// GetMod implements HclResource
func (r *ReportContainer) GetMod() *Mod {
	return r.Mod
}

// GetDeclRange implements HclResource
func (r *ReportContainer) GetDeclRange() *hcl.Range {
	return &r.DeclRange
}

// AddChild implements ModTreeItem
// this ic called from mod.addItemIntoResourceTree
func (r *ReportContainer) AddChild(child ModTreeItem) error {
	// if children are declared inline (as opposed to via the 'children' property) they will already have been added
	if len(r.ChildNames) == 0 {
		return nil
	}

	// so a children property must have been populated

	// now find which position this child is in the array
	for i, name := range r.ChildNameStrings {
		if name == child.Name() {
			r.children[i] = child
			return nil
		}
	}

	return fmt.Errorf("container '%s' has no child '%s'", r.Name(), child.Name())
}

// AddParent implements ModTreeItem
func (r *ReportContainer) AddParent(parent ModTreeItem) error {
	r.parents = append(r.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (r *ReportContainer) GetParents() []ModTreeItem {
	return r.parents
}

// GetChildren implements ModTreeItem
func (r *ReportContainer) GetChildren() []ModTreeItem {
	return r.children
}

// SetChildren is used to set the children when they are declared inline
func (r *ReportContainer) SetChildren(children []ModTreeItem) {
	r.ChildNameStrings = make([]string, len(children))
	for i, c := range children {
		r.ChildNameStrings[i] = c.Name()
	}
	r.children = children
}

// GetTitle implements ModTreeItem
func (r *ReportContainer) GetTitle() string {
	return typehelpers.SafeString(r.Title)
}

// GetDescription implements ModTreeItem
func (r *ReportContainer) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (r *ReportContainer) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (r *ReportContainer) GetPaths() []NodePath {
	// lazy load
	if len(r.Paths) == 0 {
		r.SetPaths()
	}
	return r.Paths
}

// SetPaths implements ModTreeItem
func (r *ReportContainer) SetPaths() {
	for _, parent := range r.parents {
		for _, parentPath := range parent.GetPaths() {
			r.Paths = append(r.Paths, append(parentPath, r.Name()))
		}
	}
}

// GetMetadata implements ResourceWithMetadata
func (r *ReportContainer) GetMetadata() *ResourceMetadata {
	return r.metadata
}

// SetMetadata implements ResourceWithMetadata
func (r *ReportContainer) SetMetadata(metadata *ResourceMetadata) {
	r.metadata = metadata
}

func (r *ReportContainer) Diff(new *ReportContainer) *ReportTreeItemDiffs {
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

func (r *ReportContainer) IsReport() bool {
	return r.hclType == "report"
}
