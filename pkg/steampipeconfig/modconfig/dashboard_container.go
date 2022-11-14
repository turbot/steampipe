package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/stevenle/topsort"
	"github.com/turbot/steampipe/pkg/utils"
)

// DashboardContainer is a struct representing the Dashboard and Container resource
type DashboardContainer struct {
	ResourceWithMetadataBase
	HclResourceBase
	ModTreeItemBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain"`

	Width      *int              `cty:"width" hcl:"width"  column:"width,text"`
	Display    *string           `cty:"display" hcl:"display"`
	Inputs     []*DashboardInput `cty:"inputs" column:"inputs,jsonb"`
	References []*ResourceReference

	Paths []NodePath `column:"path,jsonb"`
	// store children in a way which can be serialised via cty
	ChildNames []string `cty:"children" column:"children,jsonb"`

	runtimeDependencyGraph *topsort.Graph
}

func NewDashboardContainer(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)

	c := &DashboardContainer{
		HclResourceBase: HclResourceBase{
			ShortName:       shortName,
			FullName:        fullName,
			UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
			DeclRange:       block.DefRange,
		},
		ModTreeItemBase: ModTreeItemBase{
			Mod:      mod,
			fullName: fullName,
		},
	}
	c.SetAnonymous(block)

	return c
}

func (c *DashboardContainer) Equals(other *DashboardContainer) bool {
	diff := c.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource
func (c *DashboardContainer) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	c.ChildNames = make([]string, len(c.children))
	for i, child := range c.children {
		c.ChildNames[i] = child.Name()
	}
	return nil
}

// AddReference implements ResourceWithMetadata
func (c *DashboardContainer) AddReference(ref *ResourceReference) {
	c.References = append(c.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (c *DashboardContainer) GetReferences() []*ResourceReference {
	return c.References
}

// GetMod implements ModTreeItem
func (c *DashboardContainer) GetMod() *Mod {
	return c.Mod
}

// GetDocumentation implement ModTreeItem
func (*DashboardContainer) GetDocumentation() string {
	return ""
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

	if !utils.SafeStringsEqual(c.Display, other.Display) {
		res.AddPropertyDiff("Display")
	}

	res.populateChildDiffs(c, other)
	return res
}

func (c *DashboardContainer) SetChildren(children []ModTreeItem) {
	c.children = children
}

func (c *DashboardContainer) AddChild(child ModTreeItem) {
	c.children = append(c.children, child)
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
