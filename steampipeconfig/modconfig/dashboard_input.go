package modconfig

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/utils"
	"github.com/zclconf/go-cty/cty"
)

// DashboardInput is a struct representing a leaf dashboard node
type DashboardInput struct {
	DashboardLeafNodeBase
	ResourceWithMetadataBase

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `cty:"unqualified_name" json:"-"`

	// these properties are JSON serialised by the parent LeafRun
	Title *string         `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width *int            `cty:"width" hcl:"width" column:"width,text"  json:"-"`
	SQL   *string         `cty:"sql" hcl:"sql" column:"sql,text" json:"sql"`
	Type  *string         `cty:"type" hcl:"type" column:"type,text"  json:"type,omitempty"`
	Style *string         `cty:"style" hcl:"style" column:"style,text" json:"style,omitempty"`
	Value *string         `json:"value"`
	Base  *DashboardInput `hcl:"base" json:"-"`

	DeclRange hcl.Range  `json:"-"`
	Mod       *Mod       `cty:"mod" json:"-"`
	Paths     []NodePath `column:"path,jsonb" json:"-"`

	parents            []ModTreeItem
	dashboardContainer *DashboardContainer
}

func NewDashboardInput(block *hcl.Block, mod *Mod) *DashboardInput {
	shortName := block.Labels[0]
	i := &DashboardInput{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}
	return i
}

func (c *DashboardInput) Equals(other *DashboardInput) bool {
	diff := c.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (c *DashboardInput) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'chart.<shortName>'
func (c *DashboardInput) Name() string {
	return c.FullName
}

// OnDecoded implements HclResource
func (c *DashboardInput) OnDecoded(*hcl.Block) hcl.Diagnostics {
	c.setBaseProperties()
	return nil
}

func (c *DashboardInput) setBaseProperties() {
	if c.Base == nil {
		return
	}
	if c.Title == nil {
		c.Title = c.Base.Title
	}
	if c.Type == nil {
		c.Type = c.Base.Type
	}

	if c.Width == nil {
		c.Width = c.Base.Width
	}
	if c.SQL == nil {
		c.SQL = c.Base.SQL
	}
}

// AddReference implements HclResource
func (c *DashboardInput) AddReference(*ResourceReference) {}

// GetMod implements HclResource
func (c *DashboardInput) GetMod() *Mod {
	return c.Mod
}

// GetDeclRange implements HclResource
func (c *DashboardInput) GetDeclRange() *hcl.Range {
	return &c.DeclRange
}

// AddParent implements ModTreeItem
func (c *DashboardInput) AddParent(parent ModTreeItem) error {
	c.parents = append(c.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (c *DashboardInput) GetParents() []ModTreeItem {
	return c.parents
}

// GetChildren implements ModTreeItem
func (c *DashboardInput) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements ModTreeItem
func (c *DashboardInput) GetTitle() string {
	return typehelpers.SafeString(c.Title)
}

// GetDescription implements ModTreeItem
func (c *DashboardInput) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (c *DashboardInput) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (c *DashboardInput) GetPaths() []NodePath {
	// lazy load
	if len(c.Paths) == 0 {
		c.SetPaths()
	}

	return c.Paths
}

// SetPaths implements ModTreeItem
func (c *DashboardInput) SetPaths() {
	for _, parent := range c.parents {
		for _, parentPath := range parent.GetPaths() {
			c.Paths = append(c.Paths, append(parentPath, c.Name()))
		}
	}
}

func (c *DashboardInput) Diff(other *DashboardInput) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
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
		res.AddPropertyDiff("SQL")
	}

	if !utils.SafeIntEqual(c.Width, other.Width) {
		res.AddPropertyDiff("Width")
	}

	if !utils.SafeStringsEqual(c.Type, other.Type) {
		res.AddPropertyDiff("Type")
	}

	res.populateChildDiffs(c, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (c *DashboardInput) GetWidth() int {
	if c.Width == nil {
		return 0
	}
	return *c.Width
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (c *DashboardInput) GetUnqualifiedName() string {
	return c.UnqualifiedName
}

// SetDashboardContainer sets the parent dashboard container
func (c *DashboardInput) SetDashboardContainer(dashboardContainer *DashboardContainer) {
	// TODO [reports] also update unqualified name?
	c.dashboardContainer = dashboardContainer
	// update the full name with a sanitsed version of the parent dashboard name
	dashboardName := strings.Replace(c.dashboardContainer.UnqualifiedName, ".", "_", -1)
	c.FullName = fmt.Sprintf("%s.%s.%s", c.Mod.ShortName, dashboardName, c.UnqualifiedName)
	c.UnqualifiedName = fmt.Sprintf("%s.%s", dashboardName, c.UnqualifiedName)
}
