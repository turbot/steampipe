package modconfig

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/stevenle/topsort"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/utils"
	"github.com/zclconf/go-cty/cty"
)

const rootRuntimeDependencyNode = "rootRuntimeDependencyNode"
const runtimeDependencyDashboardScope = "self"

// DashboardContainer is a struct representing the Dashboard and Container resource
type DashboardContainer struct {
	DashboardLeafNodeBase
	ResourceWithMetadataBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain"`

	ShortName       string
	FullName        string              `cty:"name"`
	UnqualifiedName string              `cty:"unqualified_name"`
	Title           *string             `cty:"title" hcl:"title" column:"title,text"`
	Width           *int             `cty:"width" hcl:"width"  column:"width,text"`
	Args            *QueryArgs          `cty:"args" column:"args,jsonb"`
	Base            *DashboardContainer `hcl:"base"`
	Inputs          []*DashboardInput   `cty:"inputs"`

	Mod       *Mod `cty:"mod"`
	DeclRange hcl.Range
	Paths     []NodePath `column:"path,jsonb"`
	// store children in a way which can be serialised via cty
	ChildNames []string `cty:"children" column:"children,jsonb"`

	selfInputsMap map[string]*DashboardInput
	// the actual children
	children               []ModTreeItem
	parents                []ModTreeItem
	runtimeDependencyGraph *topsort.Graph

	HclType string
}

func NewDashboardContainer(block *hcl.Block, mod *Mod) *DashboardContainer {
	// TODO [reports] think about nested report???
	shortName := GetAnonymousResourceShortName(block, mod)
	c := &DashboardContainer{
		HclType:         block.Type,
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}
	c.SetAnonymous(block)

	return c
}

//// SetMod implements HclResource
//func (c *DashboardContainer) SetMod(mod *Mod) {
//	c.Mod = mod
//
//	// if this is a top level resource, and not a child, the resource names will already be set
//	// - we need to update the full name to include the mod
//	if c.UnqualifiedName != "" {
//		// add mod name to full name
//		c.FullName = fmt.Sprintf("%s.%s", mod.ShortName, c.UnqualifiedName)
//	}
//}

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
	c.setBaseProperties()
	return nil
}

func (c *DashboardContainer) setBaseProperties() {
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
		c.children = c.Base.children
		c.ChildNames = c.Base.ChildNames
	}
	if len(c.Inputs) == 0 {
		c.Inputs = c.Base.Inputs
		c.setInputMap()
	}
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

func (c *DashboardContainer) IsDashboard() bool {
	return c.HclType == BlockTypeDashboard
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

func (c *DashboardContainer) BuildRuntimeDependencyTree(workspace ResourceMapsProvider) error {
	if !c.IsDashboard() {
		return fmt.Errorf("BuildRuntimeDependencyTree should only be called for dashboards")
	}
	c.runtimeDependencyGraph = topsort.NewGraph()
	// add root node - this will depend on all other nodes
	c.runtimeDependencyGraph.AddNode(rootRuntimeDependencyNode)

	resourceFunc := func(resource HclResource) (bool, error) {
		leafNode, ok := resource.(DashboardLeafNode)
		if !ok {
			// continue walking
			return true, nil
		}
		runtimeDependencies := leafNode.GetRuntimeDependencies()
		if len(runtimeDependencies) == 0 {
			// continue walking
			return true, nil
		}
		name := resource.Name()
		if !c.runtimeDependencyGraph.ContainsNode(name) {
			c.runtimeDependencyGraph.AddNode(name)
		}

		for _, dependency := range runtimeDependencies {
			// try to resolve the target resource
			if err := dependency.ResolveSource(resource, c, workspace); err != nil {
				return false, err
			}
			if err := c.runtimeDependencyGraph.AddEdge(rootRuntimeDependencyNode, name); err != nil {
				return false, err
			}
			depString := dependency.String()
			if !c.runtimeDependencyGraph.ContainsNode(depString) {
				c.runtimeDependencyGraph.AddNode(depString)
			}
			if err := c.runtimeDependencyGraph.AddEdge(name, dependency.String()); err != nil {
				return false, err
			}
		}

		// ensure that all parameters have corresponding args populated with a value or runtime dependency

		// continue walking
		return true, nil
	}
	if err := c.WalkResources(resourceFunc); err != nil {
		return err
	}

	// ensure that dependencies can be resolved
	if _, err := c.runtimeDependencyGraph.TopSort(rootRuntimeDependencyNode); err != nil {
		return fmt.Errorf("runtime depedencies cannot be resolved: %s", err.Error())
	}
	return nil
}

func (c *DashboardContainer) GetInput(name string) (*DashboardInput, bool) {
	input, found := c.selfInputsMap[name]
	return input, found
}

func (c *DashboardContainer) SetInputs(inputs []*DashboardInput) error {
	c.Inputs = inputs
	c.setInputMap()

	// also add child containers inputs

	var duplicates []string
	resourceFunc := func(resource HclResource) (bool, error) {
		if container, ok := resource.(*DashboardContainer); ok {
			for _, i := range container.Inputs {
				// check we do not already have this input
				if _, ok := c.selfInputsMap[i.UnqualifiedName]; ok {
					duplicates = append(duplicates, i.Name())
					continue
				}
				c.Inputs = append(c.Inputs, i)
				c.selfInputsMap[i.UnqualifiedName] = i
			}
		}
		// continue walking
		return true, nil
	}
	c.WalkResources(resourceFunc)

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate input names found for %s: %s", c.Name(), strings.Join(duplicates, ","))
	}
	return nil
}

// populate our input map
func (c *DashboardContainer) setInputMap() {
	c.selfInputsMap = make(map[string]*DashboardInput)
	for _, i := range c.Inputs {
		c.selfInputsMap[i.UnqualifiedName] = i
	}
}

func (c *DashboardContainer) SetArgs(args *QueryArgs) {
	c.Args = args
}
