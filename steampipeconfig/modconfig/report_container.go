package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/utils"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// ReportContainer is a struct representing the Report and Container resource
type ReportContainer struct {
	ShortName       string
	FullName        string `cty:"name"`
	UnqualifiedName string

	Title *string `cty:"title" column:"title,text"`
	Width *int    `cty:"width"  column:"width,text"`

	Mod       *Mod `cty:"mod"`
	DeclRange hcl.Range

	Base  *ReportContainer
	Paths []NodePath `column:"path,jsonb"`
	// store children in a way which can be serialised via cty
	ChildNames []string `cty:"children"`

	// the actual children
	children []ModTreeItem
	parents  []ModTreeItem
	metadata *ResourceMetadata

	HclType   string
	anonymous bool
}

func NewReportContainer(block *hcl.Block) *ReportContainer {
	return &ReportContainer{
		DeclRange:       block.DefRange,
		HclType:         block.Type,
		ShortName:       block.Labels[0],
		FullName:        fmt.Sprintf("%s.%s", block.Type, block.Labels[0]),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, block.Labels[0]),
	}
}

func (c *ReportContainer) Equals(other *ReportContainer) bool {
	diff := c.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (c *ReportContainer) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'report.<shortName>'
func (c *ReportContainer) Name() string {
	return c.FullName
}

func (c *ReportContainer) SetAnonymous(anonymous bool) {
	c.anonymous = anonymous
}

func (c *ReportContainer) IsAnonymous() bool {
	return c.anonymous
}

// OnDecoded implements HclResource
func (c *ReportContainer) OnDecoded(block *hcl.Block) hcl.Diagnostics {
	c.setBaseProperties()
	return nil
}

func (c *ReportContainer) setBaseProperties() {
	if c.Base == nil {
		return
	}
	if c.Title == nil {
		c.Title = c.Base.Title
	}
	if c.Width == nil {
		c.Width = c.Base.Width
	}
	if len(c.children) == 0 {
		c.children = c.Base.GetChildren()
		c.ChildNames = c.Base.ChildNames
	}
}

// AddReference implements HclResource
func (c *ReportContainer) AddReference(*ResourceReference) {
	// TODO
}

// SetMod implements HclResource
func (c *ReportContainer) SetMod(mod *Mod) {
	c.Mod = mod

	// if this is a top level resource, and not a child, the resource names will already be set
	// - we need to update the full name to include the mod
	if c.UnqualifiedName != "" {
		// add mod name to full name
		c.FullName = fmt.Sprintf("%s.%s", mod.ShortName, c.UnqualifiedName)
	}
}

// GetMod implements HclResource
func (c *ReportContainer) GetMod() *Mod {
	return c.Mod
}

// GetDeclRange implements HclResource
func (c *ReportContainer) GetDeclRange() *hcl.Range {
	return &c.DeclRange
}

// AddParent implements ModTreeItem
func (c *ReportContainer) AddParent(parent ModTreeItem) error {
	c.parents = append(c.parents, parent)

	return nil
}

// GetParents implements ModTreeItem
func (c *ReportContainer) GetParents() []ModTreeItem {
	return c.parents
}

// GetChildren implements ModTreeItem
func (c *ReportContainer) GetChildren() []ModTreeItem {
	return c.children
}

// GetTitle implements ModTreeItem
func (c *ReportContainer) GetTitle() string {
	return typehelpers.SafeString(c.Title)
}

// GetDescription implements ModTreeItem
func (c *ReportContainer) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (c *ReportContainer) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (c *ReportContainer) GetPaths() []NodePath {
	// lazy load
	if len(c.Paths) == 0 {
		c.SetPaths()
	}
	return c.Paths
}

// SetPaths implements ModTreeItem
func (c *ReportContainer) SetPaths() {
	for _, parent := range c.parents {
		for _, parentPath := range parent.GetPaths() {
			c.Paths = append(c.Paths, append(parentPath, c.Name()))
		}
	}
}

// GetMetadata implements ResourceWithMetadata
func (c *ReportContainer) GetMetadata() *ResourceMetadata {
	return c.metadata
}

// SetMetadata implements ResourceWithMetadata
func (c *ReportContainer) SetMetadata(metadata *ResourceMetadata) {
	c.metadata = metadata
}

func (c *ReportContainer) Diff(other *ReportContainer) *ReportTreeItemDiffs {
	res := &ReportTreeItemDiffs{
		Item: c,
		Name: c.Name(),
	}

	if utils.SafeStringsEqual(c.FullName, other.FullName) {
		res.AddPropertyDiff("Name")
	}

	if utils.SafeStringsEqual(c.Title, other.Title) {
		res.AddPropertyDiff("Title")
	}

	if utils.SafeIntEqual(c.Width, other.Width) {
		res.AddPropertyDiff("Width")
	}

	res.populateChildDiffs(c, other)
	return res
}

func (c *ReportContainer) IsReport() bool {
	return c.HclType == "report"
}

func (c *ReportContainer) SetChildren(children []ModTreeItem) {
	c.children = children
	c.ChildNames = make([]string, len(children))
	for i, child := range children {
		c.ChildNames[i] = child.Name()
	}
}

// GetUnqualifiedName implements ReportLeafNode
func (c *ReportContainer) GetUnqualifiedName() string {
	return c.UnqualifiedName
}
