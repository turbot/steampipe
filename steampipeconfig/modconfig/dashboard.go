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

// Dashboard is a struct representing the Dashboard and Container resource
type Dashboard struct {
	DashboardLeafNodeBase
	ResourceWithMetadataBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain"`

	ShortName       string
	FullName        string            `cty:"name"`
	UnqualifiedName string            `cty:"unqualified_name"`
	Title           *string           `cty:"title" hcl:"title" column:"title,text"`
	Width           *int              `cty:"width" hcl:"width"  column:"width,text"`
	Args            *QueryArgs        `cty:"args" column:"args,jsonb"`
	Base            *Dashboard        `hcl:"base"`
	Inputs          []*DashboardInput `cty:"inputs"`

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

func NewDashboard(block *hcl.Block, mod *Mod) *Dashboard {
	// TODO [reports] think about nested report???
	shortName := GetAnonymousResourceShortName(block, mod)
	c := &Dashboard{
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

func (d *Dashboard) Equals(other *Dashboard) bool {
	diff := d.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (d *Dashboard) CtyValue() (cty.Value, error) {
	return getCtyValue(d)
}

// Name implements HclResource, ModTreeItem
func (d *Dashboard) Name() string {
	return d.FullName
}

// OnDecoded implements HclResource
func (d *Dashboard) OnDecoded(block *hcl.Block) hcl.Diagnostics {
	d.setBaseProperties()
	return nil
}

func (d *Dashboard) setBaseProperties() {
	if d.Base == nil {
		return
	}
	if d.Title == nil {
		d.Title = d.Base.Title
	}
	if d.Width == nil {
		d.Width = d.Base.Width
	}
	if len(d.children) == 0 {
		d.children = d.Base.children
		d.ChildNames = d.Base.ChildNames
	}
	if len(d.Inputs) == 0 {
		d.Inputs = d.Base.Inputs
		d.setInputMap()
	}
}

// AddReference implements HclResource
func (d *Dashboard) AddReference(*ResourceReference) {
	// TODO
}

// GetMod implements HclResource
func (d *Dashboard) GetMod() *Mod {
	return d.Mod
}

// GetDeclRange implements HclResource
func (d *Dashboard) GetDeclRange() *hcl.Range {
	return &d.DeclRange
}

// AddParent implements ModTreeItem
func (d *Dashboard) AddParent(parent ModTreeItem) error {
	d.parents = append(d.parents, parent)

	return nil
}

// GetParents implements ModTreeItem
func (d *Dashboard) GetParents() []ModTreeItem {
	return d.parents
}

// GetChildren implements ModTreeItem
func (d *Dashboard) GetChildren() []ModTreeItem {
	return d.children
}

// GetTitle implements ModTreeItem
func (d *Dashboard) GetTitle() string {
	return typehelpers.SafeString(d.Title)
}

// GetDescription implements ModTreeItem
func (d *Dashboard) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (d *Dashboard) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (d *Dashboard) GetPaths() []NodePath {
	// lazy load
	if len(d.Paths) == 0 {
		d.SetPaths()
	}
	return d.Paths
}

// SetPaths implements ModTreeItem
func (d *Dashboard) SetPaths() {
	for _, parent := range d.parents {
		for _, parentPath := range parent.GetPaths() {
			d.Paths = append(d.Paths, append(parentPath, d.Name()))
		}
	}
}

func (d *Dashboard) Diff(other *Dashboard) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: d,
		Name: d.Name(),
	}

	if !utils.SafeStringsEqual(d.FullName, other.FullName) {
		res.AddPropertyDiff("Name")
	}

	if !utils.SafeStringsEqual(d.Title, other.Title) {
		res.AddPropertyDiff("Title")
	}

	if !utils.SafeIntEqual(d.Width, other.Width) {
		res.AddPropertyDiff("Width")
	}

	res.populateChildDiffs(d, other)
	return res
}

func (d *Dashboard) SetChildren(children []ModTreeItem) {
	d.children = children
	d.ChildNames = make([]string, len(children))
	for i, child := range children {
		d.ChildNames[i] = child.Name()
	}
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (d *Dashboard) GetUnqualifiedName() string {
	return d.UnqualifiedName
}

func (d *Dashboard) WalkResources(resourceFunc func(resource HclResource) (bool, error)) error {
	for _, child := range d.children {
		continueWalking, err := resourceFunc(child.(HclResource))
		if err != nil {
			return err
		}
		if !continueWalking {
			break
		}

		if childContainer, ok := child.(*Dashboard); ok {
			if err := childContainer.WalkResources(resourceFunc); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Dashboard) BuildRuntimeDependencyTree(workspace ResourceMapsProvider) error {
	d.runtimeDependencyGraph = topsort.NewGraph()
	// add root node - this will depend on all other nodes
	d.runtimeDependencyGraph.AddNode(rootRuntimeDependencyNode)

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
		if !d.runtimeDependencyGraph.ContainsNode(name) {
			d.runtimeDependencyGraph.AddNode(name)
		}

		for _, dependency := range runtimeDependencies {
			// try to resolve the target resource
			if err := dependency.ResolveSource(resource, d, workspace); err != nil {
				return false, err
			}
			if err := d.runtimeDependencyGraph.AddEdge(rootRuntimeDependencyNode, name); err != nil {
				return false, err
			}
			depString := dependency.String()
			if !d.runtimeDependencyGraph.ContainsNode(depString) {
				d.runtimeDependencyGraph.AddNode(depString)
			}
			if err := d.runtimeDependencyGraph.AddEdge(name, dependency.String()); err != nil {
				return false, err
			}
		}

		// ensure that all parameters have corresponding args populated with a value or runtime dependency

		// continue walking
		return true, nil
	}
	if err := d.WalkResources(resourceFunc); err != nil {
		return err
	}

	// ensure that dependencies can be resolved
	if _, err := d.runtimeDependencyGraph.TopSort(rootRuntimeDependencyNode); err != nil {
		return fmt.Errorf("runtime depedencies cannot be resolved: %s", err.Error())
	}
	return nil
}

func (d *Dashboard) GetInput(name string) (*DashboardInput, bool) {
	input, found := d.selfInputsMap[name]
	return input, found
}

func (d *Dashboard) SetInputs(inputs []*DashboardInput) error {
	d.Inputs = inputs
	d.setInputMap()

	// also add child containers inputs

	var duplicates []string
	resourceFunc := func(resource HclResource) (bool, error) {
		if container, ok := resource.(*Dashboard); ok {
			for _, i := range container.Inputs {
				// check we do not already have this input
				if _, ok := d.selfInputsMap[i.UnqualifiedName]; ok {
					duplicates = append(duplicates, i.Name())
					continue
				}
				d.Inputs = append(d.Inputs, i)
				d.selfInputsMap[i.UnqualifiedName] = i
			}
		}
		// continue walking
		return true, nil
	}
	d.WalkResources(resourceFunc)

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate input names found for %s: %s", d.Name(), strings.Join(duplicates, ","))
	}
	return nil
}

// populate our input map
func (d *Dashboard) setInputMap() {
	d.selfInputsMap = make(map[string]*DashboardInput)
	for _, i := range d.Inputs {
		d.selfInputsMap[i.UnqualifiedName] = i
	}
}

func (d *Dashboard) SetArgs(args *QueryArgs) {
	d.Args = args
}
