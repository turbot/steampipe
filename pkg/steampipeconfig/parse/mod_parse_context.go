package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	filehelpers "github.com/turbot/go-kit/files"
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

type ModParseContext struct {
	ParseContext
	// the mod which is currently being parsed
	CurrentMod *modconfig.Mod
	// the workspace lock data
	WorkspaceLock *versionmap.WorkspaceLock

	Flags                ParseModFlag
	ListOptions          *filehelpers.ListOptions
	LoadedDependencyMods modconfig.ModMap

	// Variables are populated in an initial parse pass top we store them on the run context
	// so we can set them on the mod when we do the main parse

	// Variables is a map of the variables in the current mod
	// it is used to populate the variables property on the mod
	Variables map[string]*modconfig.Variable

	// DependencyVariables is a map of the variables in the dependency mods of the current mod
	// it is used to populate the variables property on the dependency
	DependencyVariables map[string]map[string]*modconfig.Variable
	ParentParseCtx      *ModParseContext

	// stack of parent resources for the currently parsed block
	// (unqualified name)
	parents []string

	// map of resource children, keyed by parent unqualified name
	blockChildMap map[string][]string

	// map of top  level blocks, for easy checking
	topLevelBlocks map[*hcl.Block]struct{}
	// map of block names, keyed by a hash of the blopck
	blockNameMap map[string]string
	// map of ReferenceTypeValueMaps keyed by mod
	// NOTE: all values from root mod are keyed with "local"
	referenceValues map[string]ReferenceTypeValueMap
}

func NewModParseContext(workspaceLock *versionmap.WorkspaceLock, rootEvalPath string, flags ParseModFlag, listOptions *filehelpers.ListOptions) *ModParseContext {
	parseContext := NewParseContext(rootEvalPath)
	c := &ModParseContext{
		ParseContext:         parseContext,
		Flags:                flags,
		WorkspaceLock:        workspaceLock,
		ListOptions:          listOptions,
		LoadedDependencyMods: make(modconfig.ModMap),

		blockChildMap: make(map[string][]string),
		blockNameMap:  make(map[string]string),
		// initialise variable maps - even though we later overwrite them
		Variables: make(map[string]*modconfig.Variable),
		referenceValues: map[string]ReferenceTypeValueMap{
			"local": make(ReferenceTypeValueMap),
		},
	}
	// add root node - this will depend on all other nodes
	c.dependencyGraph = c.newDependencyGraph()
	c.buildEvalContext()

	return c
}

func (r *ModParseContext) EnsureWorkspaceLock(mod *modconfig.Mod) error {
	// if the mod has dependencies, there must a workspace lock object in the run context
	// (mod MUST be the workspace mod, not a dependency, as we would hit this error as soon as we parse it)
	if mod.HasDependentMods() && (r.WorkspaceLock.Empty() || r.WorkspaceLock.Incomplete()) {
		return fmt.Errorf("not all dependencies are installed - run 'steampipe mod install'")
	}

	return nil
}

func (r *ModParseContext) PushParent(parent modconfig.ModTreeItem) {
	r.parents = append(r.parents, parent.GetUnqualifiedName())
}

func (r *ModParseContext) PopParent() string {
	n := len(r.parents) - 1
	res := r.parents[n]
	r.parents = r.parents[:n]
	return res
}

func (r *ModParseContext) PeekParent() string {
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
func (r *ModParseContext) AddInputVariables(inputVariables *modconfig.ModVariableMap) {
	r.setRootVariables(inputVariables.RootVariables)
	r.setDependencyVariables(inputVariables.DependencyVariables)
}

// SetVariablesForDependencyMod adds variables to the run context.
// This function is called for dependent mod run context
func (r *ModParseContext) SetVariablesForDependencyMod(mod *modconfig.Mod, dependencyVariablesMap map[string]map[string]*modconfig.Variable) {
	r.setRootVariables(dependencyVariablesMap[mod.ShortName])
	r.setDependencyVariables(dependencyVariablesMap)
}

// setRootVariables sets the Variables property
// and adds the variables to the referenceValues map (used to build the eval context)
func (r *ModParseContext) setRootVariables(variables map[string]*modconfig.Variable) {
	r.Variables = variables
	// NOTE: we add with the name "var" not "variable" as that is how variables are referenced
	r.referenceValues["local"]["var"] = VariableValueMap(variables)
}

// setDependencyVariables sets the DependencyVariables property
// and adds the dependency variables to the referenceValues map (used to build the eval context
func (r *ModParseContext) setDependencyVariables(dependencyVariables map[string]map[string]*modconfig.Variable) {
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
func (r *ModParseContext) AddMod(mod *modconfig.Mod) hcl.Diagnostics {
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

func (r *ModParseContext) SetDecodeContent(content *hcl.BodyContent, fileData map[string][]byte) {
	// put blocks into map as well
	r.topLevelBlocks = make(map[*hcl.Block]struct{}, len(r.blocks))
	for _, b := range content.Blocks {
		r.topLevelBlocks[b] = struct{}{}
	}
	r.ParseContext.SetDecodeContent(content, fileData)
}

// AddDependencies :: the block could not be resolved as it has dependencies
// 1) store block as unresolved
// 2) add dependencies to our tree of dependencies
func (r *ModParseContext) AddDependencies(block *hcl.Block, name string, dependencies map[string]*modconfig.ResourceDependency) hcl.Diagnostics {
	// TACTICAL if this is NOT a top level block, add a suffix to the block name
	// this is needed to avoid circular dependency errors if a nested block references
	// a top level block with the same name
	if !r.IsTopLevelBlock(block) {
		name = "nested." + name
	}
	return r.ParseContext.AddDependencies(block, name, dependencies)
}

// ShouldCreateDefaultMod returns whether the flag is set to create a default mod if no mod definition exists
func (r *ModParseContext) ShouldCreateDefaultMod() bool {
	return r.Flags&CreateDefaultMod == CreateDefaultMod
}

// CreatePseudoResources returns whether the flag is set to create pseudo resources
func (r *ModParseContext) CreatePseudoResources() bool {
	return r.Flags&CreatePseudoResources == CreatePseudoResources
}

// AddResource stores this resource as a variable to be added to the eval context. It alse
func (r *ModParseContext) AddResource(resource modconfig.HclResource) hcl.Diagnostics {
	diagnostics := r.storeResourceInCtyMap(resource)
	if diagnostics.HasErrors() {
		return diagnostics
	}

	// rebuild the eval context
	r.buildEvalContext()

	return nil
}

func (r *ModParseContext) GetMod(modShortName string) *modconfig.Mod {
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

func (r *ModParseContext) GetResourceMaps() *modconfig.ResourceMaps {
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

// eval functions
func (r *ModParseContext) buildEvalContext() {
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

	r.ParseContext.buildEvalContext(variables)
}

// update the cached cty value for the given resource, as long as itr does not already exist
func (r *ModParseContext) storeResourceInCtyMap(resource modconfig.HclResource) hcl.Diagnostics {
	// add resource to variable map
	ctyValue, err := resource.CtyValue()
	if err != nil {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("failed to convert resource '%s' to its cty value", resource.Name()),
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

func (r *ModParseContext) addReferenceValue(resource modconfig.HclResource, value cty.Value) hcl.Diagnostics {
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

func (r *ModParseContext) AddLoadedDependentMods(mods modconfig.ModMap) {
	for k, v := range mods {
		if _, alreadyLoaded := r.LoadedDependencyMods[k]; !alreadyLoaded {
			r.LoadedDependencyMods[k] = v
		}
	}
}

func (r *ModParseContext) IsTopLevelBlock(block *hcl.Block) bool {
	_, isTopLevel := r.topLevelBlocks[block]
	return isTopLevel
}
