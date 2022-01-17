package modconfig

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// ReportContainer is a struct representing the Report and Container resource
type ReportContainer struct {
	ShortName       string
	FullName        string `cty:"name"`
	UnqualifiedName string

	// used for introspection tables
	ChildNameStrings []string `cty:"children" column:"children,jsonb"`

	Title *string `cty:"title" column:"title,text"`
	Width *int    `cty:"width"  column:"width,text"`

	Mod       *Mod `cty:"mod"`
	DeclRange hcl.Range

	Base  *ReportContainer
	Paths []NodePath `column:"path,jsonb"`

	// the actual children
	children []ModTreeItem
	parents  []ModTreeItem
	metadata *ResourceMetadata

	hclType string
}

func NewReportContainer(block *hcl.Block) *ReportContainer {
	report := &ReportContainer{
		DeclRange: block.DefRange,
		hclType:   block.Type,
	}
	// report name is defined in hcl
	if report.IsReport() {
		report.ShortName = block.Labels[0]
		// mod name is added later
		report.FullName = fmt.Sprintf("%s.%s", block.Type, block.Labels[0])
		report.UnqualifiedName = report.FullName
	}
	// if the block has a name, set the name
	anonymous := len(block.Labels) == 0
	if !anonymous {
		report.SetName(block.Labels[0])
	}

	return report
}

func (r *ReportContainer) Equals(other *ReportContainer) bool {
	diff := r.Diff(other)
	return !diff.HasChanges()
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
	r.setBaseProperties()
	return nil
}

func (r *ReportContainer) setBaseProperties() {
	if r.Base == nil {
		return
	}
	if r.Title == nil {
		r.Title = r.Base.Title
	}
	if r.Width == nil {
		r.Width = r.Base.Width
	}
}

// AddReference implements HclResource
func (r *ReportContainer) AddReference(*ResourceReference) {
	// TODO
}

// SetMod implements HclResource
func (r *ReportContainer) SetMod(mod *Mod) {
	r.Mod = mod

	// if this is a top level resource, and not a child, the resource names will already be set
	// - we need to update the full name to include the mod
	if r.UnqualifiedName != "" {
		// add mod name to full name
		r.FullName = fmt.Sprintf("%s.%s", mod.ShortName, r.UnqualifiedName)
	}
}

// GetMod implements HclResource
func (r *ReportContainer) GetMod() *Mod {
	return r.Mod
}

// GetDeclRange implements HclResource
func (r *ReportContainer) GetDeclRange() *hcl.Range {
	return &r.DeclRange
}

// SetName implements AnonymousResource
func (r *ReportContainer) SetName(name string) {
	// if name is already set, do nothing
	if r.ShortName != "" {
		return
	}

	r.ShortName = name
	r.UnqualifiedName = fmt.Sprintf("%s.%s", r.hclType, name)

	//  if this is a child resource, the mod name and metadata will already have been set
	// (because of the decode order)
	// we need top set FullName to inmclude the mof name, and update the resource name in the metadata
	if r.Mod != nil {
		// set the full name
		r.FullName = fmt.Sprintf("%s.%s", r.Mod.ShortName, r.UnqualifiedName)
		// update the name in metadata
		r.metadata.ResourceName = r.ShortName

	} else {
		// if mod is not yet set, just set the full name to the unqualified name for now
		r.FullName = r.UnqualifiedName
	}

	// now set child name
	r.setChildNames()

}

// HclType implements AnonymousResource
func (r *ReportContainer) HclType() string {
	return r.hclType
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

func (r *ReportContainer) Diff(other *ReportContainer) *ReportTreeItemDiffs {
	res := &ReportTreeItemDiffs{
		Item: r,
		Name: r.Name(),
	}

	if r.FullName != other.FullName {
		res.AddPropertyDiff("Name")
	}

	if typehelpers.SafeString(r.Title) != typehelpers.SafeString(other.Title) {
		res.AddPropertyDiff("Title")
	}

	if r.Width == nil || other.Width == nil {
		if !(r.Width == nil && other.Width == nil) {
			res.AddPropertyDiff("Width")
		}
	} else if *r.Width != *other.Width {
		res.AddPropertyDiff("Width")
	}

	res.populateChildDiffs(r, other)
	return res
}

func (r *ReportContainer) IsReport() bool {
	return r.hclType == "report"
}

func (r *ReportContainer) SetChildren(children []ModTreeItem) {
	r.children = children

	// the children will be anonymous
	// if we have a name, set child names
	if r.ShortName != "" {
		r.setChildNames()
	}
}

func (r *ReportContainer) setChildNames() {
	// sanitise the parent (our) name
	parentName := strings.Replace(r.ShortName, ".", "_", -1)
	// build map so we can generate indexes for each child resource type
	childIndexes := make(map[string]int)
	r.ChildNameStrings = make([]string, len(r.children))
	for i, child := range r.children {
		// all children are anonymous
		anonymousChild, ok := child.(AnonymousResource)
		if !ok {
			panic("all children must support AnonymousResource")
		}
		// get the 0-based index for this child type
		hclType := anonymousChild.HclType()
		idx := childIndexes[hclType]
		childIndexes[hclType] = idx + 1

		// set the name
		name := fmt.Sprintf("%s_%s_%d", parentName, hclType, idx)
		anonymousChild.SetName(name)
		// also store the name
		r.ChildNameStrings[i] = child.Name()
	}
}
