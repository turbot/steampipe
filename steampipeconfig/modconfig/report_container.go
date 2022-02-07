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
const runtimeDependencyReportScope = "self"

// ReportContainer is a struct representing the Report and Container resource
type ReportContainer struct {
	HclResourceBase
	ResourceWithMetadataBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain"`

	ShortName       string
	FullName        string           `cty:"name"`
	UnqualifiedName string           `cty:"unqualified_name"`
	Title           *string          `cty:"title" hcl:"title" column:"title,text"`
	Width           *int             `cty:"width" hcl:"width"  column:"width,text"`
	Args            *QueryArgs       `cty:"args" column:"args,jsonb" json:"args"`
	Base            *ReportContainer `hcl:"base"`
	Inputs          []*ReportInput   `cty:"inputs"`

	Mod       *Mod `cty:"mod"`
	DeclRange hcl.Range
	Paths     []NodePath `column:"path,jsonb"`
	// store children in a way which can be serialised via cty
	ChildNames []string `cty:"children" column:"children,jsonb"`

	selfInputsMap map[string]*ReportInput
	// the actual children
	children               []ModTreeItem
	parents                []ModTreeItem
	runtimeDependencyGraph *topsort.Graph

	HclType string
}

func (c *ReportContainer) GetAnonymousChildName(t string) string {
	//TODO KAI
}

func NewReportContainer(block *hcl.Block, mod *Mod, parent HclResource) *ReportContainer {
	c := &ReportContainer{
		DeclRange:       block.DefRange,
		HclType:         block.Type,
		ShortName:       block.Labels[0],
		FullName:        fmt.Sprintf("%s.%s", block.Type, block.Labels[0]),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, block.Labels[0]),
	}
	c.SetMod(mod)
	return c
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
		c.copyChildrenFromBase()
	}
	if len(c.Inputs) == 0 {
		c.Inputs = c.Base.Inputs
		c.setInputMap()
	}

}

func (c *ReportContainer) copyChildrenFromBase() {
	var names []string
	for _, child := range c.Base.GetChildren(){
		// generate new name if anonymous
		// add to mod
		cloned := child.(ReportNode).CloneWithNewParent(c)
		c.children = append(c.children, cloned)
		c.ChildNames = append(c.ChildNames, cloned.Name())
	}
	c.ChildNames = names
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

func (c *ReportContainer) Diff(other *ReportContainer) *ReportTreeItemDiffs {
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

	if !utils.SafeIntEqual(c.Width, other.Width) {
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

func (c *ReportContainer) WalkResources(resourceFunc func(resource HclResource) (bool, error)) error {
	for _, child := range c.children {
		continueWalking, err := resourceFunc(child.(HclResource))
		if err != nil {
			return err
		}
		if !continueWalking {
			break
		}

		if childContainer, ok := child.(*ReportContainer); ok {
			if err := childContainer.WalkResources(resourceFunc); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *ReportContainer) BuildRuntimeDependencyTree(workspace ResourceMapsProvider) error {
	if !c.IsReport() {
		return fmt.Errorf("BuildRuntimeDependencyTree should only be called for reports")
	}
	c.runtimeDependencyGraph = topsort.NewGraph()
	// add root node - this will depend on all other nodes
	c.runtimeDependencyGraph.AddNode(rootRuntimeDependencyNode)

	resourceFunc := func(resource HclResource) (bool, error) {
		runtimeDependencies := resource.GetRuntimeDependencies()
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

func (c *ReportContainer) GetInput(name string) (*ReportInput, bool) {
	input, found := c.selfInputsMap[name]
	return input, found
}

func (c *ReportContainer) SetInputs(inputs []*ReportInput) error {
	c.Inputs = inputs
	c.setInputMap()

	// also add child containers inputs

	var duplicates []string
	resourceFunc := func(resource HclResource) (bool, error) {
		if container, ok := resource.(*ReportContainer); ok {
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
func (c *ReportContainer) setInputMap() {
	c.selfInputsMap = make(map[string]*ReportInput)
	for _, i := range c.Inputs {
		c.selfInputsMap[i.UnqualifiedName] = i
	}
}

func (c *ReportContainer) SetArgs(args *QueryArgs) {
	c.Args = args
}
