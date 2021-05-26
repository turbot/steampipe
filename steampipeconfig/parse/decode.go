package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
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
			locals, res := decodeLocals(block, runCtx.EvalCtx)
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
	switch modconfig.ModBlockType(block.Type) {
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

func decodeLocals(block *hcl.Block, ctx *hcl.EvalContext) ([]*modconfig.Local, *decodeResult) {
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
		val, diags := attr.Expr.Value(ctx)
		// handle any resulting diags, which may specify dependencies
		res.handleDecodeDiags(diags)

		// add to our list
		locals = append(locals, modconfig.NewLocal(name, val, attr.Range))
	}
	return locals, res
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
		resource.OnDecoded(block)
		AddReferences(resource, block)
	}
	return resource, res
}

func decodePanel(block *hcl.Block, runCtx *RunContext) (*modconfig.Panel, *decodeResult) {
	res := &decodeResult{}
	content, diags := block.Body.Content(PanelSchema)
	res.handleDecodeDiags(diags)

	// get shell resource
	panel := modconfig.NewPanel(block)

	diags = decodeProperty(content, "title", &panel.Title, runCtx)
	res.handleDecodeDiags(diags)

	diags = decodeProperty(content, "width", &panel.Width, runCtx)
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

	content, diags := block.Body.Content(ReportSchema)
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

// if the diags contains dependency errors, add dependencies to the result
// otherwise add diags to the result
func (res *decodeResult) handleDecodeDiags(diags hcl.Diagnostics) {
	for _, diag := range diags {
		if dependency := isDependencyError(diag); dependency != nil {
			// was this error caused by a missing dependency?
			res.Depends = append(res.Depends, dependency)
		}
	}
	// only register errors if there are NOT any missing variables
	if diags.HasErrors() && len(res.Depends) == 0 {
		res.Diags = append(res.Diags, diags...)
	}
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
		}
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
