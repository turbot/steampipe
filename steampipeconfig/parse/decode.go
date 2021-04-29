package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// A consistent detail message for all "not a valid identifier" diagnostics.
const badIdentifierDetail = "A name must start with a letter or underscore and may contain only letters, digits, underscores, and dashes."
const unknownVariableError = "Unknown variable"
const missingMapElement = "Missing map element"

func decode(runCtx *RunContext) hcl.Diagnostics {
	var diags hcl.Diagnostics

	// build list of blocks to decode
	blocks, err := runCtx.BlocksToDecode()
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to determine required dependency order",
			Detail:   err.Error()})
	}
	for _, block := range blocks {
		moreDiags := validateName(block)
		if diags.HasErrors() {
			diags = append(diags, moreDiags...)
			continue
		}

		switch modconfig.ModBlockType(block.Type) {
		case modconfig.BlockTypeMod:
			// pass the shell mod - it will be mutated
			res := decodeMod(block, runCtx.Mod, runCtx.EvalCtx)
			diags = append(diags, handleDecodeResult(runCtx.Mod, res, block, runCtx)...)
			//if res.Diags.HasErrors() {
			//	diags = append(diags, res.Diags...)
			//}
			//if len(res.Depends) > 0 {
			//	runCtx.AddDependencies(block, runCtx.Mod.FullName(), res.Depends)
			//}

		case modconfig.BlockTypeQuery:
			query := &modconfig.Query{
				Name:      block.Labels[0],
				DeclRange: block.DefRange,
			}
			res := decodeResource(block, query, runCtx.EvalCtx)
			diags = append(diags, handleDecodeResult(query, res, block, runCtx)...)

		case modconfig.BlockTypeControl:
			control := &modconfig.Control{
				Name:      block.Labels[0],
				DeclRange: block.DefRange,
			}
			res := decodeResource(block, control, runCtx.EvalCtx)
			diags = append(diags, handleDecodeResult(control, res, block, runCtx)...)

		case modconfig.BlockTypeControlGroup:
			controlGroup := &modconfig.ControlGroup{
				Name:      block.Labels[0],
				DeclRange: block.DefRange,
			}
			res := decodeResource(block, controlGroup, runCtx.EvalCtx)
			diags = append(diags, handleDecodeResult(controlGroup, res, block, runCtx)...)

		case modconfig.BlockTypeLocals:
			locals, res := decodeLocals(block, runCtx.EvalCtx)
			for _, local := range locals {
				diags = append(diags, handleDecodeResult(local, res, block, runCtx)...)
			}
		}
	}
	return diags
}

func handleDecodeResult(resource modconfig.HclResource, res *decodeResult, block *hcl.Block, runCtx *RunContext) hcl.Diagnostics {
	var diags hcl.Diagnostics
	if res.Success() {
		// if resource supports metadata, save it
		if resourceWithMetadata, ok := resource.(modconfig.ResourceWithMetadata); ok {
			metadata := GetMetadataForParsedResource(resource.FullName(), block, runCtx.FileData, runCtx.Mod)
			resourceWithMetadata.SetMetadata(metadata)
		}
		moreDiags := runCtx.AddResource(resource, block)
		if diags.HasErrors() {
			diags = append(diags, moreDiags...)
		}
	} else {
		if res.Diags.HasErrors() {
			diags = append(diags, res.Diags...)
		}
		if len(res.Depends) > 0 {
			runCtx.AddDependencies(block, resource.FullName(), res.Depends)
		}
	}
	return diags
}

func decodeResource(block *hcl.Block, resource modconfig.HclResource, ctx *hcl.EvalContext) *decodeResult {
	content, diags := block.Body.Content(resource.Schema())
	if diags.HasErrors() {
		return &decodeResult{Diags: diags}
	}

	return decodeAttributes(resource, content, ctx)
}

func decodeMod(block *hcl.Block, mod *modconfig.Mod, ctx *hcl.EvalContext) *decodeResult {
	content, diags := block.Body.Content(mod.Schema())
	if diags.HasErrors() {
		return &decodeResult{Diags: diags}
	}

	res := decodeAttributes(mod, content, ctx)

	for _, block := range content.Blocks {
		switch block.Type {
		// TODO add parsing of requires block
		case "opengraph":
			opengraph := &modconfig.OpenGraph{
				DeclRange: block.DefRange,
			}
			res := decodeResource(block, opengraph, ctx)

			res.Merge(res)
			if res.Success() {
				mod.OpenGraph = opengraph
			}
		}
	}

	return res
}

func decodeLocals(block *hcl.Block, ctx *hcl.EvalContext) ([]*modconfig.Local, *decodeResult) {
	// this implemented differently
	attrs, diags := block.Body.JustAttributes()
	if len(attrs) == 0 {
		return nil, &decodeResult{Diags: diags}
	}

	locals := make([]*modconfig.Local, 0, len(attrs))
	for name, attr := range attrs {
		if !hclsyntax.ValidIdentifier(name) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid local value name",
				Detail:   badIdentifierDetail,
				Subject:  &attr.NameRange,
			})
		}

		val, moreDiags := attr.Expr.Value(ctx)
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
			continue
		}

		locals = append(locals, &modconfig.Local{
			Name:      name,
			Value:     val,
			DeclRange: attr.Range,
		})
	}
	return locals, &decodeResult{Diags: diags}
}

func decodeAttributes(resource modconfig.HclResource, content *hcl.BodyContent, ctx *hcl.EvalContext) *decodeResult {
	res := &decodeResult{}
	for _, attributeDetails := range modconfig.GetAttributeDetails(resource) {
		res.Merge(decodeAttribute(attributeDetails, content, ctx))
	}
	return res
}

func decodeAttribute(attributeDetails modconfig.AttributeDetails, content *hcl.BodyContent, ctx *hcl.EvalContext) *decodeResult {
	attribute := attributeDetails.Attribute
	dest := attributeDetails.Dest

	var diags hcl.Diagnostics
	var dependencies []hcl.Traversal
	if content.Attributes[attribute] != nil {
		expr := content.Attributes[attribute].Expr
		dependencies, diags = decodeExpression(expr, dest, ctx)
	}
	return &decodeResult{Diags: diags, Depends: dependencies}
}

func decodeExpression(expr hcl.Expression, dest interface{}, ctx *hcl.EvalContext) ([]hcl.Traversal, hcl.Diagnostics) {
	diags := gohcl.DecodeExpression(expr, ctx, dest)
	var dependencies []hcl.Traversal
	for _, diag := range diags {
		if IsMissingVariableError(diag) {
			// was this error caused by a missing dependency?
			dependencies = append(dependencies, expr.(*hclsyntax.ScopeTraversalExpr).Traversal)
		}
	}
	// if there were missing variable errors, suppress the errors and just return the dependencies
	if len(dependencies) > 0 {
		diags = nil
	}

	return dependencies, diags
}

func IsMissingVariableError(diag *hcl.Diagnostic) bool {
	return diag.Summary == unknownVariableError || diag.Summary == missingMapElement
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

//
//
//func parseModVersion(block *hcl.Block) (*modconfig.ModVersion, hcl.Diagnostics) {
//	var diags hcl.Diagnostics
//	var dest = &modconfig.ModVersion{}
//
//	diags = gohcl.DecodeBody(block.Body, nil, dest)
//	if diags.HasErrors() {
//		return nil, diags
//	}
//
//	return dest, nil
//}
//
//func parsePluginDependency(block *hcl.Block) (*modconfig.PluginDependency, hcl.Diagnostics) {
//	var diags hcl.Diagnostics
//	var dest = &modconfig.PluginDependency{}
//
//	diags = gohcl.DecodeBody(block.Body, nil, dest)
//	if diags.HasErrors() {
//		return nil, diags
//	}
//
//	return dest, nil
//}
