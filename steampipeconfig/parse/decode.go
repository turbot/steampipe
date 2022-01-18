package parse

import (
	"fmt"

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
	runCtx.ClearDependencies()
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to determine required dependency order",
			Detail:   err.Error()})
	}
	for _, block := range blocks {
		_, res := decodeBlock(runCtx, block)
		if !res.Success() {
			diags = append(diags, res.Diags...)
			continue
		}
	}

	return diags
}

func decodeBlock(runCtx *RunContext, block *hcl.Block) ([]modconfig.HclResource, *decodeResult) {
	var resource modconfig.HclResource
	var resources []modconfig.HclResource
	var res = &decodeResult{}

	// if opts specifies block types, check whether this type is included
	if !runCtx.ShouldIncludeBlock(block) {
		return nil, res
	}
	// if this block is anonymous, give it a unique name
	anonymousBlock := runCtx.IsBlockAnonymous(block)
	if anonymousBlock {
		// replace the labels
		block.Labels = []string{runCtx.GetAnonymousResourceName(block)}
	}
	runCtx.PushDecodeBlock(block)
	defer runCtx.PopDecodeBlock()

	// check name is valid
	diags := validateName(block)
	if diags.HasErrors() {
		res.addDiags(diags)
		return nil, res
	}

	switch block.Type {
	case modconfig.BlockTypeLocals:
		// special case decode logic for locals
		var locals []*modconfig.Local
		locals, res = decodeLocals(block, runCtx)
		for _, local := range locals {
			resources = append(resources, local)
		}

	case modconfig.BlockTypePanel:
		resource, res = decodePanel(block, runCtx)
		resources = append(resources, resource)
	case modconfig.BlockTypeContainer, modconfig.BlockTypeReport:
		resource, res = decodeContainerReport(block, runCtx)
		resources = append(resources, resource)
	case modconfig.BlockTypeVariable:
		resource, res = decodeVariable(block, runCtx)
		resources = append(resources, resource)
	case modconfig.BlockTypeControl:
		resource, res = decodeControl(block, runCtx)
		resources = append(resources, resource)
	case modconfig.BlockTypeBenchmark:
		resource, res = decodeBenchmark(block, runCtx)
		resources = append(resources, resource)
	case modconfig.BlockTypeQuery:
		resource, res = decodeQuery(block, runCtx)
		resources = append(resources, resource)
	default:
		// all other blocks are treated the same:
		resource, res = decodeResource(block, runCtx)
		resources = append(resources, resource)
	}

	for _, resource := range resources {
		// if this resource type implements AnonymousResource, set its anonymous status
		if anon, ok := resource.(modconfig.AnonymousResource); ok {
			anon.SetAnonymous(anonymousBlock)
		}
		// handle the result
		// - if successful, add resource to mod and variables maps
		// - if there are dependencies, add them to run context
		handleDecodeResult(resource, res, block, runCtx)
	}

	return resources, res
}

// generic decode function for any resource we do not have custom decode logic for
func decodeResource(block *hcl.Block, runCtx *RunContext) (modconfig.HclResource, *decodeResult) {
	res := &decodeResult{}
	// get shell resource
	resource, diags := resourceForBlock(block, runCtx)
	// handle any resulting diags, which may specify dependencies
	res.handleDecodeDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	diags = gohcl.DecodeBody(block.Body, runCtx.EvalCtx, resource)
	// handle any resulting diags, which may specify dependencies
	res.handleDecodeDiags(diags)

	return resource, res
}

// return a shell resource for the given block
func resourceForBlock(block *hcl.Block, runCtx *RunContext) (modconfig.HclResource, hcl.Diagnostics) {
	var resource modconfig.HclResource
	switch block.Type {
	case modconfig.BlockTypeMod:
		// runCtx already contains the current mod
		resource = runCtx.CurrentMod
	case modconfig.BlockTypeQuery:
		resource = modconfig.NewQuery(block)
	case modconfig.BlockTypeControl:
		resource = modconfig.NewControl(block)
	case modconfig.BlockTypeContainer:
		resource = modconfig.NewReportContainer(block)
	case modconfig.BlockTypeReport:
		resource = modconfig.NewReportContainer(block)
	case modconfig.BlockTypePanel:
		resource = modconfig.NewPanel(block)
	case modconfig.BlockTypeBenchmark:
		resource = modconfig.NewBenchmark(block)
	default:
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("resourceForBlock called for unsupported block type %s"),
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
		locals = append(locals, modconfig.NewLocal(name, val, attr.Range))
	}
	return locals, res
}

func decodeVariable(block *hcl.Block, runCtx *RunContext) (*modconfig.Variable, *decodeResult) {
	res := &decodeResult{}

	var variable *modconfig.Variable
	v, diags := var_config.DecodeVariableBlock(block, false)
	// handle any resulting diags, which may specify dependencies
	res.handleDecodeDiags(diags)

	// call post-decode hook
	if res.Success() {
		variable = modconfig.NewVariable(v)
	}

	return variable, res

}

func decodeQuery(block *hcl.Block, runCtx *RunContext) (*modconfig.Query, *decodeResult) {
	res := &decodeResult{}

	q := modconfig.NewQuery(block)

	content, diags := block.Body.Content(QueryBlockSchema)

	if !hclsyntax.ValidIdentifier(q.ShortName) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid control name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[0],
		})
	}

	if attr, exists := content.Attributes["description"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &q.Description)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["documentation"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &q.Documentation)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["search_path"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &q.SearchPath)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["search_path_prefix"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &q.SearchPathPrefix)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["sql"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &q.SQL)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["tags"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &q.Tags)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["title"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &q.Title)
		diags = append(diags, valDiags...)
	}
	for _, block := range content.Blocks {
		if block.Type == modconfig.BlockTypeParam {
			param, moreDiags := decodeParam(block, runCtx, q.FullName)
			if !moreDiags.HasErrors() {
				q.Params = append(q.Params, param)
				// also add references to query
				moreDiags = AddReferences(q, block, runCtx)
			}
			diags = append(diags, moreDiags...)

		}
	}

	// handle any resulting diags, which may specify dependencies
	res.handleDecodeDiags(diags)

	return q, res
}

func decodeParam(block *hcl.Block, runCtx *RunContext, parentName string) (*modconfig.ParamDef, hcl.Diagnostics) {
	def := modconfig.NewParamDef(block, parentName)

	content, diags := block.Body.Content(ParamDefBlockSchema)

	if attr, exists := content.Attributes["description"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &def.Description)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["default"]; exists {
		v, diags := attr.Expr.Value(runCtx.EvalCtx)
		if diags.HasErrors() {
			return nil, diags
		}
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
	return def, diags

}

func decodeControl(block *hcl.Block, runCtx *RunContext) (*modconfig.Control, *decodeResult) {
	res := &decodeResult{}

	c := modconfig.NewControl(block)

	content, diags := block.Body.Content(ControlBlockSchema)

	if !hclsyntax.ValidIdentifier(c.ShortName) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid control name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[0],
		})
	}

	if attr, exists := content.Attributes["description"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &c.Description)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["documentation"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &c.Documentation)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["search_path"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &c.SearchPath)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["search_path_prefix"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &c.SearchPathPrefix)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["severity"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &c.Severity)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["sql"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &c.SQL)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["query"]; exists {
		// either Query or SQL property may be set -  if Query property already set, error
		if c.SQL != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("%s has both 'SQL' and 'query' property set - only 1 of these may be set", c.FullName),
				Subject:  &attr.Range,
			})
		} else {
			valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &c.Query)
			diags = append(diags, valDiags...)
		}
	}

	if attr, exists := content.Attributes["tags"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &c.Tags)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["title"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &c.Title)
		diags = append(diags, valDiags...)
	}
	if attr, exists := content.Attributes["args"]; exists {
		if params, diags := decodeControlArgs(attr, runCtx.EvalCtx, c.FullName); !diags.HasErrors() {
			c.Args = params
		}
	}

	for _, block := range content.Blocks {
		if block.Type == modconfig.BlockTypeParam {
			// param block cannot be set if a query property is set - it is only valid if inline SQL ids defined
			if c.Query != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("%s has 'query' property set so cannot define param blocks", c.FullName),
					Subject:  &block.DefRange,
				})
			}
			paramDef, moreDiags := decodeParam(block, runCtx, c.FullName)
			if !moreDiags.HasErrors() {
				c.Params = append(c.Params, paramDef)
				// add and references contained in the param block to the control refs
				moreDiags = AddReferences(c, block, runCtx)
			}
			diags = append(diags, moreDiags...)
		}
	}

	// verify the control has either a query or a sql attribute
	if c.Query == nil && c.SQL == nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s must define either a 'sql' property or a 'query' property", c.FullName),
			Subject:  &block.DefRange,
		})
	}

	// handle any resulting diags, which may specify dependencies
	res.handleDecodeDiags(diags)

	return c, res
}

func decodeControlArgs(attr *hcl.Attribute, evalCtx *hcl.EvalContext, controlName string) (*modconfig.QueryArgs, hcl.Diagnostics) {
	var params = modconfig.NewQueryArgs()
	v, diags := attr.Expr.Value(evalCtx)
	if diags.HasErrors() {
		return nil, diags
	}

	var err error
	ty := v.Type()

	switch {
	case ty.IsObjectType():
		params.Args, err = ctyObjectToMapOfPgStrings(v)
	case ty.IsTupleType():
		params.ArgsList, err = ctyTupleToArrayOfPgStrings(v)
	default:
		err = fmt.Errorf("'params' property must be either a map or an array")
	}

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s has invalid parameter config", controlName),
			Detail:   err.Error(),
			Subject:  &attr.Range,
		})
	}
	return params, diags
}

func decodeContainerReport(block *hcl.Block, runCtx *RunContext) (*modconfig.ReportContainer, *decodeResult) {
	res := &decodeResult{}

	content, diags := block.Body.Content(ReportBlockSchema)
	res.handleDecodeDiags(diags)

	report := modconfig.NewReportContainer(block)
	diags = decodeProperty(content, "title", &report.Title, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "base", &report.Base, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "width", &report.Width, runCtx)
	res.handleDecodeDiags(diags)

	// now add children
	if !res.Success() {
		return report, res
	}

	var children []modconfig.ModTreeItem
	if len(content.Blocks) > 0 {
		var childrenRes *decodeResult
		children, childrenRes = decodeInlineChildren(content, runCtx)
		// now set children
		report.SetChildren(children)
		res.Merge(childrenRes)
	}

	return report, res
}

func decodeBenchmark(block *hcl.Block, runCtx *RunContext) (*modconfig.Benchmark, *decodeResult) {
	res := &decodeResult{}

	content, diags := block.Body.Content(BenchmarkBlockSchema)
	res.handleDecodeDiags(diags)

	benchmark := modconfig.NewBenchmark(block)

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
		children, diags := decodeChildren(benchmark.ChildNames, block, supportedChildren, runCtx)
		res.handleDecodeDiags(diags)

		// now set children and child name strings
		benchmark.Children = children
		benchmark.ChildNameStrings = getChildNameString(children)
	}
	return benchmark, res
}

func decodePanel(block *hcl.Block, runCtx *RunContext) (*modconfig.Panel, *decodeResult) {
	res := &decodeResult{}
	content, rest, diags := block.Body.PartialContent(PanelBlockSchema)

	// get shell resource
	panel := modconfig.NewPanel(block)

	// populate dynamic properties
	diags = populatePanelDynamicProperties(rest, panel, runCtx)

	diags = decodeProperty(content, "title", &panel.Title, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "type", &panel.Type, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "width", &panel.Width, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "sql", &panel.SQL, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "base", &panel.Base, runCtx)
	res.handleDecodeDiags(diags)

	return panel, res
}

func populatePanelDynamicProperties(rest hcl.Body, panel *modconfig.Panel, runCtx *RunContext) hcl.Diagnostics {
	var diags hcl.Diagnostics
	b, ok := rest.(*hclsyntax.Body)
	if ok {
		// build map of extra attributes
		panelAttributesMap := PanelBlockSchemaAttributesMap()
		for _, attr := range b.Attributes {
			if !panelAttributesMap[attr.Name] {
				var val string
				moreDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &val)

				if moreDiags.HasErrors() {
					diags = append(diags, moreDiags...)
					continue
				}
				panel.Properties[attr.Name] = val
			}
		}
	}
	return diags
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
func handleDecodeResult(resource modconfig.HclResource, res *decodeResult, block *hcl.Block, runCtx *RunContext) hcl.Diagnostics {
	anonymousResource := len(block.Labels) == 0
	if res.Success() {
		// call post decode hook
		// NOTE: must do this BEFORE adding resource to run context to ensure we respect the base property
		moreDiags := resource.OnDecoded(block)
		res.addDiags(moreDiags)

		// add references
		moreDiags = AddReferences(resource, block, runCtx)
		res.addDiags(moreDiags)

		// if resource is NOT a mod, set mod pointer on hcl resource and add resource to current mod
		if _, ok := resource.(*modconfig.Mod); !ok {
			resource.SetMod(runCtx.CurrentMod)
			// if resource is NOT anonymous, add resource to mod
			// - this will fail if the mod already has a resource with the same name
			if !anonymousResource {
				// we cannot add anonymous resources at this point - they will be added after their names are set
				moreDiags = runCtx.CurrentMod.AddResource(resource)
				res.addDiags(moreDiags)
			}
		}

		// if resource supports metadata, save it
		if resourceWithMetadata, ok := resource.(modconfig.ResourceWithMetadata); ok {
			body := block.Body.(*hclsyntax.Body)
			moreDiags = addResourceMetadata(resourceWithMetadata, body.SrcRange, runCtx)
			res.addDiags(moreDiags)
		}

		// // if resource is NOT anonymous, add into the run context
		if !anonymousResource {
			moreDiags = runCtx.AddResource(resource)
			res.addDiags(moreDiags)
		}
	} else {
		if len(res.Depends) > 0 {
			runCtx.AddDependencies(block, resource.Name(), res.Depends)
		}
	}
	// return the diags
	return res.Diags
}

func addResourceMetadata(resourceWithMetadata modconfig.ResourceWithMetadata, srcRange hcl.Range, runCtx *RunContext) hcl.Diagnostics {
	var diags hcl.Diagnostics
	metadata, err := GetMetadataForParsedResource(resourceWithMetadata.Name(), srcRange, runCtx.FileData, runCtx.CurrentMod)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err.Error(),
			Subject:  &srcRange,
		})
	} else {
		resourceWithMetadata.SetMetadata(metadata)
	}
	return diags
}

// determine whether the diag is a dependency error, and if so, return a dependency object
func isDependencyError(diag *hcl.Diagnostic) *dependency {
	if helpers.StringSliceContains(missingVariableErrors, diag.Summary) {
		return &dependency{diag.Expression.Range(), diag.Expression.Variables()}
	}
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
