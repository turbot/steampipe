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
		// check name is valid
		moreDiags := validateName(block)
		if diags.HasErrors() {
			diags = append(diags, moreDiags...)
			continue
		}

		// special case decoding for locals
		if block.Type == modconfig.BlockTypeLocals {
			// special case decode logic for locals
			locals, res := decodeLocals(block, runCtx.EvalCtx)
			for _, local := range locals {
				// handle the result
				// - if successful, add resource to mod and variables maps
				// - if there are dependencies, add them to run context
				moreDiags = handleDecodeResult(local, res, block, runCtx)
				diags = append(diags, moreDiags...)
			}
			continue
		}

		// all other blocks are treated the same:
		// decode the resource
		resource, res := decodeResource(block, runCtx)

		// handle the result
		// - if successful, add resource to mod and variables maps
		// - if there are dependencies, add them to run context
		moreDiags = handleDecodeResult(resource, res, block, runCtx)
		diags = append(diags, moreDiags...)
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
	case modconfig.BlockTypeControlGroup:
		resource = modconfig.NewControlGroup(block)
	}
	return resource
}

func decodeLocals(block *hcl.Block, ctx *hcl.EvalContext) ([]*modconfig.Local, *decodeResult) {
	attrs, diags := block.Body.JustAttributes()
	if len(attrs) == 0 {
		return nil, &decodeResult{Diags: diags}
	}

	// build list of locals
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

		// add to our list
		locals = append(locals, modconfig.NewLocal(name, val, attr.Range))
	}
	return locals, &decodeResult{Diags: diags}
}

func decodeResource(block *hcl.Block, runCtx *RunContext) (modconfig.HclResource, *decodeResult) {
	// get shell resource
	resource := resourceForBlock(block, runCtx)

	res := &decodeResult{}
	moreDiags := gohcl.DecodeBody(block.Body, runCtx.EvalCtx, resource)
	for _, diag := range moreDiags {
		if IsMissingVariableError(diag) {
			// was this error caused by a missing dependency?
			res.Depends = append(res.Depends, diag.Expression.Variables()...)
		}
	}
	return resource, res
}

func handleDecodeResult(resource modconfig.HclResource, res *decodeResult, block *hcl.Block, runCtx *RunContext) hcl.Diagnostics {
	var diags hcl.Diagnostics
	if res.Success() {
		// if resource supports metadata, save it
		if resourceWithMetadata, ok := resource.(modconfig.ResourceWithMetadata); ok {
			metadata := GetMetadataForParsedResource(resource.Name(), block, runCtx.FileData, runCtx.Mod)
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
			runCtx.AddDependencies(block, resource.Name(), res.Depends)
		}
	}
	return diags
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
