package parse

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/stevenle/topsort"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/steampipeconfig/hclhelpers"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/zclconf/go-cty/cty"
)

const rootDependencyNode = "rootDependencyNode"

type unresolvedBlock struct {
	Name         string
	Block        *hcl.Block
	Dependencies []*dependency
}

func (b unresolvedBlock) String() string {
	depStrings := make([]string, len(b.Dependencies))
	for i, dep := range b.Dependencies {
		depStrings[i] = fmt.Sprintf(`%s -> %s`, b.Name, dep.String())
	}
	return strings.Join(depStrings, "\n")
}

type RunContext struct {
	Mod              *modconfig.Mod
	UnresolvedBlocks map[string]*unresolvedBlock
	FileData         map[string][]byte
	dependencyGraph  *topsort.Graph
	//dependencies     map[string]bool
	// store any objects which are depdnecy targets
	variables map[string]map[string]cty.Value
	EvalCtx   *hcl.EvalContext
	blocks    hcl.Blocks
}

func NewRunContext(mod *modconfig.Mod, content *hcl.BodyContent, fileData map[string][]byte) (*RunContext, hcl.Diagnostics) {
	c := &RunContext{
		Mod:              mod,
		FileData:         fileData,
		UnresolvedBlocks: make(map[string]*unresolvedBlock),

		variables: make(map[string]map[string]cty.Value),
		blocks:    content.Blocks,
	}
	// add root node - this will depend on all other nodes
	c.dependencyGraph = c.newDependencyGraph()
	c.buildEvalContext()

	// add mod and any existing mod resources to the variables
	diags := c.addModToVariables()

	c.addSteampipeVariables()
	return c, diags
}

// add enums to the variables which may be referenced from within the hcl
func (c *RunContext) addSteampipeVariables() {
	// add steampipe variables
	c.variables["steampipe"] = map[string]cty.Value{
		"panel": cty.ObjectVal(map[string]cty.Value{
			"markdown":         cty.StringVal("steampipe.panel.markdown"),
			"barchart":         cty.StringVal("steampipe.panel.barchart"),
			"stackedbarchart":  cty.StringVal("steampipe.panel.stackedbarchart"),
			"counter":          cty.StringVal("steampipe.panel.counter"),
			"linechart":        cty.StringVal("steampipe.panel.linechart"),
			"multilinechart":   cty.StringVal("steampipe.panel.multilinechart"),
			"piechart":         cty.StringVal("steampipe.panel.piechart"),
			"placeholder":      cty.StringVal("steampipe.panel.placeholder"),
			"control_list":     cty.StringVal("steampipe.panel.control_list"),
			"control_progress": cty.StringVal("steampipe.panel.control_progress"),
			"control_table":    cty.StringVal("steampipe.panel.control_table"),
			"status":           cty.StringVal("steampipe.panel.status"),
			"table":            cty.StringVal("steampipe.panel.table"),
			"resource_detail":  cty.StringVal("steampipe.panel.resource_detail"),
			"resource_tags":    cty.StringVal("steampipe.panel.resource_tags"),
		}),
	}
}

func (c *RunContext) newDependencyGraph() *topsort.Graph {
	dependencyGraph := topsort.NewGraph()
	// add root node - this will depend on all other nodes
	dependencyGraph.AddNode(rootDependencyNode)
	return dependencyGraph
}

func (c *RunContext) ClearDependencies() {
	c.UnresolvedBlocks = make(map[string]*unresolvedBlock)
	c.dependencyGraph = c.newDependencyGraph()
}

// AddDependencies :: the block could not be resolved as it has dependencies
// 1) store block as unresolved
// 2) add dependencies to our tree of dependencies
func (c *RunContext) AddDependencies(block *hcl.Block, name string, dependencies []*dependency) hcl.Diagnostics {
	var diags hcl.Diagnostics
	// store unresolved block
	c.UnresolvedBlocks[name] = &unresolvedBlock{Name: name, Block: block, Dependencies: dependencies}

	// store dependency in tree - d
	if !c.dependencyGraph.ContainsNode(name) {
		c.dependencyGraph.AddNode(name)
	}
	// add root dependency
	c.dependencyGraph.AddEdge(rootDependencyNode, name)

	for _, dep := range dependencies {
		// each dependency object may have multiple traversals
		for _, t := range dep.Traversals {
			d := hclhelpers.TraversalAsString(t)

			// 'd' may be a property path - when storing dependencies we only care about the resource names
			dependencyResource, err := modconfig.PropertyPathToResourceName(d)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "failed to convert cty value - asJson failed",
					Detail:   err.Error()})
				continue

			}
			if !c.dependencyGraph.ContainsNode(dependencyResource) {
				c.dependencyGraph.AddNode(dependencyResource)
			}
			c.dependencyGraph.AddEdge(name, dependencyResource)
		}
	}
	return nil
}

func (c *RunContext) getDependencyOrder() ([]string, error) {
	rawDeps, err := c.dependencyGraph.TopSort(rootDependencyNode)
	if err != nil {
		return nil, err
	}

	// now remove the variable names and dedupe
	var deps []string
	for _, d := range rawDeps {
		if d == rootDependencyNode {
			continue
		}

		propertyPath, err := modconfig.ParseResourcePropertyPath(d)
		if err != nil {
			return nil, err
		}
		dep := modconfig.BuildModResourceName(propertyPath.ItemType, propertyPath.Name)
		if !helpers.StringSliceContains(deps, dep) {
			deps = append(deps, dep)
		}
	}
	return deps, nil
}

func (c *RunContext) BlocksToDecode() (hcl.Blocks, error) {
	depOrder, err := c.getDependencyOrder()
	if err != nil {
		return nil, err
	}
	if len(depOrder) == 0 {
		return c.blocks, nil
	}

	var blocksToDecode hcl.Blocks
	for _, name := range depOrder {
		// depOrder is all the blocks required to resolve dependencies.
		// if this one is unparsed, added to list
		block, ok := c.UnresolvedBlocks[name]
		if ok {
			blocksToDecode = append(blocksToDecode, block.Block)
		}
	}
	return blocksToDecode, nil
}

// state

// EvalComplete :: Are all elements in the dependency tree fully evaluated
func (c *RunContext) EvalComplete() bool {
	return len(c.UnresolvedBlocks) == 0
}

// eval functions
func (c *RunContext) buildEvalContext() {
	// convert variables to cty values
	variables := make(map[string]cty.Value)
	for k, v := range c.variables {
		variables[k] = cty.ObjectVal(v)
	}
	// create evaluation context
	c.EvalCtx = &hcl.EvalContext{
		Variables: variables,
		Functions: ContextFunctions(c.Mod.ModPath),
	}
}

// AddResource :: store this resource as a variable to be added to the eval ccontext
func (c *RunContext) AddResource(resource modconfig.HclResource, block *hcl.Block) hcl.Diagnostics {
	diagnostics := c.addResourceToVariables(resource, block)
	if diagnostics.HasErrors() {
		return diagnostics
	}

	// try to add query to mod - this will fail if the mod already has a query with the same name
	return c.Mod.AddResource(resource, block)
}

func (c *RunContext) addResourceToVariables(resource modconfig.HclResource, block *hcl.Block) hcl.Diagnostics {
	// add resource to variable map
	ctyValue, err := resource.CtyValue()
	if err != nil {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("failed to convert resource '%s'to its cty value", resource.Name()),
			Detail:   err.Error(),
			Subject:  &block.DefRange,
		}}
	}

	parsedName, err := modconfig.ParseResourceName(resource.Name())
	if err != nil {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("failed to parse resourece name %s", resource.Name()),
			Detail:   err.Error(),
			Subject:  &block.DefRange,
		}}
	}

	typeString := parsedName.TypeString()
	variablesForType, ok := c.variables[typeString]
	if !ok {
		variablesForType = make(map[string]cty.Value)
	}
	variablesForType[parsedName.Name] = ctyValue
	c.variables[typeString] = variablesForType
	// rebuild the eval context
	c.buildEvalContext()

	// remove this resource from unparsed blocks
	if _, ok := c.UnresolvedBlocks[resource.Name()]; ok {
		delete(c.UnresolvedBlocks, resource.Name())
	}
	return nil
}

func (c *RunContext) addModToVariables() hcl.Diagnostics {
	var diags hcl.Diagnostics
	// create empty block to pass
	block := &hcl.Block{}

	moreDiags := c.addResourceToVariables(c.Mod, block)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
	}
	// add all mappable resources to variables
	for _, q := range c.Mod.Queries {
		moreDiags := c.addResourceToVariables(q, block)
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		}
	}
	return diags
}

func (c *RunContext) FormatDependencies() string {
	// first get the dependency order
	dependencyOrder, err := c.getDependencyOrder()
	if err != nil {
		return err.Error()
	}
	// build array of dependency strings - processes dependencies in reverse order for presentation reasons
	numDeps := len(dependencyOrder)
	depStrings := make([]string, numDeps)
	for i := 0; i < len(dependencyOrder); i++ {
		srcIdx := len(dependencyOrder) - i - 1
		resourceName := dependencyOrder[srcIdx]
		// find dependency
		dep, ok := c.UnresolvedBlocks[resourceName]

		if ok {
			depStrings[i] = dep.String()
		} else {
			// this could happen if there is a dependency on a missing item
			depStrings[i] = fmt.Sprintf("  MISSING: %s", resourceName)
		}
	}

	return helpers.Tabify(strings.Join(depStrings, "\n"), "   ")
}
