package parse

import (
	"fmt"

	"github.com/turbot/steampipe/utils"

	typehelpers "github.com/turbot/go-kit/types"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig/var_config"
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

func decode(runCtx *RunContext, opts *ParseModOptions) hcl.Diagnostics {
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
		// if opts specifies block types, check whether this type is included
		if len(opts.BlockTypes) > 0 && !helpers.StringSliceContains(opts.BlockTypes, block.Type) {
			continue
		}
		// check name is valid
		moreDiags := validateName(block)
		if diags.HasErrors() {
			diags = append(diags, moreDiags...)
			continue
		}

		//var decodeResults []*decodeResult
		// special case decoding for locals
		switch block.Type {
		case modconfig.BlockTypeLocals:
			// special case decode logic for locals
			locals, res := decodeLocals(block, runCtx)
			for _, local := range locals {
				// handle the result
				// - if successful, add resource to mod and variables maps
				// - if there are dependencies, add them to run context
				moreDiags = handleDecodeResult(local, res, block, runCtx)
				diags = append(diags, moreDiags...)
			}
		case modconfig.BlockTypePanel:
			// special case decode logic for locals
			panel, res := decodePanel(block, runCtx)
			moreDiags = handleDecodeResult(panel, res, block, runCtx)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
			}
		case modconfig.BlockTypeReport:
			// special case decode logic for locals
			report, res := decodeReport(block, runCtx)
			moreDiags = handleDecodeResult(report, res, block, runCtx)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
			}
		case modconfig.BlockTypeVariable:
			// special case decode logic for locals
			variable, res := decodeVariable(block, runCtx)
			moreDiags = handleDecodeResult(variable, res, block, runCtx)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
			}
		case modconfig.BlockTypeControl:
			// special case decode logic for locals
			control, res := decodeControl(block, runCtx)
			moreDiags = handleDecodeResult(control, res, block, runCtx)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
			}
		default:
			// all other blocks are treated the same:
			resource, res := decodeResource(block, runCtx)
			moreDiags = handleDecodeResult(resource, res, block, runCtx)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
			}
		}
	}
	return diags
}

// return a shell resource for the given block
func resourceForBlock(block *hcl.Block, runCtx *RunContext) modconfig.HclResource {
	var resource modconfig.HclResource
	switch block.Type {
	case modconfig.BlockTypeMod:
		// runCtx already contains the shell mod
		resource = runCtx.Mod
	case modconfig.BlockTypeQuery:
		resource = modconfig.NewQuery(block)
	case modconfig.BlockTypeControl:
		resource = modconfig.NewControl(block)
	case modconfig.BlockTypeReport:
		resource = modconfig.NewReport(block)
	case modconfig.BlockTypePanel:
		resource = modconfig.NewPanel(block)
	case modconfig.BlockTypeBenchmark:
		resource = modconfig.NewBenchmark(block)
	}
	return resource
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

		if diags := variable.OnDecoded(block); diags.HasErrors() {
			res.addDiags(diags)
		}
		// TODO for now we do not implement storing references for variables
		// - special case code is required to exclude types from the reference detection code
		//AddReferences(variable, block)
	}

	return variable, res

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
		// either Query or SQL property may be set -  if Query property already set, error
		if c.Query != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("%s has both 'SQL' and 'query' property set - only 1 of these may be set", c.FullName),
				Subject:  &attr.Range,
			})
		} else {
			valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &c.SQL)
			diags = append(diags, valDiags...)
		}

	}
	if attr, exists := content.Attributes["query"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, runCtx.EvalCtx, &c.Query)
		diags = append(diags, valDiags...)
		if !valDiags.HasErrors() {
			// either Query or SQL property may be set - if SQL property already set, error
			if typehelpers.SafeString(c.SQL) != "" {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("%s has both 'SQL' and 'query' property set - only 1 of these may be set", c.FullName),
					Subject:  &attr.Range,
				})
			} else {
				// set SQL to the query NAME
				c.SQL = utils.ToStringPointer(c.Query.FullName)
			}
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
	if attr, exists := content.Attributes["params"]; exists {
		valDiags := setControlParams(attr, runCtx.EvalCtx, c)
		diags = append(diags, valDiags...)

	}

	// handle any resulting diags, which may specify dependencies
	res.handleDecodeDiags(diags)

	// call post-decode hook
	if res.Success() {
		if diags := c.OnDecoded(block); diags.HasErrors() {
			res.addDiags(diags)
		}
		AddReferences(c, block)
	}

	return c, res

}

func setControlParams(attr *hcl.Attribute, evalCtx *hcl.EvalContext, c *modconfig.Control) hcl.Diagnostics {
	v, diags := attr.Expr.Value(evalCtx)
	if diags.HasErrors() {
		return diags
		//if err != nil {
		//	diags = append(diags, &hcl.Diagnostic{Severity: hcl.DiagError, Summary: err.Error()})
		//}
	}
	ty := v.Type()
	var err error
	switch {
	case ty.IsObjectType():
		c.Params.Params, err = ctyObjectToPostgresMap(v)
	case ty.IsTupleType():
		c.Params.ParamsList, err = ctyTupleToPostgresArray(v)
	default:
		err = fmt.Errorf("'params' property must be either a map or an array")
	}

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s has invalid parameter config", c.FullName),
			Detail:   err.Error(),
			Subject:  &attr.Range,
		})
	}

	return diags
}

func decodeResource(block *hcl.Block, runCtx *RunContext) (modconfig.HclResource, *decodeResult) {
	// get shell resource
	resource := resourceForBlock(block, runCtx)

	res := &decodeResult{}
	diags := gohcl.DecodeBody(block.Body, runCtx.EvalCtx, resource)
	// handle any resulting diags, which may specify dependencies
	res.handleDecodeDiags(diags)

	// call post-decode hook
	if res.Success() {
		if diags := resource.OnDecoded(block); diags.HasErrors() {
			res.addDiags(diags)
		}
		AddReferences(resource, block)
	}
	return resource, res
}

func decodePanel(block *hcl.Block, runCtx *RunContext) (*modconfig.Panel, *decodeResult) {
	res := &decodeResult{}
	content, diags := block.Body.Content(PanelBlockSchema)
	res.handleDecodeDiags(diags)

	// get shell resource
	panel := modconfig.NewPanel(block)

	diags = decodeProperty(content, "title", &panel.Title, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "type", &panel.Type, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "width", &panel.Width, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "height", &panel.Height, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "source", &panel.Source, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "text", &panel.Text, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "sql", &panel.SQL, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeReportBlocks(panel, content, runCtx)
	res.handleDecodeDiags(diags)

	return panel, res
}

func decodeReport(block *hcl.Block, runCtx *RunContext) (*modconfig.Report, *decodeResult) {
	res := &decodeResult{}

	content, diags := block.Body.Content(ReportBlockSchema)
	res.handleDecodeDiags(diags)

	report := modconfig.NewReport(block)
	diags = decodeProperty(content, "title", &report.Title, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeReportBlocks(report, content, runCtx)
	res.handleDecodeDiags(diags)

	return report, res
}

func decodeReportBlocks(resource modconfig.ModTreeItem, content *hcl.BodyContent, runCtx *RunContext) hcl.Diagnostics {
	var diags hcl.Diagnostics
	for _, b := range content.Blocks {
		var childResource modconfig.ModTreeItem
		var decodeResult *decodeResult
		switch b.Type {
		case modconfig.BlockTypePanel:
			childResource, decodeResult = decodePanel(b, runCtx)
		case modconfig.BlockTypeReport:
			childResource, decodeResult = decodeReport(b, runCtx)
		}

		// add this panel to the mod
		moreDiags := handleDecodeResult(childResource.(modconfig.HclResource), decodeResult, b, runCtx)
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		}
		if decodeResult.Success() {
			resource.AddChild(childResource)
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
	var diags hcl.Diagnostics
	if res.Success() {
		// if resource supports metadata, save it
		if resourceWithMetadata, ok := resource.(modconfig.ResourceWithMetadata); ok {
			diags = addResourceMetadata(resource, block, runCtx, diags, resourceWithMetadata)
		}
		// add resource into the run context
		moreDiags := runCtx.AddResource(resource, block)
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		}
	} else {
		if res.Diags.HasErrors() {
			diags = append(diags, res.Diags...)
		}
		if len(res.Depends) > 0 {
			runCtx.AddDependencies(block, resource.Name(), res.Depends)
		}
	}
	// update result diags
	res.Diags = diags
	return res.Diags
}

func addResourceMetadata(resource modconfig.HclResource, block *hcl.Block, runCtx *RunContext, diags hcl.Diagnostics, resourceWithMetadata modconfig.ResourceWithMetadata) hcl.Diagnostics {
	metadata, err := GetMetadataForParsedResource(resource.Name(), block, runCtx.FileData, runCtx.Mod)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err.Error(),
			Subject:  &block.DefRange,
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
