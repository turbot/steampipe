package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/stevenle/topsort"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

const rootDependencyNode = "rootDependencyNode"

type RunContext struct {
	Mod              *modconfig.Mod
	UnresolvedBlocks map[string]*hcl.Block
	FileData         map[string][]byte
	dependencyGraph  *topsort.Graph
	//dependencies     map[string]bool
	// store any objects which are depdnecy targets
	variables map[string]cty.Value
	EvalCtx   *hcl.EvalContext
	blocks    hcl.Blocks
}

func NewRunContext(mod *modconfig.Mod, content *hcl.BodyContent, fileData map[string][]byte) (*RunContext, hcl.Diagnostics) {
	c := &RunContext{
		Mod:              mod,
		FileData:         fileData,
		UnresolvedBlocks: make(map[string]*hcl.Block),
		dependencyGraph:  topsort.NewGraph(),
		//dependencies:     make(map[string]bool),
		variables: make(map[string]cty.Value),
		blocks:    content.Blocks,
	}
	// add root node - this will depend on all other nodes
	c.dependencyGraph.AddNode(rootDependencyNode)
	c.buildEvalContext()

	// add mod and any existing mod resources to the variables
	diags := c.addModToVariables()

	return c, diags
}

// AddDependencies :: the block could not be resolved as it has dependencies
// 1) store block as unresolved
// 2) add dependencies to our tree of depdnecie
func (c *RunContext) AddDependencies(block *hcl.Block, name string, dependencies []hcl.Traversal) hcl.Diagnostics {
	var diags hcl.Diagnostics
	// store unresolved block
	c.UnresolvedBlocks[name] = block

	// store dependency in tree - d
	if !c.dependencyGraph.ContainsNode(name) {
		c.dependencyGraph.AddNode(name)
	}
	// add root dependency
	c.dependencyGraph.AddEdge(rootDependencyNode, name)

	for _, dep := range dependencies {
		d := TraversalAsString(dep)

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

		// store the raw dependency properties
		//c.dependencies[d] = true
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
			blocksToDecode = append(blocksToDecode, block)
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
	// create evaluation context
	c.EvalCtx = &hcl.EvalContext{
		Variables: c.variables,
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
	if !c.Mod.AddResource(resource) {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("mod defines more that one query named %s", resource.Name()),
			Subject:  &block.DefRange,
		})
	}

	return diagnostics

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

	var goMap map[string]cty.Value
	typeString := parsedName.TypeString()
	ctyMap, ok := c.variables[typeString]
	if ok {
		gocty.FromCtyValue(ctyMap, &goMap)
	} else {
		goMap = make(map[string]cty.Value)
	}
	goMap[parsedName.Name] = ctyValue
	c.variables[typeString] = cty.MapVal(goMap)

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
