package modconfig

import (
	"fmt"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/stevenle/topsort"
	"github.com/turbot/steampipe/pkg/utils"
)

const rootRuntimeDependencyNode = "rootRuntimeDependencyNode"
const RuntimeDependencyDashboardScope = "self"

// Dashboard is a struct representing the Dashboard  resource
type Dashboard struct {
	ResourceWithMetadataImpl
	ModTreeItemImpl
	WithProviderImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Width   *int              `cty:"width" hcl:"width"  column:"width,text"`
	Display *string           `cty:"display" hcl:"display" column:"display,text"`
	Inputs  []*DashboardInput `cty:"inputs" column:"inputs,jsonb"`
	UrlPath string            `cty:"url_path"  column:"url_path,jsonb"`
	Base    *Dashboard        `hcl:"base"`
	// store children in a way which can be serialised via cty
	ChildNames []string `cty:"children" column:"children,jsonb"`
	// map of all inputs in our resource tree
	selfInputsMap          map[string]*DashboardInput
	runtimeDependencyGraph *topsort.Graph
}

func NewDashboard(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)

	c := &Dashboard{
		ModTreeItemImpl: ModTreeItemImpl{
			HclResourceImpl: HclResourceImpl{
				ShortName:       shortName,
				FullName:        fullName,
				UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
				DeclRange:       BlockRange(block),
				blockType:       block.Type,
			},
			Mod: mod,
		},
	}
	c.SetAnonymous(block)
	c.setUrlPath()

	return c
}

// NewQueryDashboard creates a dashboard to wrap a query/control
// this is used for snapshot generation
func NewQueryDashboard(q ModTreeItem) (*Dashboard, error) {
	parsedName, err := ParseResourceName(q.Name())
	if err != nil {
		return nil, err
	}
	dashboardName := BuildFullResourceName(q.GetMod().ShortName, BlockTypeDashboard, parsedName.Name)

	var dashboard = &Dashboard{
		ResourceWithMetadataImpl: ResourceWithMetadataImpl{
			metadata: &ResourceMetadata{},
		},
		ModTreeItemImpl: ModTreeItemImpl{
			HclResourceImpl: HclResourceImpl{
				ShortName:       parsedName.Name,
				FullName:        dashboardName,
				UnqualifiedName: fmt.Sprintf("%s.%s", BlockTypeDashboard, parsedName),
				Title:           utils.ToStringPointer(q.GetTitle()),
				Description:     utils.ToStringPointer(q.GetDescription()),
				Documentation:   utils.ToStringPointer(q.GetDocumentation()),
				Tags:            q.GetTags(),
				blockType:       BlockTypeDashboard,
			},
			Mod: q.GetMod(),
		},
	}

	dashboard.setUrlPath()

	chart, err := NewQueryDashboardTable(q)
	if err != nil {
		return nil, err
	}
	dashboard.children = []ModTreeItem{chart}

	return dashboard, nil
}

func (d *Dashboard) setUrlPath() {
	d.UrlPath = fmt.Sprintf("/%s", d.FullName)
}

func (d *Dashboard) Equals(other *Dashboard) bool {
	diff := d.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource
func (d *Dashboard) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	d.setBaseProperties()

	d.ChildNames = make([]string, len(d.children))
	for i, child := range d.children {
		d.ChildNames[i] = child.Name()
	}

	return nil
}

// GetWidth implements DashboardLeafNode
func (d *Dashboard) GetWidth() int {
	if d.Width == nil {
		return 0
	}
	return *d.Width
}

// GetDisplay implements DashboardLeafNode
func (d *Dashboard) GetDisplay() string {
	return typehelpers.SafeString(d.Display)
}

// GetType implements DashboardLeafNode
func (d *Dashboard) GetType() string {
	return ""
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

	if len(d.Tags) != len(other.Tags) {
		res.AddPropertyDiff("Tags")
	} else {
		for k, v := range d.Tags {
			if otherVal := other.Tags[k]; v != otherVal {
				res.AddPropertyDiff("Tags")
			}
		}
	}

	if !utils.SafeStringsEqual(d.Description, other.Description) {
		res.AddPropertyDiff("Description")
	}

	if !utils.SafeStringsEqual(d.Documentation, other.Documentation) {
		res.AddPropertyDiff("Documentation")
	}

	res.populateChildDiffs(d, other)
	return res
}

func (d *Dashboard) SetChildren(children []ModTreeItem) {
	d.children = children
}

func (d *Dashboard) AddChild(child ModTreeItem) {
	d.children = append(d.children, child)

	switch c := child.(type) {
	case *DashboardInput:
		d.Inputs = append(d.Inputs, c)
	case *DashboardWith:
		d.AddWith(c)
	}
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

		if container, ok := child.(*DashboardContainer); ok {
			if err := container.WalkResources(resourceFunc); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Dashboard) ValidateRuntimeDependencies(workspace ResourceMapsProvider) error {
	d.runtimeDependencyGraph = topsort.NewGraph()
	// add root node - this will depend on all other nodes
	d.runtimeDependencyGraph.AddNode(rootRuntimeDependencyNode)

	// define a walk function which determines whether the resource has runtime dependencies and if so,
	// add to the graph
	resourceFunc := func(resource HclResource) (bool, error) {
		wp, ok := resource.(WithProvider)
		if !ok {
			// continue walking
			return true, nil
		}

		if err := d.validateRuntimeDependenciesForResource(resource, workspace); err != nil {
			return false, err
		}

		// if the query provider has any 'with' blocks, add these dependencies as well
		for _, with := range wp.GetWiths() {
			if err := d.validateRuntimeDependenciesForResource(with, workspace); err != nil {
				return false, err
			}
		}

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

func (d *Dashboard) validateRuntimeDependenciesForResource(resource HclResource, workspace ResourceMapsProvider) error {
	// TODO  [node_reuse] re-add parse time validation https://github.com/turbot/steampipe/issues/2925
	return nil
	//rdp := resource.(RuntimeDependencyProvider)
	//// WHAT ABOUT CHILDREN
	//if len(runtimeDependencies) == 0 {
	//	return nil
	//}
	//name := resource.Name()
	//if !d.runtimeDependencyGraph.ContainsNode(name) {
	//	d.runtimeDependencyGraph.AddNode(name)
	//}
	//
	//for _, dependency := range runtimeDependencies {
	//	// try to resolve the dependency source resource
	//	if err := dependency.ValidateSource(d, workspace); err != nil {
	//		return err
	//	}
	//	if err := d.runtimeDependencyGraph.AddEdge(rootRuntimeDependencyNode, name); err != nil {
	//		return err
	//	}
	//	depString := dependency.String()
	//	if !d.runtimeDependencyGraph.ContainsNode(depString) {
	//		d.runtimeDependencyGraph.AddNode(depString)
	//	}
	//	if err := d.runtimeDependencyGraph.AddEdge(name, dependency.String()); err != nil {
	//		return err
	//	}
	//}
	//return nil
}

func (d *Dashboard) GetInput(name string) (*DashboardInput, bool) {
	input, found := d.selfInputsMap[name]
	return input, found
}

func (d *Dashboard) GetInputs() map[string]*DashboardInput {
	return d.selfInputsMap
}

func (d *Dashboard) InitInputs() hcl.Diagnostics {
	// add all our direct child inputs to a map
	// (we must do this before adding child container inputs to detect dupes)
	duplicates := d.setInputMap()

	//  add child containers and dashboard inputs
	resourceFunc := func(resource HclResource) (bool, error) {
		if container, ok := resource.(*DashboardContainer); ok {
			for _, i := range container.Inputs {
				// check we do not already have this input
				if _, ok := d.selfInputsMap[i.UnqualifiedName]; ok {
					duplicates = append(duplicates, i.Name())

				}
				d.Inputs = append(d.Inputs, i)
				d.selfInputsMap[i.UnqualifiedName] = i
			}
		}
		// continue walking
		return true, nil
	}
	if err := d.WalkResources(resourceFunc); err != nil {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Dashboard '%s' WalkResources failed", d.Name()),
			Detail:   err.Error(),
			Subject:  &d.DeclRange,
		}}
	}

	if len(duplicates) > 0 {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Dashboard '%s' contains duplicate input names for: %s", d.Name(), strings.Join(duplicates, ",")),
			Subject:  &d.DeclRange,
		}}
	}

	var diags hcl.Diagnostics
	//  ensure they inputs not have cyclical dependencies
	if err := d.validateInputDependencies(d.Inputs); err != nil {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Failed to resolve input dependency order for dashboard '%s'", d.Name()),
			Detail:   err.Error(),
			Subject:  &d.DeclRange,
		}}
	}
	// now 'claim' all inputs and add to mod
	for _, input := range d.Inputs {
		input.SetDashboard(d)
		moreDiags := d.Mod.AddResource(input)
		diags = append(diags, moreDiags...)
	}

	return diags
}

// populate our input map
func (d *Dashboard) setInputMap() []string {
	var duplicates []string
	d.selfInputsMap = make(map[string]*DashboardInput)
	for _, i := range d.Inputs {
		if _, ok := d.selfInputsMap[i.UnqualifiedName]; ok {
			duplicates = append(duplicates, i.UnqualifiedName)
		} else {
			d.selfInputsMap[i.UnqualifiedName] = i
		}
	}
	return duplicates
}

// CtyValue implements CtyValueProvider
func (d *Dashboard) CtyValue() (cty.Value, error) {
	return GetCtyValue(d)
}

func (d *Dashboard) setBaseProperties() {
	if d.Base == nil {
		return
	}
	// copy base into the HclResourceImpl 'base' property so it is accessible to all nested structs
	d.base = d.Base
	// call into parent nested struct setBaseProperties
	d.ModTreeItemImpl.setBaseProperties()

	if d.Width == nil {
		d.Width = d.Base.Width
	}

	if len(d.children) == 0 {
		d.children = d.Base.children
		d.ChildNames = d.Base.ChildNames
	}

	d.addBaseInputs(d.Base.Inputs)
}

func (d *Dashboard) addBaseInputs(baseInputs []*DashboardInput) {
	if len(baseInputs) == 0 {
		return
	}
	// rebuild Inputs and children
	inheritedInputs := make([]*DashboardInput, len(baseInputs))
	inheritedChildren := make([]ModTreeItem, len(baseInputs))

	for i, baseInput := range baseInputs {
		input := baseInput.Clone()
		input.SetDashboard(d)
		// add to mod
		d.Mod.AddResource(input)
		// add to our inputs
		inheritedInputs[i] = input
		inheritedChildren[i] = input

	}
	// add inputs to beginning of our existing inputs (if any)
	d.Inputs = append(inheritedInputs, d.Inputs...)
	// add inputs to beginning of our children
	d.children = append(inheritedChildren, d.children...)
	d.setInputMap()
}

// ensure that dependencies between inputs are resolveable
func (d *Dashboard) validateInputDependencies(inputs []*DashboardInput) error {
	dependencyGraph := topsort.NewGraph()
	rootDependencyNode := "dashboard"
	dependencyGraph.AddNode(rootDependencyNode)
	for _, i := range inputs {
		for _, runtimeDep := range i.GetRuntimeDependencies() {
			depName := runtimeDep.PropertyPath.ToResourceName()
			to := depName
			from := i.UnqualifiedName
			if !dependencyGraph.ContainsNode(from) {
				dependencyGraph.AddNode(from)
			}
			if !dependencyGraph.ContainsNode(to) {
				dependencyGraph.AddNode(to)
			}
			if err := dependencyGraph.AddEdge(from, to); err != nil {
				return err
			}
			if err := dependencyGraph.AddEdge(rootDependencyNode, i.UnqualifiedName); err != nil {
				return err
			}
		}
	}

	// now verify we can get a dependency order
	_, err := dependencyGraph.TopSort(rootDependencyNode)
	return err
}
