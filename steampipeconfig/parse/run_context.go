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

// ReferenceTypeValueMap is a map of reference value maps, keyed by type
type ReferenceTypeValueMap map[string]map[string]cty.Value

type RunContext struct {
	// we only store the root mod so we can tell whether a given mod should be treated as "local"
	RootMod          *modconfig.Mod
	CurrentMod       *modconfig.Mod
	UnresolvedBlocks map[string]*unresolvedBlock
	FileData         map[string][]byte
	dependencyGraph  *topsort.Graph
	// map of ReferenceTypeValueMaps keyed by mod
	// NOTE: all values from root mod are keyed with "local"
	referenceValues map[string]ReferenceTypeValueMap

	EvalCtx *hcl.EvalContext
	blocks  hcl.Blocks
}

func NewRunContext() *RunContext {
	c := &RunContext{
		UnresolvedBlocks: make(map[string]*unresolvedBlock),
		referenceValues: map[string]ReferenceTypeValueMap{
			"local": make(ReferenceTypeValueMap),
		},
	}
	// add root node - this will depend on all other nodes
	c.dependencyGraph = c.newDependencyGraph()

	// add enums to the variables which may be referenced from within the hcl
	c.addSteampipeEnums()

	return c
}

func (c *RunContext) AddVariables(inputVariables map[string]cty.Value) {
	// NOTE: we add with the name "var" not "variable" as that is how variables are referenced
	c.referenceValues["local"]["var"] = inputVariables
}

// AddMod is used to add a mod with any pseudo resources to the eval context
// - in practice this will be a shell mod with just pseudo resources - other resources will be added as they are parsed
func (c *RunContext) AddMod(mod *modconfig.Mod, content *hcl.BodyContent, fileData map[string][]byte) hcl.Diagnostics {
	if len(c.UnresolvedBlocks) > 0 {
		// should never happen
		panic("calling SetContent on runContext but there are unresolved blocks from a previous parse")
	}
	// if root mod is not set, set it now
	// NOTE this will only be called for LoadVariables which does not set RootMod
	if c.RootMod == nil {
		c.RootMod = mod
	}
	c.CurrentMod = mod
	c.FileData = fileData
	c.blocks = content.Blocks

	var diags hcl.Diagnostics

	moreDiags := c.storeResourceInCtyMap(mod)
	diags = append(diags, moreDiags...)
	// add mod resources
	for _, q := range mod.Queries {
		moreDiags := c.storeResourceInCtyMap(q)
		diags = append(diags, moreDiags...)
	}
	for _, q := range mod.Controls {
		moreDiags := c.storeResourceInCtyMap(q)
		diags = append(diags, moreDiags...)
	}
	for _, q := range mod.Locals {
		moreDiags := c.storeResourceInCtyMap(q)
		diags = append(diags, moreDiags...)
	}
	for _, q := range mod.Reports {
		moreDiags := c.storeResourceInCtyMap(q)
		diags = append(diags, moreDiags...)
	}
	for _, q := range mod.Panels {
		moreDiags := c.storeResourceInCtyMap(q)
		diags = append(diags, moreDiags...)
	}

	// rebuild the eval context from the ctyMap
	c.EvalCtx = c.ctyMapToEvalContext()
	return diags
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

// BlocksToDecode builds a list of blocks to decode, the order of which is determined by the depdnency order
func (c *RunContext) BlocksToDecode() (hcl.Blocks, error) {
	depOrder, err := c.getDependencyOrder()
	if err != nil {
		return nil, err
	}
	if len(depOrder) == 0 {
		return c.blocks, nil
	}

	// NOTE: a block may appear more than once in unresolved blocks
	// if it defines muleiple unresolved resources, e.g a locals block

	// make a map of blocks we have already included, keyed by the block def range
	blocksMap := make(map[string]bool)
	var blocksToDecode hcl.Blocks
	for _, name := range depOrder {
		// depOrder is all the blocks required to resolve dependencies.
		// if this one is unparsed, added to list
		block, ok := c.UnresolvedBlocks[name]
		if ok && !blocksMap[block.Block.DefRange.String()] {
			blocksToDecode = append(blocksToDecode, block.Block)
			// add to map
			blocksMap[block.Block.DefRange.String()] = true
		}
	}
	return blocksToDecode, nil
}

// EvalComplete :: Are all elements in the dependency tree fully evaluated
func (c *RunContext) EvalComplete() bool {
	return len(c.UnresolvedBlocks) == 0
}

// add enums to the referenceValues which may be referenced from within the hcl
func (c *RunContext) addSteampipeEnums() {
	c.referenceValues["local"]["steampipe"] = map[string]cty.Value{
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
			"graph":            cty.StringVal("steampipe.panel.graph"),
			"sankey_diagram":   cty.StringVal("steampipe.panel.sankey_diagram"),
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

// return the optimal run order required to resolve dependencies
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

// eval functions
func (c *RunContext) ctyMapToEvalContext() *hcl.EvalContext {
	// convert variables to cty values
	variables := make(map[string]cty.Value)

	// first add local values
	for k, v := range c.referenceValues["local"] {
		variables[k] = cty.ObjectVal(v)
	}
	// now for each mod add all the values
	for mod, modMap := range c.referenceValues {
		// mod map is map[string]map[string]cty.Value
		// for each element (i.e. map[string]cty.Value) convert to cty object
		refTypeMap := make(map[string]cty.Value)
		for refType, typeValueMap := range modMap {
			refTypeMap[refType] = cty.ObjectVal(typeValueMap)
		}
		// now convert the cty map to a cty object
		variables[mod] = cty.ObjectVal(refTypeMap)
	}

	//create evaluation context
	return &hcl.EvalContext{
		Variables: variables,
		Functions: ContextFunctions(c.RootMod.FilePath),
	}
}

// AddResource stores this resource as a variable to be added to the eval context. It alse
func (c *RunContext) AddResource(resource modconfig.HclResource) hcl.Diagnostics {
	diagnostics := c.storeResourceInCtyMap(resource)
	if diagnostics.HasErrors() {
		return diagnostics
	}

	// rebuild the eval context
	c.EvalCtx = c.ctyMapToEvalContext()

	return nil
}

// update the cached cty value for the given resource, as long as itr does not already exist
func (c *RunContext) storeResourceInCtyMap(resource modconfig.HclResource) hcl.Diagnostics {
	// add resource to variable map
	ctyValue, err := resource.CtyValue()
	if err != nil {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("failed to convert resource '%s'to its cty value", resource.Name()),
			Detail:   err.Error(),
			Subject:  resource.GetDeclRange(),
		}}
	}

	// add into the reference value map
	if diags := c.addReferenceValue(resource, ctyValue); diags.HasErrors() {
		return diags
	}

	// remove this resource from unparsed blocks
	if _, ok := c.UnresolvedBlocks[resource.Name()]; ok {
		delete(c.UnresolvedBlocks, resource.Name())
	}
	return nil
}

func (c *RunContext) addReferenceValue(resource modconfig.HclResource, value cty.Value) hcl.Diagnostics {
	parsedName, err := modconfig.ParseResourceName(resource.Name())
	if err != nil {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("failed to parse resource name %s", resource.Name()),
			Detail:   err.Error(),
			Subject:  resource.GetDeclRange(),
		}}
	}

	// TODO VALIDATE MOD NAME CLASHES?
	// TODO MOD RESERVED NAMES
	// TODO HANDLE ALIASES

	key := parsedName.Name
	typeString := parsedName.ItemType

	// the resource name will not have a mod - but the run context knows which mod we are parsing

	mod := c.CurrentMod

	modName := mod.ShortName
	if mod.FilePath == c.RootMod.FilePath {
		modName = "local"
	}
	variablesForMod, ok := c.referenceValues[modName]
	// do we have a map of reference values for this dep mod?
	if !ok {
		// no - create one
		variablesForMod = make(ReferenceTypeValueMap)
		c.referenceValues[modName] = variablesForMod
	}
	// do we have a map of reference values for this type
	variablesForType, ok := variablesForMod[typeString]
	if !ok {
		// no - create one
		variablesForType = make(map[string]cty.Value)
	}

	// DO NOT update the cached cty values if the value already exists
	// this can happen in the case of variables where we initialise the context with values read from file
	// or passed on the command line,	// does this item exist in the map
	if _, ok := variablesForType[key]; !ok {
		variablesForType[key] = value
		variablesForMod[typeString] = variablesForType
		c.referenceValues[modName] = variablesForMod
	}

	return nil
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
