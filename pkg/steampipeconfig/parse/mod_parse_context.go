package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/versionmap"
	"github.com/zclconf/go-cty/cty"
	"runtime/debug"
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

	Flags       ParseModFlag
	ListOptions *filehelpers.ListOptions
	// map of loaded dependency mods, keyed by DependencyPath (including version)
	// there may be multiple versions of same mod in this map
	loadedDependencyMods modconfig.ModMap

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
	// a map of just the top level dependencies of the CurrentMod, keyed my full mod DepdencyName (with no version)
	topLevelDependencyMods modconfig.ModMap
}

func NewModParseContext(workspaceLock *versionmap.WorkspaceLock, rootEvalPath string, flags ParseModFlag, listOptions *filehelpers.ListOptions) *ModParseContext {
	parseContext := NewParseContext(rootEvalPath)
	c := &ModParseContext{
		ParseContext:         parseContext,
		Flags:                flags,
		WorkspaceLock:        workspaceLock,
		ListOptions:          listOptions,
		loadedDependencyMods: make(modconfig.ModMap),

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

func (m *ModParseContext) EnsureWorkspaceLock(mod *modconfig.Mod) error {
	// if the mod has dependencies, there must a workspace lock object in the run context
	// (mod MUST be the workspace mod, not a dependency, as we would hit this error as soon as we parse it)
	if mod.HasDependentMods() && (m.WorkspaceLock.Empty() || m.WorkspaceLock.Incomplete()) {
		return fmt.Errorf("not all dependencies are installed - run 'steampipe mod install'")
	}

	return nil
}

func (m *ModParseContext) PushParent(parent modconfig.ModTreeItem) {
	m.parents = append(m.parents, parent.GetUnqualifiedName())
}

func (m *ModParseContext) PopParent() string {
	n := len(m.parents) - 1
	res := m.parents[n]
	m.parents = m.parents[:n]
	return res
}

func (m *ModParseContext) PeekParent() string {
	if len(m.parents) == 0 {
		return m.CurrentMod.Name()
	}
	return m.parents[len(m.parents)-1]
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
func (m *ModParseContext) AddInputVariables(inputVariables *modconfig.ModVariableMap) {
	m.setRootVariables(inputVariables.RootVariables)
	m.setDependencyVariables(inputVariables.DependencyVariables)
}

// SetVariablesForDependencyMod adds variables to the run context.
// This function is called for dependent mod run context
func (m *ModParseContext) SetVariablesForDependencyMod(mod *modconfig.Mod, dependencyVariablesMap map[string]map[string]*modconfig.Variable) {
	m.setRootVariables(dependencyVariablesMap[mod.ShortName])
	m.setDependencyVariables(dependencyVariablesMap)
}

// setRootVariables sets the Variables property
// and adds the variables to the referenceValues map (used to build the eval context)
func (m *ModParseContext) setRootVariables(variables map[string]*modconfig.Variable) {
	m.Variables = variables
	// NOTE: we add with the name "var" not "variable" as that is how variables are referenced
	m.referenceValues["local"]["var"] = VariableValueMap(variables)
}

// setDependencyVariables sets the DependencyVariables property
// and adds the dependency variables to the referenceValues map (used to build the eval context
func (m *ModParseContext) setDependencyVariables(dependencyVariables map[string]map[string]*modconfig.Variable) {
	m.DependencyVariables = dependencyVariables
	// NOTE: we add with the name "var" not "variable" as that is how variables are referenced
	// add top level variables
	// add dependency mod variables, scoped by mod name
	for depModName, depVars := range m.DependencyVariables {
		// create map for this dependency if needed
		if m.referenceValues[depModName] == nil {
			m.referenceValues[depModName] = make(ReferenceTypeValueMap)
		}
		m.referenceValues[depModName]["var"] = VariableValueMap(depVars)
	}
}

// AddMod is used to add a mod to the eval context
func (m *ModParseContext) AddMod(mod *modconfig.Mod) hcl.Diagnostics {
	if len(m.UnresolvedBlocks) > 0 {
		// should never happen
		panic("calling SetContent on runContext but there are unresolved blocks from a previous parse")
	}

	var diags hcl.Diagnostics

	moreDiags := m.storeResourceInCtyMap(mod)
	diags = append(diags, moreDiags...)

	resourceFunc := func(item modconfig.HclResource) (bool, error) {
		// add all mod resources except variables into cty map
		if _, ok := item.(*modconfig.Variable); !ok {
			moreDiags := m.storeResourceInCtyMap(item)
			diags = append(diags, moreDiags...)
		}
		// continue walking
		return true, nil
	}
	mod.WalkResources(resourceFunc)

	// rebuild the eval context
	m.buildEvalContext()
	return diags
}

func (m *ModParseContext) SetDecodeContent(content *hcl.BodyContent, fileData map[string][]byte) {
	// put blocks into map as well
	m.topLevelBlocks = make(map[*hcl.Block]struct{}, len(m.blocks))
	for _, b := range content.Blocks {
		m.topLevelBlocks[b] = struct{}{}
	}
	m.ParseContext.SetDecodeContent(content, fileData)
}

// AddDependencies :: the block could not be resolved as it has dependencies
// 1) store block as unresolved
// 2) add dependencies to our tree of dependencies
func (m *ModParseContext) AddDependencies(block *hcl.Block, name string, dependencies map[string]*modconfig.ResourceDependency) hcl.Diagnostics {
	// TACTICAL if this is NOT a top level block, add a suffix to the block name
	// this is needed to avoid circular dependency errors if a nested block references
	// a top level block with the same name
	if !m.IsTopLevelBlock(block) {
		name = "nested." + name
	}
	return m.ParseContext.AddDependencies(block, name, dependencies)
}

// ShouldCreateDefaultMod returns whether the flag is set to create a default mod if no mod definition exists
func (m *ModParseContext) ShouldCreateDefaultMod() bool {
	return m.Flags&CreateDefaultMod == CreateDefaultMod
}

// CreatePseudoResources returns whether the flag is set to create pseudo resources
func (m *ModParseContext) CreatePseudoResources() bool {
	return m.Flags&CreatePseudoResources == CreatePseudoResources
}

// AddResource stores this resource as a variable to be added to the eval context. It alse
func (m *ModParseContext) AddResource(resource modconfig.HclResource) hcl.Diagnostics {
	diagnostics := m.storeResourceInCtyMap(resource)
	if diagnostics.HasErrors() {
		return diagnostics
	}

	// rebuild the eval context
	m.buildEvalContext()

	return nil
}

// GetMod finds the mod with given short name, looking only in first level dependencies
// this is used to resolve resource references
// specifically when the 'children' property of dashboards and benchmarks refers to resource in a dependency mod
func (m *ModParseContext) GetMod(modShortName string) *modconfig.Mod {
	if modShortName == m.CurrentMod.ShortName {
		return m.CurrentMod
	}
	// we need to iterate through dependency mods of the current mod
	key := m.CurrentMod.GetInstallCacheKey()
	deps := m.WorkspaceLock.InstallCache[key]
	for _, dep := range deps {
		depMod, ok := m.loadedDependencyMods[dep.FullName()]
		if ok && depMod.ShortName == modShortName {
			return depMod
		}
	}
	return nil
}

func (m *ModParseContext) GetResourceMaps() *modconfig.ResourceMaps {
	// use the current mod as the base resource map
	resourceMap := m.CurrentMod.GetResourceMaps()
	// get a map of top level loaded dep mods
	deps := m.GetTopLevelDependencyMods()

	dependencyResourceMaps := make([]*modconfig.ResourceMaps, 0, len(deps))

	// merge in the top level resources of the dependency mods
	for _, dep := range deps {
		dependencyResourceMaps = append(dependencyResourceMaps, dep.GetResourceMaps().TopLevelResources())
	}

	resourceMap = resourceMap.Merge(dependencyResourceMaps)
	return resourceMap
}

func (m *ModParseContext) GetResource(parsedName *modconfig.ParsedResourceName) (resource modconfig.HclResource, found bool) {
	return m.GetResourceMaps().GetResource(parsedName)
}

// eval functions
func (m *ModParseContext) buildEvalContext() {
	// convert variables to cty values
	variables := make(map[string]cty.Value)

	// now for each mod add all the values
	for mod, modMap := range m.referenceValues {
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

	m.ParseContext.buildEvalContext(variables)
}

// update the cached cty value for the given resource, as long as itr does not already exist
func (m *ModParseContext) storeResourceInCtyMap(resource modconfig.HclResource) hcl.Diagnostics {
	// add resource to variable map
	ctyValue, diags := m.getResourceCtyValue(resource)
	if diags.HasErrors() {
		return diags
	}

	// add into the reference value map
	if diags := m.addReferenceValue(resource, ctyValue); diags.HasErrors() {
		return diags
	}

	// remove this resource from unparsed blocks
	delete(m.UnresolvedBlocks, resource.Name())

	return nil
}

func (m *ModParseContext) getResourceCtyValue(resource modconfig.HclResource) (cty.Value, hcl.Diagnostics) {
	ctyValue, err := resource.(modconfig.CtyValueProvider).CtyValue()
	if err != nil {
		return cty.Zero, m.errToCtyValueDiags(resource, err)
	}
	// if this is a value map, merge in the values of base structs
	// if it is NOT a value map, the resource must have overridden CtyValue so do not merge base structs
	if ctyValue.Type().FriendlyName() != "object" {
		return ctyValue, nil
	}
	// TODO [node_reuse] fetch nested structs and serialise automatically https://github.com/turbot/steampipe/issues/2924
	valueMap := ctyValue.AsValueMap()
	if valueMap == nil {
		valueMap = make(map[string]cty.Value)
	}
	base := resource.GetHclResourceImpl()
	if err := m.mergeResourceCtyValue(base, valueMap); err != nil {
		return cty.Zero, m.errToCtyValueDiags(resource, err)
	}

	if qp, ok := resource.(modconfig.QueryProvider); ok {
		base := qp.GetQueryProviderImpl()
		if err := m.mergeResourceCtyValue(base, valueMap); err != nil {
			return cty.Zero, m.errToCtyValueDiags(resource, err)
		}
	}

	if treeItem, ok := resource.(modconfig.ModTreeItem); ok {
		base := treeItem.GetModTreeItemImpl()
		if err := m.mergeResourceCtyValue(base, valueMap); err != nil {
			return cty.Zero, m.errToCtyValueDiags(resource, err)
		}
	}
	return cty.ObjectVal(valueMap), nil
}

// merge the cty value of the given interface into valueMap
// (note: this mutates valueMap)
func (m *ModParseContext) mergeResourceCtyValue(resource modconfig.CtyValueProvider, valueMap map[string]cty.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(string(debug.Stack()))
			err = fmt.Errorf("panic in mergeResourceCtyValue: %s", helpers.ToError(r).Error())
		}
	}()
	ctyValue, err := resource.CtyValue()
	if err != nil {
		return err
	}
	if ctyValue == cty.Zero {
		return nil
	}
	// merge results
	for k, v := range ctyValue.AsValueMap() {
		valueMap[k] = v
	}
	return nil
}

func (m *ModParseContext) errToCtyValueDiags(resource modconfig.HclResource, err error) hcl.Diagnostics {
	return hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("failed to convert resource '%s' to its cty value", resource.Name()),
		Detail:   err.Error(),
		Subject:  resource.GetDeclRange(),
	}}
}

func (m *ModParseContext) addReferenceValue(resource modconfig.HclResource, value cty.Value) hcl.Diagnostics {
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
	mod := m.CurrentMod
	modName := mod.ShortName
	if mod.ModPath == m.RootEvalPath {
		modName = "local"
	}
	variablesForMod, ok := m.referenceValues[modName]
	// do we have a map of reference values for this dep mod?
	if !ok {
		// no - create one
		variablesForMod = make(ReferenceTypeValueMap)
		m.referenceValues[modName] = variablesForMod
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
		m.referenceValues[modName] = variablesForMod
	}

	return nil
}

//func (m *ModParseContext) `AddLoadedDependentMods`(mods modconfig.ModMap) {
//	update to include version in name
//	for k, v := range mods {
//		if _, alreadyLoaded := m.LoadedDependencyMods[k]; !alreadyLoaded {
//			m.LoadedDependencyMods[k] = v
//		}
//	}
//}

func (m *ModParseContext) IsTopLevelBlock(block *hcl.Block) bool {
	_, isTopLevel := m.topLevelBlocks[block]
	return isTopLevel
}

func (m *ModParseContext) GetLoadedDependencyMod(requiredModVersion *modconfig.ModVersionConstraint, mod *modconfig.Mod) (*modconfig.Mod, error) {
	// if we have a locked version, update the required version to reflect this
	lockedVersion, err := m.WorkspaceLock.GetLockedModVersionConstraint(requiredModVersion, mod)
	if err != nil {
		return nil, err
	}
	if lockedVersion == nil {
		return nil, fmt.Errorf("not all dependencies are installed - run 'steampipe mod install'")
	}
	// use the full name of the locked version as key
	d, _ := m.loadedDependencyMods[lockedVersion.FullName()]
	return d, nil
}

func (m *ModParseContext) AddLoadedDependencyMod(mod *modconfig.Mod) {
	// should never happen
	if mod.DependencyPath == "" {
		return
	}
	m.loadedDependencyMods[mod.DependencyPath] = mod
}

// GetTopLevelDependencyMods build a mod map of top level loaded dependencies, keyed by mod name
func (m *ModParseContext) GetTopLevelDependencyMods() modconfig.ModMap {
	// lazy load m.topLevelDependencyMods
	if m.topLevelDependencyMods != nil {
		return m.topLevelDependencyMods
	}
	// get install cache key fpor this mod (short name for top level mod or ModDependencyPath for dep mods)
	installCacheKey := m.CurrentMod.GetInstallCacheKey()
	deps := m.WorkspaceLock.InstallCache[installCacheKey]
	m.topLevelDependencyMods = make(modconfig.ModMap, len(deps))

	// merge in the dependency mods
	for _, dep := range deps {
		key := dep.FullName()
		loadedDepMod := m.loadedDependencyMods[key]
		if loadedDepMod != nil {
			// as key use the ModDependencyPath _without_ the version
			m.topLevelDependencyMods[loadedDepMod.DependencyName] = loadedDepMod
		}
	}
	return m.topLevelDependencyMods
}
