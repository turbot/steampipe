package parse

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig/var_config"
	"github.com/turbot/steampipe/utils"
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

func decode(runCtx *RunContext) hcl.Diagnostics {
	var diags hcl.Diagnostics

	// build list of blocks to decode
	blocks, err := runCtx.BlocksToDecode()

	// now clear dependencies from run context - they will be rebuilt
	// TODO [reports] ALSO CLEAR MOD CHILDREN WHICH MAY HAVE BEEN PARTIALLY ADDED
	// THEN WE CAN UPDATE THE DUPE CHECKING CODE
	runCtx.ClearDependencies()

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to determine required dependency order",
			Detail:   err.Error()})
	}

	for _, block := range blocks {
		resources, res := decodeBlock(block, runCtx)
		if !res.Success() {
			diags = append(diags, res.Diags...)
			continue
		}
		resourceDiags := addResourcesToMod(runCtx, resources...)
		diags = append(diags, resourceDiags...)

	}

	return diags
}

func addResourcesToMod(runCtx *RunContext, resources ...modconfig.HclResource) hcl.Diagnostics {
	var diags hcl.Diagnostics
	for _, resource := range resources {
		if _, ok := resource.(*modconfig.Mod); !ok {
			moreDiags := runCtx.CurrentMod.AddResource(resource)
			diags = append(diags, moreDiags...)
		}
	}
	return diags
}

func decodeBlock(block *hcl.Block, runCtx *RunContext) ([]modconfig.HclResource, *decodeResult) {
	var resource modconfig.HclResource
	var resources []modconfig.HclResource
	var res = &decodeResult{}

	// if opts specifies block types, check whether this type is included
	if !runCtx.ShouldIncludeBlock(block) {
		return nil, res
	}

	// has this block already been decoded?
	// (this could happen if it is a child block and has been decoded before its parent as part of second decode phase)
	if resource, ok := runCtx.GetDecodedResourceForBlock(block); ok {
		return []modconfig.HclResource{resource}, res
	}

	// check name is valid
	diags := validateName(block)
	if diags.HasErrors() {
		res.addDiags(diags)
		return nil, res
	}

	// now do the actual decode
	if helpers.StringSliceContains(modconfig.QueryProviderBlocks, block.Type) {
		resource, res = decodeQueryProvider(block, runCtx)
		resources = append(resources, resource)
	} else {
		switch block.Type {
		case modconfig.BlockTypeLocals:
			// special case decode logic for locals
			var locals []*modconfig.Local
			locals, res = decodeLocals(block, runCtx)
			for _, local := range locals {
				resources = append(resources, local)
			}
		case modconfig.BlockTypeDashboard:
			resource, res = decodeDashboard(block, runCtx)
			resources = append(resources, resource)
		case modconfig.BlockTypeContainer:
			resource, res = decodeDashboardContainer(block, runCtx)
			resources = append(resources, resource)
		case modconfig.BlockTypeVariable:
			resource, res = decodeVariable(block, runCtx)
			resources = append(resources, resource)
		case modconfig.BlockTypeBenchmark:
			resource, res = decodeBenchmark(block, runCtx)
			resources = append(resources, resource)
		default:
			// all other blocks are treated the same:
			resource, res = decodeResource(block, runCtx)
			resources = append(resources, resource)
		}
	}

	for _, resource := range resources {
		// handle the result
		// - if there are dependencies, add to run context
		handleDecodeResult(resource, res, block, runCtx)
	}

	return resources, res
}

// generic decode function for any resource we do not have custom decode logic for
func decodeResource(block *hcl.Block, runCtx *RunContext) (modconfig.HclResource, *decodeResult) {
	res := &decodeResult{}
	// get shell resource
	resource, diags := resourceForBlock(block, runCtx)
	res.handleDecodeDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	diags = gohcl.DecodeBody(block.Body, runCtx.EvalCtx, resource)
	if len(diags) > 0 {
		res.handleDecodeDiags(diags)
	}
	return resource, res
}

// return a shell resource for the given block
func resourceForBlock(block *hcl.Block, runCtx *RunContext) (modconfig.HclResource, hcl.Diagnostics) {
	var resource modconfig.HclResource
	// runCtx already contains the current mod
	mod := runCtx.CurrentMod
	blockName := runCtx.DetermineBlockName(block)
	switch block.Type {
	case modconfig.BlockTypeMod:
		resource = mod
	case modconfig.BlockTypeQuery:
		resource = modconfig.NewQuery(block, mod, blockName)
	case modconfig.BlockTypeControl:
		resource = modconfig.NewControl(block, mod, blockName)
	case modconfig.BlockTypeBenchmark:
		resource = modconfig.NewBenchmark(block, mod, blockName)
	case modconfig.BlockTypeDashboard:
		resource = modconfig.NewDashboard(block, mod, blockName)
	case modconfig.BlockTypeContainer:
		resource = modconfig.NewDashboardContainer(block, mod, blockName)
	case modconfig.BlockTypeChart:
		resource = modconfig.NewDashboardChart(block, mod, blockName)
	case modconfig.BlockTypeCard:
		resource = modconfig.NewDashboardCard(block, mod, blockName)
	case modconfig.BlockTypeHierarchy:
		resource = modconfig.NewDashboardHierarchy(block, mod, blockName)
	case modconfig.BlockTypeImage:
		resource = modconfig.NewDashboardImage(block, mod, blockName)
	case modconfig.BlockTypeInput:
		resource = modconfig.NewDashboardInput(block, mod, blockName)
	case modconfig.BlockTypeTable:
		resource = modconfig.NewDashboardTable(block, mod, blockName)
	case modconfig.BlockTypeText:
		resource = modconfig.NewDashboardText(block, mod, blockName)
	default:
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("resourceForBlock called for unsupported block type %s", block.Type),
			Subject:  &block.DefRange,
		},
		}
	}
	return resource, nil
}

func decodeLocals(block *hcl.Block, runCtx *RunContext) ([]*modconfig.Local, *decodeResult) {
	res := &decodeResult{}
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
		val, diags := attr.Expr.Value(runCtx.EvalCtx)
		// handle any resulting diags, which may specify dependencies
		res.handleDecodeDiags(diags)

		// add to our list
		locals = append(locals, modconfig.NewLocal(name, val, attr.Range, runCtx.CurrentMod))
	}
	return locals, res
}

func decodeVariable(block *hcl.Block, runCtx *RunContext) (*modconfig.Variable, *decodeResult) {
	res := &decodeResult{}

	var variable *modconfig.Variable
	content, diags := block.Body.Content(VariableBlockSchema)
	res.handleDecodeDiags(diags)

	v, diags := var_config.DecodeVariableBlock(block, content, false)
	res.handleDecodeDiags(diags)

	if res.Success() {
		variable = modconfig.NewVariable(v, runCtx.CurrentMod)
	}

	return variable, res

}

func decodeParam(block *hcl.Block, runCtx *RunContext, parentName string) (*modconfig.ParamDef, hcl.Diagnostics) {
	def := modconfig.NewParamDef(block)

	content, diags := block.Body.Content(ParamDefBlockSchema)

	if attr, exists := content.Attributes["description"]; exists {
		moreDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &def.Description)
		diags = append(diags, moreDiags...)
	}
	if attr, exists := content.Attributes["default"]; exists {
		v, moreDiags := attr.Expr.Value(runCtx.EvalCtx)
		diags = append(diags, moreDiags...)

		if !moreDiags.HasErrors() {
			// convert the raw default into a postgres representation
			if valStr, err := ctyToPostgresString(v); err == nil {
				def.Default = utils.ToStringPointer(valStr)
			} else {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("%s has invalid parameter config", parentName),
					Detail:   err.Error(),
					Subject:  &attr.Range,
				})
			}
		}
	}
	return def, diags
}

func decodeQueryProvider(block *hcl.Block, runCtx *RunContext) (modconfig.HclResource, *decodeResult) {
	res := &decodeResult{}

	// get shell resource
	resource, diags := resourceForBlock(block, runCtx)
	res.handleDecodeDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	// do a partial decode using QueryProviderBlockSchema
	// this will be used to pull out attributes which need manual decoding
	content, remain, diags := block.Body.PartialContent(QueryProviderBlockSchema)
	res.handleDecodeDiags(diags)
	if !res.Success() {
		return nil, res
	}

	// handle invalid block types
	res.addDiags(validateBlocks(remain.(*hclsyntax.Body), QueryProviderBlockSchema, resource))

	// decode the body into 'resource' to populate all properties that can be automatically decoded
	diags = gohcl.DecodeBody(remain, runCtx.EvalCtx, resource)
	// handle any resulting diags, which may specify dependencies
	res.handleDecodeDiags(diags)

	// cast resource to a QueryProvider
	queryProvider, ok := resource.(modconfig.QueryProvider)
	if !ok {
		// coding error
		panic(fmt.Sprintf("block type %s not convertible to a QueryProvider", block.Type))
	}

	sqlAttr, sqlPropertySet := content.Attributes["sql"]
	_, queryPropertySet := content.Attributes["query"]

	if sqlPropertySet && queryPropertySet {
		// either Query or SQL property may be set -  if Query property already set, error
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s has both 'SQL' and 'query' property set - only 1 of these may be set", resource.Name()),
			Subject:  &sqlAttr.Range,
		})

		res.addDiags(diags)
	}

	if attr, exists := content.Attributes["args"]; exists {
		args, runtimeDependencies, diags := decodeArgs(attr, runCtx.EvalCtx, queryProvider)
		if diags.HasErrors() {
			// handle dependencies
			res.handleDecodeDiags(diags)
		} else {
			queryProvider.SetArgs(args)
			queryProvider.AddRuntimeDependencies(runtimeDependencies)
		}

	}

	var params []*modconfig.ParamDef
	for _, block := range content.Blocks {
		// only paramdefs ar defined in the schema
		if block.Type != modconfig.BlockTypeParam {
			panic(fmt.Sprintf("invalid child block type %s", block.Type))
		}

		// param block cannot be set if a query property is set - it is only valid if inline SQL ids defined
		if queryPropertySet {
			diags = append(diags, invalidParamDiags(resource, block))
		}
		paramDef, moreDiags := decodeParam(block, runCtx, resource.Name())
		if !moreDiags.HasErrors() {
			params = append(params, paramDef)
			// add and references contained in the param block to the control refs
			moreDiags = AddReferences(resource, block, runCtx)
		}
		diags = append(diags, moreDiags...)
	}

	queryProvider.SetParams(params)

	// handle any resulting diags, which may specify dependencies
	res.handleDecodeDiags(diags)

	return resource, res
}

func invalidParamDiags(resource modconfig.HclResource, block *hcl.Block) *hcl.Diagnostic {
	return &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("%s has 'query' property set so cannot define param blocks", resource.Name()),
		Subject:  &block.DefRange,
	}
}

func decodeDashboard(block *hcl.Block, runCtx *RunContext) (*modconfig.Dashboard, *decodeResult) {
	res := &decodeResult{}
	dashboard := modconfig.NewDashboard(block, runCtx.CurrentMod, runCtx.DetermineBlockName(block))

	// do a partial decode using QueryProviderBlockSchema
	// this will be used to pull out attributes which need manual decoding
	content, remain, diags := block.Body.PartialContent(DashboardBlockSchema)
	res.handleDecodeDiags(diags)

	// handle invalid block types
	res.addDiags(validateBlocks(remain.(*hclsyntax.Body), DashboardBlockSchema, dashboard))
	if !res.Success() {
		return nil, res
	}

	// decode the body into 'dashboardContainer' to populate all properties that can be automatically decoded
	diags = gohcl.DecodeBody(remain, runCtx.EvalCtx, dashboard)
	// handle any resulting diags, which may specify dependencies
	res.handleDecodeDiags(diags)

	if dashboard.Base != nil && len(dashboard.Base.ChildNames) > 0 {
		supportedChildren := []string{modconfig.BlockTypeContainer, modconfig.BlockTypeChart, modconfig.BlockTypeControl, modconfig.BlockTypeCard, modconfig.BlockTypeHierarchy, modconfig.BlockTypeImage, modconfig.BlockTypeInput, modconfig.BlockTypeTable, modconfig.BlockTypeText}
		// TODO: we should be passing in the block for the Base resource - but this is only used for diags
		// and we do not expect to get any (as this function has already succeeded when the base was originally parsed)
		children, _ := resolveChildrenFromNames(dashboard.Base.ChildNames, block, supportedChildren, runCtx)
		dashboard.Base.SetChildren(children)
	}
	if !res.Success() {
		return dashboard, res
	}

	// now decode child blocks
	if len(content.Blocks) > 0 {
		blocksRes := decodeDashboardBlocks(content, dashboard, runCtx)
		res.Merge(blocksRes)
	}

	return dashboard, res
}

func decodeDashboardBlocks(content *hcl.BodyContent, dashboard *modconfig.Dashboard, runCtx *RunContext) *decodeResult {
	var res = &decodeResult{}
	var inputs []*modconfig.DashboardInput

	// set dashboard as parent on the run context - this is used when generating names for anonymous blocks
	runCtx.PushParent(dashboard)
	defer func() {
		runCtx.PopParent()
	}()

	for _, b := range content.Blocks {
		// decode block
		resources, blockRes := decodeBlock(b, runCtx)
		res.Merge(blockRes)
		if !blockRes.Success() {
			continue
		}

		// we expect either inputs or child report nodes
		for _, resource := range resources {
			if b.Type == modconfig.BlockTypeInput {
				input := resource.(*modconfig.DashboardInput)
				// add dashboard name to input
				input.SetDashboard(dashboard)
				inputs = append(inputs, input)
				dashboard.AddChild(input)
				// inputs get added to the mod in SetInputs
			} else {
				// add the resource to the mod
				res.addDiags(addResourcesToMod(runCtx, resource))
				// add to the dashboard children
				// (we expect this cast to always succeed)
				if child, ok := resource.(modconfig.ModTreeItem); ok {
					dashboard.AddChild(child)
				}
			}
		}
	}

	err := dashboard.SetInputs(inputs)
	if err != nil {
		res.addDiags(hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Duplicate input names",
			Detail:   err.Error(),
			Subject:  &dashboard.DeclRange,
		}})
	}

	return res
}

func decodeDashboardContainer(block *hcl.Block, runCtx *RunContext) (*modconfig.DashboardContainer, *decodeResult) {
	res := &decodeResult{}
	container := modconfig.NewDashboardContainer(block, runCtx.CurrentMod, runCtx.DetermineBlockName(block))

	// do a partial decode using QueryProviderBlockSchema
	// this will be used to pull out attributes which need manual decoding
	content, remain, diags := block.Body.PartialContent(DashboardContainerBlockSchema)
	res.handleDecodeDiags(diags)
	if !res.Success() {
		return nil, res
	}

	// handle invalid block types
	res.addDiags(validateBlocks(remain.(*hclsyntax.Body), DashboardContainerBlockSchema, container))

	// decode the body into 'dashboardContainer' to populate all properties that can be automatically decoded
	diags = gohcl.DecodeBody(remain, runCtx.EvalCtx, container)
	// handle any resulting diags, which may specify dependencies
	res.handleDecodeDiags(diags)

	// now decode child blocks
	if len(content.Blocks) > 0 {
		blocksRes := decodeDashboardContainerBlocks(content, container, runCtx)
		res.Merge(blocksRes)
	}

	return container, res
}

func decodeDashboardContainerBlocks(content *hcl.BodyContent, dashboardContainer *modconfig.DashboardContainer, runCtx *RunContext) *decodeResult {
	var res = &decodeResult{}
	var inputs []*modconfig.DashboardInput
	// set container as parent on the run context - this is used when generating names for anonymous blocks
	runCtx.PushParent(dashboardContainer)
	defer func() {
		runCtx.PopParent()
	}()

	for _, b := range content.Blocks {
		// use generic block decoding
		resources, blockRes := decodeBlock(b, runCtx)
		res.Merge(blockRes)
		if !blockRes.Success() {
			continue
		}

		// all blocks are child report nodes
		for _, resource := range resources {
			if b.Type == modconfig.BlockTypeInput {
				input := resource.(*modconfig.DashboardInput)
				// add dashboard name to input
				inputs = append(inputs, input)
				// now we have namespaced the input, add to mod
				res.addDiags(runCtx.CurrentMod.AddResource(resource))
			}

			// add the resource to the mod
			addResourcesToMod(runCtx, resource)

			if child, ok := resource.(modconfig.ModTreeItem); ok {
				dashboardContainer.AddChild(child)
			}
		}
	}

	return res
}

func decodeBenchmark(block *hcl.Block, runCtx *RunContext) (*modconfig.Benchmark, *decodeResult) {
	res := &decodeResult{}

	benchmark := modconfig.NewBenchmark(block, runCtx.CurrentMod, runCtx.DetermineBlockName(block))
	content, diags := block.Body.Content(BenchmarkBlockSchema)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "children", &benchmark.ChildNames, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "description", &benchmark.Description, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "documentation", &benchmark.Documentation, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "tags", &benchmark.Tags, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "title", &benchmark.Title, runCtx)
	res.handleDecodeDiags(diags)

	// now add children
	if res.Success() {
		supportedChildren := []string{modconfig.BlockTypeBenchmark, modconfig.BlockTypeControl}
		children, diags := resolveChildrenFromNames(benchmark.ChildNames.StringList(), block, supportedChildren, runCtx)
		res.handleDecodeDiags(diags)

		// now set children and child name strings
		benchmark.Children = children
		benchmark.ChildNameStrings = getChildNameStringsFromModTreeItem(children)
	}

	// decode report specific properties
	diags = decodeProperty(content, "base", &benchmark.Base, runCtx)
	res.handleDecodeDiags(diags)
	if benchmark.Base != nil && len(benchmark.Base.ChildNames) > 0 {
		supportedChildren := []string{modconfig.BlockTypeBenchmark, modconfig.BlockTypeControl}
		// TODO: we should be passing in the block for the Base resource - but this is only used for diags
		// and we do not expect to get any (as this function has already succeeded when the base was originally parsed)
		children, _ := resolveChildrenFromNames(benchmark.Base.ChildNameStrings, block, supportedChildren, runCtx)
		benchmark.Base.Children = children
	}
	diags = decodeProperty(content, "width", &benchmark.Width, runCtx)
	res.handleDecodeDiags(diags)
	return benchmark, res
}

func decodeProperty(content *hcl.BodyContent, property string, dest interface{}, runCtx *RunContext) hcl.Diagnostics {
	var diags hcl.Diagnostics
	if title, ok := content.Attributes[property]; ok {
		diags = gohcl.DecodeExpression(title.Expr, runCtx.EvalCtx, dest)
	}
	return diags
}

// handleDecodeResult
// if decode was successful:
// - generate and set resource metadata
// - add resource to RunContext (which adds it to the mod)handleDecodeResult
func handleDecodeResult(resource modconfig.HclResource, res *decodeResult, block *hcl.Block, runCtx *RunContext) {
	if res.Success() {
		anonymousResource := resourceIsAnonymous(resource)

		// call post decode hook
		// NOTE: must do this BEFORE adding resource to run context to ensure we respect the base property
		moreDiags := resource.OnDecoded(block, runCtx)
		res.addDiags(moreDiags)

		// add references
		moreDiags = AddReferences(resource, block, runCtx)
		res.addDiags(moreDiags)

		// if resource is NOT anonymous, add into the run context
		if !anonymousResource {
			moreDiags = runCtx.AddResource(resource)
			res.addDiags(moreDiags)
		}

		// if resource supports metadata, save it
		if resourceWithMetadata, ok := resource.(modconfig.ResourceWithMetadata); ok {
			body := block.Body.(*hclsyntax.Body)
			moreDiags = addResourceMetadata(resourceWithMetadata, body.SrcRange, runCtx)
			res.addDiags(moreDiags)
		}
	} else {
		if len(res.Depends) > 0 {
			moreDiags := runCtx.AddDependencies(block, resource.GetUnqualifiedName(), res.Depends)
			res.addDiags(moreDiags)
		}
	}
}

func resourceIsAnonymous(resource modconfig.HclResource) bool {
	// (if a resource anonymous it must support ResourceWithMetadata)
	resourceWithMetadata, ok := resource.(modconfig.ResourceWithMetadata)
	anonymousResource := ok && resourceWithMetadata.IsAnonymous()
	return anonymousResource
}

func addResourceMetadata(resourceWithMetadata modconfig.ResourceWithMetadata, srcRange hcl.Range, runCtx *RunContext) hcl.Diagnostics {
	metadata, err := GetMetadataForParsedResource(resourceWithMetadata.Name(), srcRange, runCtx.FileData, runCtx.CurrentMod)
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

// Validate all blocks are supported
// We use partial decoding so that we can automatically decode as many properties as possible
// and only manually decode properties requiring special logic.
// The problem is the partial decode does not return errors for invalid blocks, so we must implement our own
func validateBlocks(body *hclsyntax.Body, schema *hcl.BodySchema, resource modconfig.HclResource) hcl.Diagnostics {
	var diags hcl.Diagnostics

	// identify any blocks specified by hcl tags
	var supportedBlocks []string
	v := reflect.TypeOf(helpers.DereferencePointer(resource))
	for i := 0; i < v.NumField(); i++ {
		tag := v.FieldByIndex([]int{i}).Tag.Get("hcl")
		if idx := strings.LastIndex(tag, ",block"); idx != -1 {
			supportedBlocks = append(supportedBlocks, tag[:idx])
		}
	}
	// ad din blocks specified in the schema
	for _, b := range schema.Blocks {
		supportedBlocks = append(supportedBlocks, b.Type)
	}

	// now check for invalid blocks
	for _, block := range body.Blocks {
		if !helpers.StringSliceContains(supportedBlocks, block.Type) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf(`Unsupported block type: Blocks of type "%s" are not expected here.`, block.Type),
				Subject:  &block.TypeRange,
			})
		}
	}

	return diags
}
