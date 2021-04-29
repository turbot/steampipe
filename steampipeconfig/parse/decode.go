package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// A consistent detail message for all "not a valid identifier" diagnostics.
const badIdentifierDetail = "A name must start with a letter or underscore and may contain only letters, digits, underscores, and dashes."

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
		blockType := modconfig.ModBlockType(block.Type)
		switch blockType {
		case modconfig.BlockTypeMod:
			// pass the shell mod - it will be mutated
			res := decodeMod(block, runCtx.Mod, runCtx.EvalCtx)
			if res.Diags.HasErrors() {
				diags = append(diags, res.Diags...)
			}
			if len(res.Depends) > 0 {
				runCtx.AddDependencies(block, runCtx.Mod.Name(), res.Depends)
			}

		case modconfig.BlockTypeQuery:
			query, res := decodeQuery(block, runCtx.EvalCtx)
			diags = append(diags, handleDecodeResult(query, res, block, runCtx)...)

		case modconfig.BlockTypeControl:
			control, res := decodeControl(block, runCtx.EvalCtx)
			diags = append(diags, handleDecodeResult(control, res, block, runCtx)...)

		case modconfig.BlockTypeControlGroup:
			controlGroup, res := decodeControlGroup(block, runCtx.EvalCtx)
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
		metadata := GetMetadataForParsedResource(resource.Name(), block, runCtx.FileData, runCtx.Mod)
		resource.SetMetadata(metadata)
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

func decodeQuery(block *hcl.Block, ctx *hcl.EvalContext) (*modconfig.Query, *decodeResult) {
	query := &modconfig.Query{
		ShortName: block.Labels[0],
		DeclRange: block.DefRange,
	}

	content, diags := block.Body.Content(query.Schema())
	if diags.HasErrors() {
		return nil, &decodeResult{Diags: diags}
	}

	if !hclsyntax.ValidIdentifier(query.ShortName) {
		return nil, &decodeResult{
			Diags: hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid query name",
				Detail:   badIdentifierDetail,
				Subject:  &block.LabelRanges[0],
			}},
		}
	}

	res := &decodeResult{}
	for attribute, dest := range modconfig.HclProperties(query) {
		res.Merge(parseAttribute(attribute, dest, content, ctx))
	}
	return query, res
}

func decodeControl(block *hcl.Block, ctx *hcl.EvalContext) (*modconfig.Control, *decodeResult) {
	control := &modconfig.Control{
		ShortName: block.Labels[0],
		DeclRange: block.DefRange,
	}

	content, diags := block.Body.Content(control.Schema())
	if diags.HasErrors() {
		return nil, &decodeResult{Diags: diags}
	}

	if !hclsyntax.ValidIdentifier(control.ShortName) {
		return nil, &decodeResult{
			Diags: hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid query name",
				Detail:   badIdentifierDetail,
				Subject:  &block.LabelRanges[0],
			}},
		}
	}

	res := &decodeResult{}
	for attribute, dest := range modconfig.HclProperties(control) {
		res.Merge(parseAttribute(attribute, dest, content, ctx))
	}
	return control, res
}

func decodeControlGroup(block *hcl.Block, ctx *hcl.EvalContext) (*modconfig.ControlGroup, *decodeResult) {
	controlGroup := &modconfig.ControlGroup{
		ShortName: block.Labels[0],
		DeclRange: block.DefRange,
	}

	content, diags := block.Body.Content(controlGroup.Schema())
	if diags.HasErrors() {
		return nil, &decodeResult{Diags: diags}
	}

	if !hclsyntax.ValidIdentifier(controlGroup.ShortName) {
		return nil, &decodeResult{
			Diags: hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid query name",
				Detail:   badIdentifierDetail,
				Subject:  &block.LabelRanges[0],
			}},
		}
	}

	res := &decodeResult{}
	for attribute, dest := range modconfig.HclProperties(controlGroup) {
		res.Merge(parseAttribute(attribute, dest, content, ctx))
	}
	return controlGroup, res
}

func decodeLocals(block *hcl.Block, ctx *hcl.EvalContext) ([]*modconfig.Local, *decodeResult) {
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
			ShortName: name,
			Value:     val,
			DeclRange: attr.Range,
		})
	}
	return locals, &decodeResult{Diags: diags}
}

func decodeMod(block *hcl.Block, mod *modconfig.Mod, ctx *hcl.EvalContext) *decodeResult {
	mod.ShortName = block.Labels[0]
	content, diags := block.Body.Content(mod.Schema())
	if diags.HasErrors() {
		return &decodeResult{Diags: diags}
	}

	if !hclsyntax.ValidIdentifier(mod.ShortName) {
		return &decodeResult{
			Diags: hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid query name",
				Detail:   badIdentifierDetail,
				Subject:  &block.LabelRanges[0],
			}},
		}
	}

	res := &decodeResult{}
	for attribute, dest := range modconfig.HclProperties(mod) {
		res.Merge(parseAttribute(attribute, dest, content, ctx))
	}

	for _, block := range content.Blocks {
		switch block.Type {
		// TODO add parsing of requires block
		case "opengraph":
			opengraph, res := decodeOpenGraph(block, ctx)
			res.Merge(res)
			if res.Success() {
				mod.OpenGraph = opengraph
			}
		}
	}

	return res
}

func decodeOpenGraph(block *hcl.Block, ctx *hcl.EvalContext) (*modconfig.OpenGraph, *decodeResult) {
	res := &decodeResult{}

	opengraph := &modconfig.OpenGraph{}

	content, diags := block.Body.Content(opengraph.Schema())
	if diags.HasErrors() {
		return nil, &decodeResult{Diags: diags}
	}

	for attribute, dest := range modconfig.HclProperties(opengraph) {
		res.Merge(parseAttribute(attribute, dest, content, ctx))
	}

	return opengraph, res
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
