package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/stevenle/topsort"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/utils"
	"github.com/zclconf/go-cty/cty"
)

// DashboardContainer is a struct representing the Dashboard and Container resource
type DashboardContainer struct {
	DashboardLeafNodeBase
	ResourceWithMetadataBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain"`

	ShortName       string
	FullName        string  `cty:"name"`
	UnqualifiedName string  `cty:"unqualified_name"`
	Title           *string `cty:"title" hcl:"title" column:"title,text"`
	Width           *int    `cty:"width" hcl:"width"  column:"width,text"`

	Mod       *Mod `cty:"mod"`
	DeclRange hcl.Range
	Paths     []NodePath `column:"path,jsonb"`
	// store children in a way which can be serialised via cty
	ChildNames []string `cty:"children" column:"children,jsonb"`

	// the actual children
	children               []ModTreeItem
	parents                []ModTreeItem
	runtimeDependencyGraph *topsort.Graph
}

func NewDashboardContainer(block *hcl.Block, mod *Mod) *DashboardContainer {
	shortName := GetAnonymousResourceShortName(block, mod)
	c := &DashboardContainer{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}
	c.SetAnonymous(block)

	return c
}

func (c *DashboardContainer) Equals(other *DashboardContainer) bool {
	diff := c.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (c *DashboardContainer) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

// Name implements HclResource, ModTreeItem
func (c *DashboardContainer) Name() string {
	return c.FullName
}

// OnDecoded implements HclResource
func (c *DashboardContainer) OnDecoded(block *hcl.Block) hcl.Diagnostics {
	return nil
}

// AddReference implements HclResource
func (c *DashboardContainer) AddReference(*ResourceReference) {
	// TODO
}

// GetMod implements HclResource
func (c *DashboardContainer) GetMod() *Mod {
	return c.Mod
}

// GetDeclRange implements HclResource
func (c *DashboardContainer) GetDeclRange() *hcl.Range {
	return &c.DeclRange
}

// AddParent implements ModTreeItem
func (c *DashboardContainer) AddParent(parent ModTreeItem) error {
	c.parents = append(c.parents, parent)

	return nil
}

// GetParents implements ModTreeItem
func (c *DashboardContainer) GetParents() []ModTreeItem {
	return c.parents
}

// GetChildren implements ModTreeItem
func (c *DashboardContainer) GetChildren() []ModTreeItem {
	return c.children
}

// GetTitle implements ModTreeItem
func (c *DashboardContainer) GetTitle() string {
	return typehelpers.SafeString(c.Title)
}

// GetDescription implements ModTreeItem
func (c *DashboardContainer) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (c *DashboardContainer) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (c *DashboardContainer) GetPaths() []NodePath {
	// lazy load
	if len(c.Paths) == 0 {
		c.SetPaths()
	}
	return c.Paths
}

// SetPaths implements ModTreeItem
func (c *DashboardContainer) SetPaths() {
	for _, parent := range c.parents {
		for _, parentPath := range parent.GetPaths() {
			c.Paths = append(c.Paths, append(parentPath, c.Name()))
		}
	}
}

func (c *DashboardContainer) Diff(other *DashboardContainer) *DashboardTreeItemDiffs {
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

	if !utils.SafeIntEqual(c.Width, other.Width) {
		res.AddPropertyDiff("Width")
	}

	res.populateChildDiffs(c, other)
	return res
}

func (c *DashboardContainer) SetChildren(children []ModTreeItem) {
	c.children = children
	c.ChildNames = make([]string, len(children))
	for i, child := range children {
		c.ChildNames[i] = child.Name()
	}
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (c *DashboardContainer) GetUnqualifiedName() string {
	return c.UnqualifiedName
}

func (c *DashboardContainer) WalkResources(resourceFunc func(resource HclResource) (bool, error)) error {
	for _, child := range c.children {
		continueWalking, err := resourceFunc(child.(HclResource))
		if err != nil {
			return err
		}
		if !continueWalking {
			break
		}

		if childContainer, ok := child.(*DashboardContainer); ok {
			if err := childContainer.WalkResources(resourceFunc); err != nil {
				return err
			}
		}
	}
	return nil
}
