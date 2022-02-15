package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/utils"
	"github.com/zclconf/go-cty/cty"
)

// DashboardImage is a struct representing a leaf reporting node
type DashboardImage struct {
	ReportLeafNodeBase
	ResourceWithMetadataBase

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	// these properties are JSON serialised by the parent LeafRun
	Title *string         `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width *int            `cty:"width" hcl:"width" column:"width,text"  json:"-"`
	SQL   *string      `cty:"sql" hcl:"sql" column:"sql,text" json:"-"`
	Src   *string      `cty:"src" hcl:"src" column:"src,text"  json:"src,omitempty"`
	Alt   *string         `cty:"alt" hcl:"alt" column:"alt,text"  json:"alt,omitempty"`
	Base  *DashboardImage `hcl:"base" json:"-"`

	DeclRange hcl.Range  `json:"-"`
	Mod       *Mod       `cty:"mod" json:"-"`
	Paths     []NodePath `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewReportImage(block *hcl.Block, mod *Mod) *DashboardImage {
	shortName := GetAnonymousResourceShortName(block, mod)
	i := &DashboardImage{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}
	i.SetAnonymous(block)
	return i
}

func (c *DashboardImage) Equals(other *DashboardImage) bool {
	diff := c.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (c *DashboardImage) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'image.<shortName>'
func (c *DashboardImage) Name() string {
	return c.FullName
}

// OnDecoded implements HclResource
func (c *DashboardImage) OnDecoded(*hcl.Block) hcl.Diagnostics {
	c.setBaseProperties()
	return nil
}

func (c *DashboardImage) setBaseProperties() {
	if c.Base == nil {
		return
	}
	if c.Title == nil {
		c.Title = c.Base.Title
	}
	if c.Src == nil {
		c.Src = c.Base.Src
	}
	if c.Alt == nil {
		c.Alt = c.Base.Alt
	}
	if c.Width == nil {
		c.Width = c.Base.Width
	}
	if c.SQL == nil {
		c.SQL = c.Base.SQL
	}
}

// AddReference implements HclResource
func (c *DashboardImage) AddReference(*ResourceReference) {}

// GetMod implements HclResource
func (c *DashboardImage) GetMod() *Mod {
	return c.Mod
}

// GetDeclRange implements HclResource
func (c *DashboardImage) GetDeclRange() *hcl.Range {
	return &c.DeclRange
}

// AddParent implements ModTreeItem
func (c *DashboardImage) AddParent(parent ModTreeItem) error {
	c.parents = append(c.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (c *DashboardImage) GetParents() []ModTreeItem {
	return c.parents
}

// GetChildren implements ModTreeItem
func (c *DashboardImage) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements ModTreeItem
func (c *DashboardImage) GetTitle() string {
	return typehelpers.SafeString(c.Title)
}

// GetDescription implements ModTreeItem
func (c *DashboardImage) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (c *DashboardImage) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (c *DashboardImage) GetPaths() []NodePath {
	// lazy load
	if len(c.Paths) == 0 {
		c.SetPaths()
	}

	return c.Paths
}

// SetPaths implements ModTreeItem
func (c *DashboardImage) SetPaths() {
	for _, parent := range c.parents {
		for _, parentPath := range parent.GetPaths() {
			c.Paths = append(c.Paths, append(parentPath, c.Name()))
		}
	}
}

func (c *DashboardImage) Diff(other *DashboardImage) *DashboardTreeItemDiffs {
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

	if !utils.SafeStringsEqual(c.Src, other.Src) {
		res.AddPropertyDiff("Src")
	}

	if !utils.SafeStringsEqual(c.Alt, other.Alt) {
		res.AddPropertyDiff("Alt")
	}

	res.populateChildDiffs(c, other)

	return res
}

// GetSQL implements DashboardLeafNode
func (c *DashboardImage) GetSQL() string {
	return typehelpers.SafeString(c.SQL)
}

// GetWidth implements DashboardLeafNode
func (c *DashboardImage) GetWidth() int {
	if c.Width == nil {
		return 0
	}
	return *c.Width
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (c *DashboardImage) GetUnqualifiedName() string {
	return c.UnqualifiedName
}
