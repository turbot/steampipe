package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/utils"
	"github.com/zclconf/go-cty/cty"
)

// ReportImage is a struct representing a leaf reporting node
type ReportImage struct {
	HclResourceBase
	ResourceWithMetadataBase

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	// these properties are JSON serialised by the parent LeafRun
	Title *string      `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width *int         `cty:"width" hcl:"width" column:"width,text"  json:"-"`
	SQL   *string      `cty:"sql" hcl:"sql" column:"sql,text" json:"-"`
	Src   *string      `cty:"src" hcl:"src" column:"src,text"  json:"src,omitempty"`
	Alt   *string      `cty:"alt" hcl:"alt" column:"alt,text"  json:"alt,omitempty"`
	Base  *ReportImage `hcl:"base" json:"-"`

	DeclRange hcl.Range  `json:"-"`
	Mod       *Mod       `cty:"mod" json:"-"`
	Paths     []NodePath `column:"path,jsonb" json:"-"`

	parents []ModTreeItem
}

func NewReportImage(block *hcl.Block, mod *Mod) *ReportImage {
	shortName := GetAnonymousResourceShortName(block, mod)
	i := &ReportImage{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}
	i.SetAnonymous(block)
	return i
}

func (c *ReportImage) Equals(other *ReportImage) bool {
	diff := c.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (c *ReportImage) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'image.<shortName>'
func (c *ReportImage) Name() string {
	return c.FullName
}

// OnDecoded implements HclResource
func (c *ReportImage) OnDecoded(*hcl.Block) hcl.Diagnostics {
	c.setBaseProperties()
	return nil
}

func (c *ReportImage) setBaseProperties() {
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
func (c *ReportImage) AddReference(*ResourceReference) {}

// GetMod implements HclResource
func (c *ReportImage) GetMod() *Mod {
	return c.Mod
}

// GetDeclRange implements HclResource
func (c *ReportImage) GetDeclRange() *hcl.Range {
	return &c.DeclRange
}

// AddParent implements ModTreeItem
func (c *ReportImage) AddParent(parent ModTreeItem) error {
	c.parents = append(c.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (c *ReportImage) GetParents() []ModTreeItem {
	return c.parents
}

// GetChildren implements ModTreeItem
func (c *ReportImage) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements ModTreeItem
func (c *ReportImage) GetTitle() string {
	return typehelpers.SafeString(c.Title)
}

// GetDescription implements ModTreeItem
func (c *ReportImage) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (c *ReportImage) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (c *ReportImage) GetPaths() []NodePath {
	// lazy load
	if len(c.Paths) == 0 {
		c.SetPaths()
	}

	return c.Paths
}

// SetPaths implements ModTreeItem
func (c *ReportImage) SetPaths() {
	for _, parent := range c.parents {
		for _, parentPath := range parent.GetPaths() {
			c.Paths = append(c.Paths, append(parentPath, c.Name()))
		}
	}
}

func (c *ReportImage) Diff(other *ReportImage) *ReportTreeItemDiffs {
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

// GetSQL implements ReportLeafNode
func (c *ReportImage) GetSQL() string {
	return typehelpers.SafeString(c.SQL)
}

// GetWidth implements ReportLeafNode
func (c *ReportImage) GetWidth() int {
	if c.Width == nil {
		return 0
	}
	return *c.Width
}

// GetUnqualifiedName implements ReportLeafNode, ModTreeItem
func (c *ReportImage) GetUnqualifiedName() string {
	return c.UnqualifiedName
}
