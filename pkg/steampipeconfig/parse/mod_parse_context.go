package parse

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/steampipe/pkg/steampipeconfig/inputvars"
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

	Flags       ParseModFlag
	ListOptions *filehelpers.ListOptions

	// Variables are populated in an initial parse pass top we store them on the run context
	// so we can set them on the mod when we do the main parse

	// Variables is a tree of maps of the variables in the current mod and child dependency mods
	Variables *modconfig.ModVariableMap

	ParentParseCtx *ModParseContext

	// stack of parent resources for the currently parsed block
	// (unqualified name)
	parents []string

	// map of resource children, keyed by parent unqualified name
	blockChildMap map[string][]string

	// map of top  level blocks, for easy checking
	topLevelBlocks map[*hcl.Block]struct{}
	// map of block names, keyed by a hash of the blopck
	blockNameMap map[string]string
	// map of ReferenceTypeValueMaps keyed by mod name
	// NOTE: all values from root mod are keyed with "local"
	referenceValues map[string]ReferenceTypeValueMap

	// a map of just the top level dependencies of the CurrentMod, keyed my full mod DependencyName (with no version)
	topLevelDependencyMods modconfig.ModMap
	// if we are loading dependency mod, this contains the details
	DependencyConfig *ModDependencyConfig
}

func NewModParseContext(workspaceLock *versionmap.WorkspaceLock, rootEvalPath string, flags ParseModFlag, listOptions *filehelpers.ListOptions) *ModParseContext {
	parseContext := NewParseContext(rootEvalPath)
	c := &ModParseContext{
		ParseContext:  parseContext,
		Flags:         flags,
		WorkspaceLock: workspaceLock,
		ListOptions:   listOptions,

		topLevelDependencyMods: make(modconfig.ModMap),
		blockChildMap:          make(map[string][]string),
		blockNameMap:           make(map[string]string),
		// initialise reference maps - even though we later overwrite them
		referenceValues: map[string]ReferenceTypeValueMap{
			"local": make(ReferenceTypeValueMap),
		},
	}
	// add root node - this will depend on all other nodes
	c.dependencyGraph = c.newDependencyGraph()
	c.buildEvalContext()

	return c
}

func NewChildModParseContext(parent *ModParseContext, modVersion *versionmap.ResolvedVersionConstraint, rootEvalPath string) *ModParseContext {
	// create a child run context
	child := NewModParseContext(
		parent.WorkspaceLock,
		rootEvalPath,
		parent.Flags,
		parent.ListOptions)
	// copy our block tpyes
	child.BlockTypes = parent.BlockTypes
	// set the child's parent
	child.ParentParseCtx = parent
	// set the dependency config
	child.DependencyConfig = NewDependencyConfig(modVersion)
	// set variables if parent has any
	if parent.Variables != nil {
		childVars, ok := parent.Variables.DependencyVariables[modVersion.Name]
		if ok {
			child.Variables = childVars
			child.Variables.PopulatePublicVariables()
			child.AddVariablesToEvalContext()
		}
	}

	return child
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

// VariableValueCtyMap converts a map of variables to a map of the underlying cty value
func VariableValueCtyMap(variables map[string]*modconfig.Variable) map[string]cty.Value {
	ret := make(map[string]cty.Value, len(variables))
	for k, v := range variables {
		ret[k] = v.Value
	}
	return ret
}

// AddInputVariableValues adds evaluated variables to the run context.
// This function is called for the root run context after loading all input variables
func (m *ModParseContext) AddInputVariableValues(inputVariables *modconfig.ModVariableMap) {
	// store the variables
	m.Variables = inputVariables

	// now add variables into eval context
	m.AddVariablesToEvalContext()
}

func (m *ModParseContext) AddVariablesToEvalContext() {
	m.addRootVariablesToReferenceMap()
	m.addDependencyVariablesToReferenceMap()
	m.buildEvalContext()
}

// addRootVariablesToReferenceMap sets the Variables property
// and adds the variables to the referenceValues map (used to build the eval context)
func (m *ModParseContext) addRootVariablesToReferenceMap() {

	variables := m.Variables.RootVariables
	// write local variables directly into referenceValues map
	// NOTE: we add with the name "var" not "variable" as that is how variables are referenced
	m.referenceValues["local"]["var"] = VariableValueCtyMap(variables)
}

// addDependencyVariablesToReferenceMap adds the dependency variables to the referenceValues map
// (used to build the eval context)
func (m *ModParseContext) addDependencyVariablesToReferenceMap() {
	// retrieve the resolved dependency versions for the parent mod
	resolvedVersions := m.WorkspaceLock.InstallCache[m.Variables.Mod.GetInstallCacheKey()]

	for depModName, depVars := range m.Variables.DependencyVariables {
		alias := resolvedVersions[depModName].Alias
		if m.referenceValues[alias] == nil {
			m.referenceValues[alias] = make(ReferenceTypeValueMap)
		}
		m.referenceValues[alias]["var"] = VariableValueCtyMap(depVars.RootVariables)
	}
}

// AddModResources is used to add mod resources to the eval context
func (m *ModParseContext) AddModResources(mod *modconfig.Mod) hcl.Diagnostics {
	if len(m.UnresolvedBlocks) > 0 {
		// should never happen
		panic("calling AddModResources on ModParseContext but there are unresolved blocks from a previous parse")
	}

	var diags hcl.Diagnostics

	moreDiags := m.storeResourceInReferenceValueMap(mod)
	diags = append(diags, moreDiags...)

	// do not add variables (as they have already been added)
	// if the resource is for a dependency mod, do not add locals
	shouldAdd := func(item modconfig.HclResource) bool {
		if item.BlockType() == modconfig.BlockTypeVariable ||
			item.BlockType() == modconfig.BlockTypeLocals && item.(modconfig.ModTreeItem).GetMod().ShortName != m.CurrentMod.ShortName {
			return false
		}
		return true
	}

	resourceFunc := func(item modconfig.HclResource) (bool, error) {
		// add all mod resources (except those excluded) into cty map
		if shouldAdd(item) {
			moreDiags := m.storeResourceInReferenceValueMap(item)
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

// AddResource stores this resource as a variable to be added to the eval context.
func (m *ModParseContext) AddResource(resource modconfig.HclResource) hcl.Diagnostics {
	diagnostics := m.storeResourceInReferenceValueMap(resource)
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
		depMod, ok := m.topLevelDependencyMods[dep.Name]
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

// build the eval context from the cached reference values
func (m *ModParseContext) buildEvalContext() {
	// convert reference values to cty objects
	referenceValues := make(map[string]cty.Value)

	// now for each mod add all the values
	for mod, modMap := range m.referenceValues {
		if mod == "local" {
			for k, v := range modMap {
				referenceValues[k] = cty.ObjectVal(v)
			}
			continue
		}

		// mod map is map[string]map[string]cty.Value
		// for each element (i.e. map[string]cty.Value) convert to cty object
		refTypeMap := make(map[string]cty.Value)
		if mod == "local" {
			for k, v := range modMap {
				referenceValues[k] = cty.ObjectVal(v)
			}
		} else {
			for refType, typeValueMap := range modMap {
				refTypeMap[refType] = cty.ObjectVal(typeValueMap)
			}
		}
		// now convert the referenceValues itself to a cty object
		referenceValues[mod] = cty.ObjectVal(refTypeMap)
	}

	// rebuild the eval context
	m.ParseContext.buildEvalContext(referenceValues)
}

// store the resource as a cty value in the reference valuemap
func (m *ModParseContext) storeResourceInReferenceValueMap(resource modconfig.HclResource) hcl.Diagnostics {
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

// convert a HclResource into a cty value, taking into account nested structs
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

	// most resources will have a mod property - use this if available
	var mod *modconfig.Mod
	if modTreeItem, ok := resource.(modconfig.ModTreeItem); ok {
		mod = modTreeItem.GetMod()
	}
	// fall back to current mod
	if mod == nil {
		mod = m.CurrentMod
	}

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

func (m *ModParseContext) IsTopLevelBlock(block *hcl.Block) bool {
	_, isTopLevel := m.topLevelBlocks[block]
	return isTopLevel
}

func (m *ModParseContext) AddLoadedDependencyMod(mod *modconfig.Mod) {
	m.topLevelDependencyMods[mod.DependencyName] = mod
}

// GetTopLevelDependencyMods build a mod map of top level loaded dependencies, keyed by mod name
func (m *ModParseContext) GetTopLevelDependencyMods() modconfig.ModMap {
	return m.topLevelDependencyMods
}

func (m *ModParseContext) SetCurrentMod(mod *modconfig.Mod) error {
	m.CurrentMod = mod
	// now we have the mod, load any arg values from the mod require - these will be passed to dependency mods
	return m.loadModRequireArgs()
}

// when reloading a mod dependency tree to resolve require args values, this function is called after each mod is loaded
// to load the require arg values and update the variable values
func (m *ModParseContext) loadModRequireArgs() error {
	//if we have not loaded variable definitions yet, do not load require args
	if m.Variables == nil {
		return nil
	}

	depModVarValues, err := inputvars.CollectVariableValuesFromModRequire(m.CurrentMod, m.WorkspaceLock)
	if err != nil {
		return err
	}
	if len(depModVarValues) == 0 {
		return nil
	}
	// if any mod require args have an unknown value, we have failed to resolve them - raise an error
	if err := m.validateModRequireValues(depModVarValues); err != nil {
		return err
	}
	// now update the variables map with the input values
	depModVarValues.SetVariableValues(m.Variables)

	// now add  overridden variables into eval context - in case the root mod references any dependency variable values
	m.AddVariablesToEvalContext()

	return nil
}

func (m *ModParseContext) validateModRequireValues(depModVarValues inputvars.InputValues) error {
	if len(depModVarValues) == 0 {
		return nil
	}
	var missingVarExpressions []string
	requireBlock := m.getModRequireBlock()
	if requireBlock == nil {
		return fmt.Errorf("require args extracted but no require block found for %s", m.CurrentMod.Name())
	}

	for k, v := range depModVarValues {
		// if we successfully resolved this value, continue
		if v.Value.IsKnown() {
			continue
		}
		parsedVarName, err := modconfig.ParseResourceName(k)
		if err != nil {
			return err
		}

		// re-parse the require block manually to extract the range and unresolved arg value expression
		var errorString string
		errorString, err = m.getErrorStringForUnresolvedArg(parsedVarName, requireBlock)
		if err != nil {
			// if there was an error retrieving details, return less specific error string
			errorString = fmt.Sprintf("\"%s\"  (%s %s)", k, m.CurrentMod.Name(), m.CurrentMod.GetDeclRange().Filename)
		}

		missingVarExpressions = append(missingVarExpressions, errorString)
	}

	if errorCount := len(missingVarExpressions); errorCount > 0 {
		if errorCount == 1 {
			return fmt.Errorf("failed to resolve dependency mod argument value: %s", missingVarExpressions[0])
		}

		return fmt.Errorf("failed to resolve %d dependency mod arguments %s:\n\t%s", errorCount, utils.Pluralize("value", errorCount), strings.Join(missingVarExpressions, "\n\t"))
	}
	return nil
}

func (m *ModParseContext) getErrorStringForUnresolvedArg(parsedVarName *modconfig.ParsedResourceName, requireBlock *hclsyntax.Block) (_ string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()
	// which mod and variable is this is this for
	modShortName := parsedVarName.Mod
	varName := parsedVarName.Name
	var modDependencyName string
	// determine the mod dependency name as that is how it will be keyed in the require map
	for depName, modVersion := range m.WorkspaceLock.InstallCache[m.CurrentMod.GetInstallCacheKey()] {
		if modVersion.Alias == modShortName {
			modDependencyName = depName
			break
		}
	}

	// iterate through require blocks looking for mod blocks
	for _, b := range requireBlock.Body.Blocks {
		// only interested in mod blocks
		if b.Type != "mod" {
			continue
		}
		// if this is not the mod we're looking for, continue
		if b.Labels[0] != modDependencyName {
			continue
		}
		// now find the failed arg
		argsAttr, ok := b.Body.Attributes["args"]
		if !ok {
			return "", fmt.Errorf("no args block found for %s", modDependencyName)
		}
		// iterate over args looking for the correctly named item
		for _, a := range argsAttr.Expr.(*hclsyntax.ObjectConsExpr).Items {
			thisVarName, err := a.KeyExpr.Value(&hcl.EvalContext{})
			if err != nil {
				return "", err
			}

			// is this the var we are looking for?
			if thisVarName.AsString() != varName {
				continue
			}

			// this is the var, get the value expression
			expr, ok := a.ValueExpr.(*hclsyntax.ScopeTraversalExpr)
			if !ok {
				return "", fmt.Errorf("failed to get args details for %s", parsedVarName.ToResourceName())
			}
			// ok we have the expression - build the error string
			exprString := hclhelpers.TraversalAsString(expr.Traversal)
			r := expr.Range()
			sourceRange := fmt.Sprintf("%s:%d", r.Filename, r.Start.Line)
			res := fmt.Sprintf("\"%s = %s\" (%s %s)",
				parsedVarName.ToResourceName(),
				exprString,
				m.CurrentMod.Name(),
				sourceRange)
			return res, nil

		}
	}
	return "", fmt.Errorf("failed to get args details for %s", parsedVarName.ToResourceName())
}

func (m *ModParseContext) getModRequireBlock() *hclsyntax.Block {
	for _, b := range m.CurrentMod.ResourceWithMetadataBaseRemain.(*hclsyntax.Body).Blocks {
		if b.Type == modconfig.BlockTypeRequire {
			return b
		}
	}
	return nil

}
