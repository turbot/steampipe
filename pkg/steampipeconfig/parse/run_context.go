package parse

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/stevenle/topsort"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/hclhelpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/versionmap"
	"github.com/zclconf/go-cty/cty"
)

const rootDependencyNode = "rootDependencyNode"

type ParseModFlag uint32

const (
	CreateDefaultMod ParseModFlag = 1 << iota
	CreatePseudoResources
)

/*
	ReferenceTypeValueMap is the raw data used to build the evaluation context

When resolving hcl references like :
- query.q1
- var.v1
- mod1.query.my_query.sql

ReferenceTypeValueMap is keyed  by resource type, then by resource name
*/
type ReferenceTypeValueMap map[string]map[string]cty.Value

type RunContext struct {
	// the mod which is currently being parsed
	CurrentMod *modconfig.Mod
	// the workspace lock data
	WorkspaceLock    *versionmap.WorkspaceLock
	UnresolvedBlocks map[string]*unresolvedBlock
	FileData         map[string][]byte
	// the eval context used to decode references in HCL
	EvalCtx *hcl.EvalContext

	Flags                ParseModFlag
	ListOptions          *filehelpers.ListOptions
	LoadedDependencyMods modconfig.ModMap
	RootEvalPath         string
	// if set, only decode these blocks
	BlockTypes []string
	// if set, exclude these block types
	BlockTypeExclusions []string

	// Variables are populated in an initial parse pass top we store them on the run context
	// so we can set them on the mod when we do the main parse

	// Variables is a map of the variables in the current mod
	// it is used to populate the variables property on the mod
	Variables map[string]*modconfig.Variable

	// DependencyVariables is a map of the variables in the dependency mods of the current mod
	// it is used to populate the variables property on the dependency
	DependencyVariables map[string]map[string]*modconfig.Variable
	ParentRunCtx        *RunContext

	// stack of parent resources for the currently parsed block
	// (unqualified name)
	parents []string

	// map of resource children, keyed by parent unqualified name
	blockChildMap   map[string][]string
	dependencyGraph *topsort.Graph
	// map of ReferenceTypeValueMaps keyed by mod
	// NOTE: all values from root mod are keyed with "local"
	referenceValues map[string]ReferenceTypeValueMap
	blocks          hcl.Blocks
	// map of top  level blocks, for easy checking
	topLevelBlocks map[*hcl.Block]struct{}
	// map of block names, keyed by a hash of the blopck
	blockNameMap map[string]string
}

func NewRunContext(workspaceLock *versionmap.WorkspaceLock, rootEvalPath string, flags ParseModFlag, listOptions *filehelpers.ListOptions) *RunContext {
	c := &RunContext{
		Flags:                flags,
		RootEvalPath:         rootEvalPath,
		WorkspaceLock:        workspaceLock,
		ListOptions:          listOptions,
		LoadedDependencyMods: make(modconfig.ModMap),
		UnresolvedBlocks:     make(map[string]*unresolvedBlock),
		referenceValues: map[string]ReferenceTypeValueMap{
			"local": make(ReferenceTypeValueMap),
		},
		blockChildMap: make(map[string][]string),
		blockNameMap:  make(map[string]string),
		// initialise variable maps - even though we later overwrite them
		Variables: make(map[string]*modconfig.Variable),
	}
	// add root node - this will depend on all other nodes
	c.dependencyGraph = c.newDependencyGraph()
	c.buildEvalContext()

	return c
}

func (r *RunContext) EnsureWorkspaceLock(mod *modconfig.Mod) error {
	// if the mod has dependencies, there must a workspace lock object in the run context
	// (mod MUST be the workspace mod, not a dependency, as we would hit this error as soon as we parse it)
	if mod.HasDependentMods() && (r.WorkspaceLock.Empty() || r.WorkspaceLock.Incomplete()) {
		return fmt.Errorf("not all dependencies are installed - run 'steampipe mod install'")
	}

	return nil
}

func (r *RunContext) PushParent(parent modconfig.ModTreeItem) {
	r.parents = append(r.parents, parent.GetUnqualifiedName())
}

func (r *RunContext) PopParent() string {
	n := len(r.parents) - 1
	res := r.parents[n]
	r.parents = r.parents[:n]
	return res
}

func (r *RunContext) PeekParent() string {
	if len(r.parents) == 0 {
		return r.CurrentMod.Name()
	}
	return r.parents[len(r.parents)-1]
}

// VariableValueMap converts a map of variables to a map of the underlying cty value
func VariableValueMap(variables map[string]*modconfig.Variable) map[string]cty.Value {
	ret := make(map[string]cty.Value, len(variables))
	for k, v := range variables {
		ret[k] = v.Value
	}
	return ret
}

// AddInputVariables adds variables to the run context.
// This function is called for the root run context after loading all input variables
func (r *RunContext) AddInputVariables(inputVariables *modconfig.ModVariableMap) {
	r.setRootVariables(inputVariables.RootVariables)
	r.setDependencyVariables(inputVariables.DependencyVariables)
}

// SetVariablesForDependencyMod adds variables to the run context.
// This function is called for dependent mod run context
func (r *RunContext) SetVariablesForDependencyMod(mod *modconfig.Mod, dependencyVariablesMap map[string]map[string]*modconfig.Variable) {
	r.setRootVariables(dependencyVariablesMap[mod.ShortName])
	r.setDependencyVariables(dependencyVariablesMap)
}

// setRootVariables sets the Variables property
// and adds the variables to the referenceValues map (used to build the eval context)
func (r *RunContext) setRootVariables(variables map[string]*modconfig.Variable) {
	r.Variables = variables
	// NOTE: we add with the name "var" not "variable" as that is how variables are referenced
	r.referenceValues["local"]["var"] = VariableValueMap(variables)
}

// setDependencyVariables sets the DependencyVariables property
// and adds the dependency variables to the referenceValues map (used to build the eval context
func (r *RunContext) setDependencyVariables(dependencyVariables map[string]map[string]*modconfig.Variable) {
	r.DependencyVariables = dependencyVariables
	// NOTE: we add with the name "var" not "variable" as that is how variables are referenced
	// add top level variables
	// add dependency mod variables, scoped by mod name
	for depModName, depVars := range r.DependencyVariables {
		// create map for this dependency if needed
		if r.referenceValues[depModName] == nil {
			r.referenceValues[depModName] = make(ReferenceTypeValueMap)
		}
		r.referenceValues[depModName]["var"] = VariableValueMap(depVars)
	}
}

// AddMod is used to add a mod with any pseudo resources to the eval context
// - in practice this will be a shell mod with just pseudo resources - other resources will be added as they are parsed
func (r *RunContext) AddMod(mod *modconfig.Mod) hcl.Diagnostics {
	if len(r.UnresolvedBlocks) > 0 {
		// should never happen
		panic("calling SetContent on runContext but there are unresolved blocks from a previous parse")
	}

	var diags hcl.Diagnostics

	moreDiags := r.storeResourceInCtyMap(mod)
	diags = append(diags, moreDiags...)

	resourceFunc := func(item modconfig.HclResource) (bool, error) {
		// add all mod resources except variables into cty map
		if _, ok := item.(*modconfig.Variable); !ok {
			moreDiags := r.storeResourceInCtyMap(item)
			diags = append(diags, moreDiags...)
		}
		// continue walking
		return true, nil
	}
	mod.WalkResources(resourceFunc)

	// rebuild the eval context
	r.buildEvalContext()
	return diags
}

func (r *RunContext) SetDecodeContent(content *hcl.BodyContent, fileData map[string][]byte) {
	r.blocks = content.Blocks
	// put blocks into map as well
	r.topLevelBlocks = make(map[*hcl.Block]struct{}, len(r.blocks))
	for _, b := range content.Blocks {
		r.topLevelBlocks[b] = struct{}{}
	}
	r.FileData = fileData
}

func (r *RunContext) ShouldIncludeBlock(block *hcl.Block) bool {
	if len(r.BlockTypes) > 0 && !helpers.StringSliceContains(r.BlockTypes, block.Type) {
		return false
	}
	if len(r.BlockTypeExclusions) > 0 && helpers.StringSliceContains(r.BlockTypeExclusions, block.Type) {
		return false
	}
	return true
}

func (r *RunContext) ClearDependencies() {
	r.UnresolvedBlocks = make(map[string]*unresolvedBlock)
	r.dependencyGraph = r.newDependencyGraph()
}

// AddDependencies :: the block could not be resolved as it has dependencies
// 1) store block as unresolved
// 2) add dependencies to our tree of dependencies
func (r *RunContext) AddDependencies(block *hcl.Block, name string, dependencies map[string]*modconfig.ResourceDependency) hcl.Diagnostics {
	var diags hcl.Diagnostics
	// store unresolved block
	r.UnresolvedBlocks[name] = &unresolvedBlock{Name: name, Block: block, Dependencies: dependencies}

	// store dependency in tree - d
	if !r.dependencyGraph.ContainsNode(name) {
		r.dependencyGraph.AddNode(name)
	}
	// add root dependency
	if err := r.dependencyGraph.AddEdge(rootDependencyNode, name); err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to add root dependency to graph",
			Detail:   err.Error()})
	}

	for _, dep := range dependencies {
		// each dependency object may have multiple traversals
		for _, t := range dep.Traversals {
			parsedPropertyPath, err := modconfig.ParseResourcePropertyPath(hclhelpers.TraversalAsString(t))

			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "failed to parse dependency",
					Detail:   err.Error()})
				continue

			}

			// 'd' may be a property path - when storing dependencies we only care about the resource names
			dependencyResourceName := parsedPropertyPath.ToResourceName()
			if !r.dependencyGraph.ContainsNode(dependencyResourceName) {
				r.dependencyGraph.AddNode(dependencyResourceName)
			}
			if err := r.dependencyGraph.AddEdge(name, dependencyResourceName); err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "failed to add dependency to graph",
					Detail:   err.Error()})
			}
		}
	}
	return diags
}

// BlocksToDecode builds a list of blocks to decode, the order of which is determined by the depdnency order
func (r *RunContext) BlocksToDecode() (hcl.Blocks, error) {
	depOrder, err := r.getDependencyOrder()
	if err != nil {
		return nil, err
	}
	if len(depOrder) == 0 {
		return r.blocks, nil
	}

	// NOTE: a block may appear more than once in unresolved blocks
	// if it defines multiple unresolved resources, e.g a locals block

	// make a map of blocks we have already included, keyed by the block def range
	blocksMap := make(map[string]bool)
	var blocksToDecode hcl.Blocks
	for _, name := range depOrder {
		// depOrder is all the blocks required to resolve dependencies.
		// if this one is unparsed, added to list
		block, ok := r.UnresolvedBlocks[name]
		if ok && !blocksMap[block.Block.DefRange.String()] {
			blocksToDecode = append(blocksToDecode, block.Block)
			// add to map
			blocksMap[block.Block.DefRange.String()] = true
		}
	}
	return blocksToDecode, nil
}

// EvalComplete returns whether all elements in the dependency tree fully evaluated
func (r *RunContext) EvalComplete() bool {
	return len(r.UnresolvedBlocks) == 0
}

// ShouldCreateDefaultMod returns whether the flag is set to create a default mod if no mod definition exists
func (r *RunContext) ShouldCreateDefaultMod() bool {
	return r.Flags&CreateDefaultMod == CreateDefaultMod
}

// CreatePseudoResources returns whether the flag is set to create pseudo resources
func (r *RunContext) CreatePseudoResources() bool {
	return r.Flags&CreatePseudoResources == CreatePseudoResources
}

// AddResource stores this resource as a variable to be added to the eval context. It alse
func (r *RunContext) AddResource(resource modconfig.HclResource) hcl.Diagnostics {
	diagnostics := r.storeResourceInCtyMap(resource)
	if diagnostics.HasErrors() {
		return diagnostics
	}

	// rebuild the eval context
	r.buildEvalContext()

	return nil
}

func (r *RunContext) FormatDependencies() string {
	// first get the dependency order
	dependencyOrder, err := r.getDependencyOrder()
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
		dep, ok := r.UnresolvedBlocks[resourceName]

		if ok {
			depStrings[i] = dep.String()
		} else {
			// this could happen if there is a dependency on a missing item
			depStrings[i] = fmt.Sprintf("  MISSING: %s", resourceName)
		}
	}

	return helpers.Tabify(strings.Join(depStrings, "\n"), "   ")
}

func (r *RunContext) GetMod(modShortName string) *modconfig.Mod {
	if modShortName == r.CurrentMod.ShortName {
		return r.CurrentMod
	}
	// we need to iterate through dependency mods - we cannot use modShortNameas key as it is short name
	for _, dep := range r.LoadedDependencyMods {
		if dep.ShortName == modShortName {
			return dep
		}
	}
	return nil
}

func (r *RunContext) GetResourceMaps() *modconfig.ResourceMaps {
	dependencyResourceMaps := make([]*modconfig.ResourceMaps, len(r.LoadedDependencyMods))
	idx := 0
	// use the current mod as the base resource map
	resourceMap := r.CurrentMod.GetResourceMaps()

	// merge in the dependency mods
	for _, m := range r.LoadedDependencyMods {
		dependencyResourceMaps[idx] = m.GetResourceMaps()
		idx++
	}

	resourceMap = resourceMap.Merge(dependencyResourceMaps)
	return resourceMap
}

func (r *RunContext) newDependencyGraph() *topsort.Graph {
	dependencyGraph := topsort.NewGraph()
	// add root node - this will depend on all other nodes
	dependencyGraph.AddNode(rootDependencyNode)
	return dependencyGraph
}

// return the optimal run order required to resolve dependencies
func (r *RunContext) getDependencyOrder() ([]string, error) {
	rawDeps, err := r.dependencyGraph.TopSort(rootDependencyNode)
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
func (r *RunContext) buildEvalContext() {
	// convert variables to cty values
	variables := make(map[string]cty.Value)

	// now for each mod add all the values
	for mod, modMap := range r.referenceValues {
		if mod == "local" {
			for k, v := range modMap {
				variables[k] = cty.ObjectVal(v)
			}
			continue
		}

		// mod map is map[string]map[string]cty.Value
		// for each element (i.e. map[string]cty.Value) convert to cty object
		refTypeMap := make(map[string]cty.Value)
		for refType, typeValueMap := range modMap {
			refTypeMap[refType] = cty.ObjectVal(typeValueMap)
		}
		// now convert the cty map to a cty object
		variables[mod] = cty.ObjectVal(refTypeMap)
	}

	// create evaluation context
	r.EvalCtx = &hcl.EvalContext{
		Variables: variables,
		// use the mod path as the file root for functions
		Functions: ContextFunctions(r.RootEvalPath),
	}
}

// update the cached cty value for the given resource, as long as itr does not already exist
func (r *RunContext) storeResourceInCtyMap(resource modconfig.HclResource) hcl.Diagnostics {
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
	if diags := r.addReferenceValue(resource, ctyValue); diags.HasErrors() {
		return diags
	}

	// remove this resource from unparsed blocks
	delete(r.UnresolvedBlocks, resource.Name())

	return nil
}

func (r *RunContext) addReferenceValue(resource modconfig.HclResource, value cty.Value) hcl.Diagnostics {
	parsedName, err := modconfig.ParseResourceName(resource.Name())
	if err != nil {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("failed to parse resource name %s", resource.Name()),
			Detail:   err.Error(),
			Subject:  resource.GetDeclRange(),
		}}
	}

	// TODO validate mod name clashes
	// TODO mod reserved names
	// TODO handle aliases

	key := parsedName.Name
	typeString := parsedName.ItemType

	// the resource name will not have a mod - but the run context knows which mod we are parsing
	mod := r.CurrentMod
	modName := mod.ShortName
	if mod.ModPath == r.RootEvalPath {
		modName = "local"
	}
	variablesForMod, ok := r.referenceValues[modName]
	// do we have a map of reference values for this dep mod?
	if !ok {
		// no - create one
		variablesForMod = make(ReferenceTypeValueMap)
		r.referenceValues[modName] = variablesForMod
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
		r.referenceValues[modName] = variablesForMod
	}

	return nil
}

func (r *RunContext) AddLoadedDependentMods(mods modconfig.ModMap) {
	for k, v := range mods {
		if _, alreadyLoaded := r.LoadedDependencyMods[k]; !alreadyLoaded {
			r.LoadedDependencyMods[k] = v
		}
	}
}

func (r *RunContext) IsTopLevelBlock(block *hcl.Block) bool {
	_, isTopLevel := r.topLevelBlocks[block]
	return isTopLevel
}
