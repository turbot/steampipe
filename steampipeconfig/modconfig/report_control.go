package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/utils"
	"github.com/zclconf/go-cty/cty"
)

// ReportControl is a struct representing a leaf reporting node
type ReportControl struct {
	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	// these properties are JSON serialised by the parent LeafRun
	Title *string  `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width *int     `cty:"width" hcl:"width" column:"width,text"  json:"-"`
	Base  *Control `hcl:"base" json:"-"`
	SQL   *string  `cty:"sql" hcl:"sql" json:"-"`

	DeclRange hcl.Range  `json:"-"`
	Mod       *Mod       `cty:"mod" json:"-"`
	Paths     []NodePath `column:"path,jsonb" json:"-"`

	// the underlying control
	control  *Control
	parents  []ModTreeItem
	metadata *ResourceMetadata
}

func NewReportControl(block *hcl.Block, control *Control) *ReportControl {
	return &ReportControl{
		DeclRange:       block.DefRange,
		ShortName:       block.Labels[0],
		FullName:        fmt.Sprintf("%s.%s", block.Type, block.Labels[0]),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, block.Labels[0]),
		control:         control,
		SQL:             control.SQL,
	}
}

func (c *ReportControl) Equals(other *ReportControl) bool {
	diff := c.Diff(other)
	return !diff.HasChanges()
}

func (c *ReportControl) SetControl(control *Control) {
	c.control = control
}

// CtyValue implements HclResource
func (c *ReportControl) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'counter.<shortName>'
func (c *ReportControl) Name() string {
	return c.FullName
}

// OnDecoded implements HclResource
func (c *ReportControl) OnDecoded(*hcl.Block) hcl.Diagnostics {
	c.setBaseProperties()

	return nil
}

func (c *ReportControl) setBaseProperties() {
	if c.Base == nil {
		return
	}
	// pull title up
	if c.Title == nil {
		c.Title = c.Base.Title
	}
	if c.SQL == nil {
		c.SQL = c.Base.SQL
	}

	// now merge the control properties
	c.control.Merge(c.Base)
}

// AddReference implements HclResource
func (c *ReportControl) AddReference(*ResourceReference) {}

// SetMod implements HclResource
func (c *ReportControl) SetMod(mod *Mod) {
	c.Mod = mod
	// set mod for our underlying control
	c.control.SetMod(mod)
	c.FullName = fmt.Sprintf("%s.%s", c.Mod.ShortName, c.UnqualifiedName)
}

// GetMod implements HclResource
func (c *ReportControl) GetMod() *Mod {
	return c.Mod
}

// GetDeclRange implements HclResource
func (c *ReportControl) GetDeclRange() *hcl.Range {
	return &c.DeclRange
}

// AddParent implements ModTreeItem
func (c *ReportControl) AddParent(parent ModTreeItem) error {
	c.parents = append(c.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (c *ReportControl) GetParents() []ModTreeItem {
	return c.parents
}

// GetChildren implements ModTreeItem
func (c *ReportControl) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements ModTreeItem
func (c *ReportControl) GetTitle() string {
	return typehelpers.SafeString(c.Title)
}

// GetDescription implements ModTreeItem
func (c *ReportControl) GetDescription() string {
	return typehelpers.SafeString(c.control.Description)
}

// GetTags implements ModTreeItem
func (c *ReportControl) GetTags() map[string]string {
	return c.control.Tags
}

// GetPaths implements ModTreeItem
func (c *ReportControl) GetPaths() []NodePath {
	// lazy load
	if len(c.Paths) == 0 {
		c.SetPaths()
	}

	return c.Paths
}

// SetPaths implements ModTreeItem
func (c *ReportControl) SetPaths() {
	for _, parent := range c.parents {
		for _, parentPath := range parent.GetPaths() {
			c.Paths = append(c.Paths, append(parentPath, c.Name()))
		}
	}
}

// GetMetadata implements ResourceWithMetadata
func (c *ReportControl) GetMetadata() *ResourceMetadata {
	return c.metadata
}

// SetMetadata implements ResourceWithMetadata
func (c *ReportControl) SetMetadata(metadata *ResourceMetadata) {
	c.metadata = metadata
}

func (c *ReportControl) Diff(other *ReportControl) *ReportTreeItemDiffs {
	res := &ReportTreeItemDiffs{
		Item: c,
		Name: c.Name(),
	}
	if !utils.SafeStringsEqual(c.FullName, other.FullName) {
		res.AddPropertyDiff("Name")
	}

	if !utils.SafeStringsEqual(c.Title, other.Title) {
		res.AddPropertyDiff("Title")
	}

	if !utils.SafeStringsEqual(c.SQL, other.SQL) {
		res.AddPropertyDiff("Title")
	}

	if !utils.SafeIntEqual(c.Width, other.Width) {
		res.AddPropertyDiff("Width")
	}

	if !c.control.Equals(other.control) {
		res.AddPropertyDiff("control")
	}

	res.populateChildDiffs(c, other)

	return res
}

// GetSQL implements ReportLeafNode
func (c *ReportControl) GetSQL() string {
	return typehelpers.SafeString(c.SQL)
}

// GetWidth implements ReportLeafNode
func (c *ReportControl) GetWidth() int {
	if c.Width == nil {
		return 0
	}
	return *c.Width
}

// GetUnqualifiedName implements ReportLeafNode
func (c *ReportControl) GetUnqualifiedName() string {
	return c.UnqualifiedName
}

func (c *ReportControl) GetControl() *Control {
	return c.control
}
