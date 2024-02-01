package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig/var_config"
)

// A consistent detail message for all "not a valid identifier" diagnostics.
const badIdentifierDetail = "A name must start with a letter or underscore and may contain only letters, digits, underscores, and dashes."

var missingVariableErrors = []string{
	// returned when the context variables does not have top level 'type' node (locals/control/etc)
	"Unknown variable",
	// returned when the variables have the type object but a field has not yet been populated
	"Unsupported attribute",
	"Missing map element",
}

func decode(parseCtx *ModParseContext) hcl.Diagnostics {
	var diags hcl.Diagnostics

	blocks, err := parseCtx.BlocksToDecode()
	// build list of blocks to decode
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to determine required dependency order",
			Detail:   err.Error()})
		return diags
	}

	// now clear dependencies from run context - they will be rebuilt
	parseCtx.ClearDependencies()

	for _, block := range blocks {
		if block.Type == modconfig.BlockTypeLocals {
			resources, res := decodeLocalsBlock(block, parseCtx)
			if !res.Success() {
				diags = append(diags, res.Diags...)
				continue
			}
			for _, resource := range resources {
				resourceDiags := addResourceToMod(resource, block, parseCtx)
				diags = append(diags, resourceDiags...)
			}
		} else {
			resource, res := decodeBlock(block, parseCtx)
			diags = append(diags, res.Diags...)
			if !res.Success() || resource == nil {
				continue
			}

			resourceDiags := addResourceToMod(resource, block, parseCtx)
			diags = append(diags, resourceDiags...)
		}
	}

	return diags
}

func addResourceToMod(resource modconfig.HclResource, block *hcl.Block, parseCtx *ModParseContext) hcl.Diagnostics {
	if !shouldAddToMod(resource, block, parseCtx) {
		return nil
	}
	return parseCtx.CurrentMod.AddResource(resource)

}

func shouldAddToMod(resource modconfig.HclResource, block *hcl.Block, parseCtx *ModParseContext) bool {
	switch resource.(type) {
	// do not add mods, withs
	case *modconfig.Mod, *modconfig.DashboardWith:
		return false

	case *modconfig.DashboardCategory, *modconfig.DashboardInput:
		// if this is a dashboard category or dashboard input, only add top level blocks
		// this is to allow nested categories/inputs to have the same name as top level categories
		// (nested inputs are added by Dashboard.InitInputs)
		return parseCtx.IsTopLevelBlock(block)
	default:
		return true
	}
}

// special case decode logic for locals
func decodeLocalsBlock(block *hcl.Block, parseCtx *ModParseContext) ([]modconfig.HclResource, *DecodeResult) {
	var resources []modconfig.HclResource
	var res = newDecodeResult()

	// TODO remove and call ShouldIncludeBlock from BlocksToDecode
	// https://github.com/turbot/steampipe/issues/2640
	// if opts specifies block types, then check whether this type is included
	if !parseCtx.ShouldIncludeBlock(block) {
		return nil, res
	}

	// check name is valid
	diags := validateName(block)
	if diags.HasErrors() {
		res.addDiags(diags)
		return nil, res
	}

	var locals []*modconfig.Local
	locals, res = decodeLocals(block, parseCtx)
	for _, local := range locals {
		resources = append(resources, local)
		handleModDecodeResult(local, res, block, parseCtx)
	}

	return resources, res
}

func decodeBlock(block *hcl.Block, parseCtx *ModParseContext) (modconfig.HclResource, *DecodeResult) {
	var resource modconfig.HclResource
	var res = newDecodeResult()

	// TODO remove and call ShouldIncludeBlock from BlocksToDecode
	// https://github.com/turbot/steampipe/issues/2640
	// if opts specifies block types, then check whether this type is included
	if !parseCtx.ShouldIncludeBlock(block) {
		return nil, res
	}

	// has this block already been decoded?
	// (this could happen if it is a child block and has been decoded before its parent as part of second decode phase)
	if resource, ok := parseCtx.GetDecodedResourceForBlock(block); ok {
		return resource, res
	}

	// check name is valid
	diags := validateName(block)
	if diags.HasErrors() {
		res.addDiags(diags)
		return nil, res
	}

	// now do the actual decode
	switch {
	case helpers.StringSliceContains(modconfig.NodeAndEdgeProviderBlocks, block.Type):
		resource, res = decodeNodeAndEdgeProvider(block, parseCtx)
	case helpers.StringSliceContains(modconfig.QueryProviderBlocks, block.Type):
		resource, res = decodeQueryProvider(block, parseCtx)
	default:
		switch block.Type {
		case modconfig.BlockTypeMod:
			// decodeMode has slightly different args as this code is shared with ParseModDefinition
			resource, res = decodeMod(block, parseCtx.EvalCtx, parseCtx.CurrentMod)
		case modconfig.BlockTypeDashboard:
			resource, res = decodeDashboard(block, parseCtx)
		case modconfig.BlockTypeContainer:
			resource, res = decodeDashboardContainer(block, parseCtx)
		case modconfig.BlockTypeVariable:
			resource, res = decodeVariable(block, parseCtx)
		case modconfig.BlockTypeBenchmark:
			resource, res = decodeBenchmark(block, parseCtx)
		default:
			// all other blocks are treated the same:
			resource, res = decodeResource(block, parseCtx)
		}
	}

	// handle the result
	// - if there are dependencies, add to run context
	handleModDecodeResult(resource, res, block, parseCtx)

	return resource, res
}

func decodeMod(block *hcl.Block, evalCtx *hcl.EvalContext, mod *modconfig.Mod) (*modconfig.Mod, *DecodeResult) {
	res := newDecodeResult()
	// decode the body
	diags := decodeHclBody(block.Body, evalCtx, mod, mod)
	res.handleDecodeDiags(diags)
	return mod, res
}

// generic decode function for any resource we do not have custom decode logic for
func decodeResource(block *hcl.Block, parseCtx *ModParseContext) (modconfig.HclResource, *DecodeResult) {
	res := newDecodeResult()
	// get shell resource
	resource, diags := resourceForBlock(block, parseCtx)
	res.handleDecodeDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	diags = decodeHclBody(block.Body, parseCtx.EvalCtx, parseCtx, resource)
	if len(diags) > 0 {
		res.handleDecodeDiags(diags)
	}
	return resource, res
}

// return a shell resource for the given block
func resourceForBlock(block *hcl.Block, parseCtx *ModParseContext) (modconfig.HclResource, hcl.Diagnostics) {
	var resource modconfig.HclResource
	// parseCtx already contains the current mod
	mod := parseCtx.CurrentMod
	blockName := parseCtx.DetermineBlockName(block)

	factoryFuncs := map[string]func(*hcl.Block, *modconfig.Mod, string) modconfig.HclResource{
		// for block type mod, just use the current mod
		modconfig.BlockTypeMod:       func(*hcl.Block, *modconfig.Mod, string) modconfig.HclResource { return mod },
		modconfig.BlockTypeQuery:     modconfig.NewQuery,
		modconfig.BlockTypeControl:   modconfig.NewControl,
		modconfig.BlockTypeBenchmark: modconfig.NewBenchmark,
		modconfig.BlockTypeDashboard: modconfig.NewDashboard,
		modconfig.BlockTypeContainer: modconfig.NewDashboardContainer,
		modconfig.BlockTypeChart:     modconfig.NewDashboardChart,
		modconfig.BlockTypeCard:      modconfig.NewDashboardCard,
		modconfig.BlockTypeFlow:      modconfig.NewDashboardFlow,
		modconfig.BlockTypeGraph:     modconfig.NewDashboardGraph,
		modconfig.BlockTypeHierarchy: modconfig.NewDashboardHierarchy,
		modconfig.BlockTypeImage:     modconfig.NewDashboardImage,
		modconfig.BlockTypeInput:     modconfig.NewDashboardInput,
		modconfig.BlockTypeTable:     modconfig.NewDashboardTable,
		modconfig.BlockTypeText:      modconfig.NewDashboardText,
		modconfig.BlockTypeNode:      modconfig.NewDashboardNode,
		modconfig.BlockTypeEdge:      modconfig.NewDashboardEdge,
		modconfig.BlockTypeCategory:  modconfig.NewDashboardCategory,
		modconfig.BlockTypeWith:      modconfig.NewDashboardWith,
	}

	factoryFunc, ok := factoryFuncs[block.Type]
	if !ok {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("resourceForBlock called for unsupported block type %s", block.Type),
			Subject:  hclhelpers.BlockRangePointer(block),
		},
		}
	}
	resource = factoryFunc(block, mod, blockName)
	return resource, nil
}

func decodeLocals(block *hcl.Block, parseCtx *ModParseContext) ([]*modconfig.Local, *DecodeResult) {
	res := newDecodeResult()
	attrs, diags := block.Body.JustAttributes()
	if len(attrs) == 0 {
		res.Diags = diags
		return nil, res
	}

	// build list of locals
	locals := make([]*modconfig.Local, 0, len(attrs))
	for name, attr := range attrs {
		if !hclsyntax.ValidIdentifier(name) {
			res.Diags = append(res.Diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid local value name",
				Detail:   badIdentifierDetail,
				Subject:  &attr.NameRange,
			})
			continue
		}
		// try to evaluate expression
		val, diags := attr.Expr.Value(parseCtx.EvalCtx)
		// handle any resulting diags, which may specify dependencies
		res.handleDecodeDiags(diags)

		// add to our list
		locals = append(locals, modconfig.NewLocal(name, val, attr.Range, parseCtx.CurrentMod))
	}
	return locals, res
}

func decodeVariable(block *hcl.Block, parseCtx *ModParseContext) (*modconfig.Variable, *DecodeResult) {
	res := newDecodeResult()

	var variable *modconfig.Variable
	content, diags := block.Body.Content(VariableBlockSchema)
	res.handleDecodeDiags(diags)

	v, diags := var_config.DecodeVariableBlock(block, content, false)
	res.handleDecodeDiags(diags)

	if res.Success() {
		variable = modconfig.NewVariable(v, parseCtx.CurrentMod)
	}

	return variable, res

}

func decodeQueryProvider(block *hcl.Block, parseCtx *ModParseContext) (modconfig.QueryProvider, *DecodeResult) {
	res := newDecodeResult()
	// get shell resource
	resource, diags := resourceForBlock(block, parseCtx)
	res.handleDecodeDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}
	// do a partial decode using an empty schema - use to pull out all body content in the remain block
	_, remain, diags := block.Body.PartialContent(&hcl.BodySchema{})
	res.handleDecodeDiags(diags)
	if !res.Success() {
		return nil, res
	}

	// decode the body into 'resource' to populate all properties that can be automatically decoded
	diags = decodeHclBody(remain, parseCtx.EvalCtx, parseCtx, resource)
	res.handleDecodeDiags(diags)

	// decode 'with',args and params blocks
	res.Merge(decodeQueryProviderBlocks(block, remain.(*hclsyntax.Body), resource, parseCtx))

	return resource.(modconfig.QueryProvider), res
}

func decodeQueryProviderBlocks(block *hcl.Block, content *hclsyntax.Body, resource modconfig.HclResource, parseCtx *ModParseContext) *DecodeResult {
	var diags hcl.Diagnostics
	res := newDecodeResult()
	queryProvider, ok := resource.(modconfig.QueryProvider)
	if !ok {
		// coding error
		panic(fmt.Sprintf("block type %s not convertible to a QueryProvider", block.Type))
	}

	if attr, exists := content.Attributes[modconfig.AttributeArgs]; exists {
		args, runtimeDependencies, diags := decodeArgs(attr.AsHCLAttribute(), parseCtx.EvalCtx, queryProvider)
		if diags.HasErrors() {
			// handle dependencies
			res.handleDecodeDiags(diags)
		} else {
			queryProvider.SetArgs(args)
			queryProvider.AddRuntimeDependencies(runtimeDependencies)
		}
	}

	var params []*modconfig.ParamDef
	for _, b := range content.Blocks {
		block = b.AsHCLBlock()
		switch block.Type {
		case modconfig.BlockTypeParam:
			paramDef, runtimeDependencies, moreDiags := decodeParam(block, parseCtx)
			if !moreDiags.HasErrors() {
				params = append(params, paramDef)
				queryProvider.AddRuntimeDependencies(runtimeDependencies)
				// add and references contained in the param block to the control refs
				moreDiags = AddReferences(resource, block, parseCtx)
			}
			diags = append(diags, moreDiags...)
		}
	}

	queryProvider.SetParams(params)
	res.handleDecodeDiags(diags)
	return res
}

func decodeNodeAndEdgeProvider(block *hcl.Block, parseCtx *ModParseContext) (modconfig.HclResource, *DecodeResult) {
	res := newDecodeResult()

	// get shell resource
	resource, diags := resourceForBlock(block, parseCtx)
	res.handleDecodeDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	nodeAndEdgeProvider, ok := resource.(modconfig.NodeAndEdgeProvider)
	if !ok {
		// coding error
		panic(fmt.Sprintf("block type %s not convertible to a NodeAndEdgeProvider", block.Type))
	}

	// do a partial decode using an empty schema - use to pull out all body content in the remain block
	_, r, diags := block.Body.PartialContent(&hcl.BodySchema{})
	body := r.(*hclsyntax.Body)
	res.handleDecodeDiags(diags)
	if !res.Success() {
		return nil, res
	}

	// decode the body into 'resource' to populate all properties that can be automatically decoded
	diags = decodeHclBody(body, parseCtx.EvalCtx, parseCtx, resource)
	// handle any resulting diags, which may specify dependencies
	res.handleDecodeDiags(diags)

	// decode sql args and params
	res.Merge(decodeQueryProviderBlocks(block, body, resource, parseCtx))

	// now decode child blocks
	if len(body.Blocks) > 0 {
		blocksRes := decodeNodeAndEdgeProviderBlocks(body, nodeAndEdgeProvider, parseCtx)
		res.Merge(blocksRes)
	}

	return resource, res
}

func decodeNodeAndEdgeProviderBlocks(content *hclsyntax.Body, nodeAndEdgeProvider modconfig.NodeAndEdgeProvider, parseCtx *ModParseContext) *DecodeResult {
	var res = newDecodeResult()

	for _, b := range content.Blocks {
		block := b.AsHCLBlock()
		switch block.Type {
		case modconfig.BlockTypeCategory:
			// decode block
			category, blockRes := decodeBlock(block, parseCtx)
			res.Merge(blockRes)
			if !blockRes.Success() {
				continue
			}

			// add the category to the nodeAndEdgeProvider
			res.addDiags(nodeAndEdgeProvider.AddCategory(category.(*modconfig.DashboardCategory)))

			// DO NOT add the category to the mod

		case modconfig.BlockTypeNode, modconfig.BlockTypeEdge:
			child, childRes := decodeQueryProvider(block, parseCtx)

			// TACTICAL if child has any runtime dependencies, claim them
			// this is to ensure if this resource is used as base, we can be correctly identified
			// as the publisher of the runtime dependencies
			for _, r := range child.GetRuntimeDependencies() {
				r.Provider = nodeAndEdgeProvider
			}

			// populate metadata, set references and call OnDecoded
			handleModDecodeResult(child, childRes, block, parseCtx)
			res.Merge(childRes)
			if res.Success() {
				moreDiags := nodeAndEdgeProvider.AddChild(child)
				res.addDiags(moreDiags)
			}
		case modconfig.BlockTypeWith:
			with, withRes := decodeBlock(block, parseCtx)
			res.Merge(withRes)
			if res.Success() {
				moreDiags := nodeAndEdgeProvider.AddWith(with.(*modconfig.DashboardWith))
				res.addDiags(moreDiags)
			}
		}

	}

	return res
}

func decodeDashboard(block *hcl.Block, parseCtx *ModParseContext) (*modconfig.Dashboard, *DecodeResult) {
	res := newDecodeResult()
	dashboard := modconfig.NewDashboard(block, parseCtx.CurrentMod, parseCtx.DetermineBlockName(block)).(*modconfig.Dashboard)

	// do a partial decode using an empty schema - use to pull out all body content in the remain block
	_, r, diags := block.Body.PartialContent(&hcl.BodySchema{})
	body := r.(*hclsyntax.Body)
	res.handleDecodeDiags(diags)

	// decode the body into 'dashboardContainer' to populate all properties that can be automatically decoded
	diags = decodeHclBody(body, parseCtx.EvalCtx, parseCtx, dashboard)
	// handle any resulting diags, which may specify dependencies
	res.handleDecodeDiags(diags)

	if dashboard.Base != nil && len(dashboard.Base.ChildNames) > 0 {
		supportedChildren := []string{modconfig.BlockTypeContainer, modconfig.BlockTypeChart, modconfig.BlockTypeControl, modconfig.BlockTypeCard, modconfig.BlockTypeFlow, modconfig.BlockTypeGraph, modconfig.BlockTypeHierarchy, modconfig.BlockTypeImage, modconfig.BlockTypeInput, modconfig.BlockTypeTable, modconfig.BlockTypeText}
		// TACTICAL: we should be passing in the block for the Base resource - but this is only used for diags
		// and we do not expect to get any (as this function has already succeeded when the base was originally parsed)
		children, _ := resolveChildrenFromNames(dashboard.Base.ChildNames, block, supportedChildren, parseCtx)
		dashboard.Base.SetChildren(children)
	}
	if !res.Success() {
		return dashboard, res
	}

	// now decode child blocks
	if len(body.Blocks) > 0 {
		blocksRes := decodeDashboardBlocks(body, dashboard, parseCtx)
		res.Merge(blocksRes)
	}

	return dashboard, res
}

func decodeDashboardBlocks(content *hclsyntax.Body, dashboard *modconfig.Dashboard, parseCtx *ModParseContext) *DecodeResult {
	var res = newDecodeResult()
	// set dashboard as parent on the run context - this is used when generating names for anonymous blocks
	parseCtx.PushParent(dashboard)
	defer func() {
		parseCtx.PopParent()
	}()

	for _, b := range content.Blocks {
		block := b.AsHCLBlock()

		// decode block
		resource, blockRes := decodeBlock(block, parseCtx)
		res.Merge(blockRes)
		if !blockRes.Success() {
			continue
		}

		// we expect either inputs or child report nodes
		// add the resource to the mod
		res.addDiags(addResourceToMod(resource, block, parseCtx))
		// add to the dashboard children
		// (we expect this cast to always succeed)
		if child, ok := resource.(modconfig.ModTreeItem); ok {
			dashboard.AddChild(child)
		}

	}

	moreDiags := dashboard.InitInputs()
	res.addDiags(moreDiags)

	return res
}

func decodeDashboardContainer(block *hcl.Block, parseCtx *ModParseContext) (*modconfig.DashboardContainer, *DecodeResult) {
	res := newDecodeResult()
	container := modconfig.NewDashboardContainer(block, parseCtx.CurrentMod, parseCtx.DetermineBlockName(block)).(*modconfig.DashboardContainer)

	// do a partial decode using an empty schema - use to pull out all body content in the remain block
	_, r, diags := block.Body.PartialContent(&hcl.BodySchema{})
	body := r.(*hclsyntax.Body)
	res.handleDecodeDiags(diags)
	if !res.Success() {
		return nil, res
	}

	// decode the body into 'dashboardContainer' to populate all properties that can be automatically decoded
	diags = decodeHclBody(body, parseCtx.EvalCtx, parseCtx, container)
	// handle any resulting diags, which may specify dependencies
	res.handleDecodeDiags(diags)

	// now decode child blocks
	if len(body.Blocks) > 0 {
		blocksRes := decodeDashboardContainerBlocks(body, container, parseCtx)
		res.Merge(blocksRes)
	}

	return container, res
}

func decodeDashboardContainerBlocks(content *hclsyntax.Body, dashboardContainer *modconfig.DashboardContainer, parseCtx *ModParseContext) *DecodeResult {
	var res = newDecodeResult()

	// set container as parent on the run context - this is used when generating names for anonymous blocks
	parseCtx.PushParent(dashboardContainer)
	defer func() {
		parseCtx.PopParent()
	}()

	for _, b := range content.Blocks {
		block := b.AsHCLBlock()
		resource, blockRes := decodeBlock(block, parseCtx)
		res.Merge(blockRes)
		if !blockRes.Success() {
			continue
		}

		// special handling for inputs
		if b.Type == modconfig.BlockTypeInput {
			input := resource.(*modconfig.DashboardInput)
			dashboardContainer.Inputs = append(dashboardContainer.Inputs, input)
			dashboardContainer.AddChild(input)
			// the input will be added to the mod by the parent dashboard

		} else {
			// for all other children, add to mod and children
			res.addDiags(addResourceToMod(resource, block, parseCtx))
			if child, ok := resource.(modconfig.ModTreeItem); ok {
				dashboardContainer.AddChild(child)
			}
		}
	}

	return res
}

func decodeBenchmark(block *hcl.Block, parseCtx *ModParseContext) (*modconfig.Benchmark, *DecodeResult) {
	res := newDecodeResult()
	benchmark := modconfig.NewBenchmark(block, parseCtx.CurrentMod, parseCtx.DetermineBlockName(block)).(*modconfig.Benchmark)
	content, diags := block.Body.Content(BenchmarkBlockSchema)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "children", &benchmark.ChildNames, parseCtx.EvalCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "description", &benchmark.Description, parseCtx.EvalCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "documentation", &benchmark.Documentation, parseCtx.EvalCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "tags", &benchmark.Tags, parseCtx.EvalCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "title", &benchmark.Title, parseCtx.EvalCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "type", &benchmark.Type, parseCtx.EvalCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "display", &benchmark.Display, parseCtx.EvalCtx)
	res.handleDecodeDiags(diags)

	// now add children
	if res.Success() {
		supportedChildren := []string{modconfig.BlockTypeBenchmark, modconfig.BlockTypeControl}
		children, diags := resolveChildrenFromNames(benchmark.ChildNames.StringList(), block, supportedChildren, parseCtx)
		res.handleDecodeDiags(diags)

		// now set children and child name strings
		benchmark.SetChildren(children)
		benchmark.ChildNameStrings = getChildNameStringsFromModTreeItem(children)
	}

	diags = decodeProperty(content, "base", &benchmark.Base, parseCtx.EvalCtx)
	res.handleDecodeDiags(diags)
	if benchmark.Base != nil && len(benchmark.Base.ChildNames) > 0 {
		supportedChildren := []string{modconfig.BlockTypeBenchmark, modconfig.BlockTypeControl}
		// TACTICAL: we should be passing in the block for the Base resource - but this is only used for diags
		// and we do not expect to get any (as this function has already succeeded when the base was originally parsed)
		children, _ := resolveChildrenFromNames(benchmark.Base.ChildNameStrings, block, supportedChildren, parseCtx)
		benchmark.Base.SetChildren(children)
	}
	diags = decodeProperty(content, "width", &benchmark.Width, parseCtx.EvalCtx)
	res.handleDecodeDiags(diags)
	return benchmark, res
}

func decodeProperty(content *hcl.BodyContent, property string, dest interface{}, evalCtx *hcl.EvalContext) hcl.Diagnostics {
	var diags hcl.Diagnostics
	if attr, ok := content.Attributes[property]; ok {
		diags = gohcl.DecodeExpression(attr.Expr, evalCtx, dest)
	}
	return diags
}

// handleModDecodeResult
// if decode was successful:
// - generate and set resource metadata
// - add resource to ModParseContext (which adds it to the mod)handleModDecodeResult
func handleModDecodeResult(resource modconfig.HclResource, res *DecodeResult, block *hcl.Block, parseCtx *ModParseContext) {
	if !res.Success() {
		if len(res.Depends) > 0 {
			moreDiags := parseCtx.AddDependencies(block, resource.GetUnqualifiedName(), res.Depends)
			res.addDiags(moreDiags)
		}
		return
	}
	// set whether this is a top level resource
	resource.SetTopLevel(parseCtx.IsTopLevelBlock(block))

	// call post decode hook
	// NOTE: must do this BEFORE adding resource to run context to ensure we respect the base property
	moreDiags := resource.OnDecoded(block, parseCtx)
	res.addDiags(moreDiags)

	// add references
	moreDiags = AddReferences(resource, block, parseCtx)
	res.addDiags(moreDiags)

	// validate the resource
	moreDiags = validateResource(resource)
	res.addDiags(moreDiags)
	// if we failed validation, return
	if !res.Success() {
		return
	}

	// if resource is NOT anonymous, and this is a TOP LEVEL BLOCK, add into the run context
	// NOTE: we can only reference resources defined in a top level block
	if !resourceIsAnonymous(resource) && resource.IsTopLevel() {
		moreDiags = parseCtx.AddResource(resource)
		res.addDiags(moreDiags)
	}

	// if resource supports metadata, save it
	if resourceWithMetadata, ok := resource.(modconfig.ResourceWithMetadata); ok {
		moreDiags = addResourceMetadata(resourceWithMetadata, resource.GetHclResourceImpl().DeclRange, parseCtx)
		res.addDiags(moreDiags)
	}
}

func resourceIsAnonymous(resource modconfig.HclResource) bool {
	// (if a resource anonymous it must support ResourceWithMetadata)
	resourceWithMetadata, ok := resource.(modconfig.ResourceWithMetadata)
	anonymousResource := ok && resourceWithMetadata.IsAnonymous()
	return anonymousResource
}

func addResourceMetadata(resourceWithMetadata modconfig.ResourceWithMetadata, srcRange hcl.Range, parseCtx *ModParseContext) hcl.Diagnostics {
	metadata, err := GetMetadataForParsedResource(resourceWithMetadata.Name(), srcRange, parseCtx.FileData, parseCtx.CurrentMod)
	if err != nil {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err.Error(),
			Subject:  &srcRange,
		}}
	}
	//  set on resource
	resourceWithMetadata.SetMetadata(metadata)
	return nil
}

func validateName(block *hcl.Block) hcl.Diagnostics {
	if len(block.Labels) == 0 {
		return nil
	}

	if !hclsyntax.ValidIdentifier(block.Labels[0]) {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[0],
		}}
	}
	return nil
}

// Validate all blocks and attributes are supported
// We use partial decoding so that we can automatically decode as many properties as possible
// and only manually decode properties requiring special logic.
// The problem is the partial decode does not return errors for invalid attributes/blocks, so we must implement our own
func validateHcl(blockType string, body *hclsyntax.Body, schema *hcl.BodySchema) hcl.Diagnostics {
	var diags hcl.Diagnostics

	// identify any blocks specified by hcl tags
	var supportedBlocks = make(map[string]struct{})
	var supportedAttributes = make(map[string]struct{})
	for _, b := range schema.Blocks {
		supportedBlocks[b.Type] = struct{}{}
	}
	for _, b := range schema.Attributes {
		supportedAttributes[b.Name] = struct{}{}
	}

	// now check for invalid blocks
	for _, block := range body.Blocks {
		if _, ok := supportedBlocks[block.Type]; !ok {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf(`Unsupported block type: Blocks of type '%s' are not expected here.`, block.Type),
				Subject:  &block.TypeRange,
			})
		}
	}
	for _, attribute := range body.Attributes {
		if _, ok := supportedAttributes[attribute.Name]; !ok {
			// special case code for deprecated properties
			subject := attribute.Range()
			if isDeprecated(attribute, blockType) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagWarning,
					Summary:  fmt.Sprintf(`Deprecated attribute: '%s' is deprecated for '%s' blocks and will be ignored.`, attribute.Name, blockType),
					Subject:  &subject,
				})
			} else {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf(`Unsupported attribute: '%s' not expected here.`, attribute.Name),
					Subject:  &subject,
				})
			}
		}
	}

	return diags
}

func isDeprecated(attribute *hclsyntax.Attribute, blockType string) bool {
	switch attribute.Name {
	case "search_path", "search_path_prefix":
		return blockType == modconfig.BlockTypeQuery || blockType == modconfig.BlockTypeControl
	default:
		return false
	}
}
