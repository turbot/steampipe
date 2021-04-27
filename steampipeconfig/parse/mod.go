package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

func ParseMod(block *hcl.Block, mod *modconfig.Mod, ctx *hcl.EvalContext) ([]hcl.Traversal, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	content, diags := block.Body.Content(modSchema)
	if diags.HasErrors() {
		return nil, diags
	}

	depends, moreDiags := parseModAttributes(content, mod, ctx)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
	}

	for _, block := range content.Blocks {
		switch block.Type {
		// TODO add parsing of requires block
		//case "mod_depends":
		//	modDependency, moreDiags := parseModVersion(block)
		//	if moreDiags.HasErrors() {
		//		diags = append(diags, moreDiags...)
		//		break
		//	}
		//	mod.ModDepends = append(mod.ModDepends, modDependency)
		//case "plugin_depends":
		//	pluginDependency, moreDiags := parsePluginDependency(block)
		//	if moreDiags.HasErrors() {
		//		diags = append(diags, moreDiags...)
		//		break
		//	}
		//	mod.PluginDepends = append(mod.PluginDepends, pluginDependency)

		case "opengraph":
			opengraph, moreDepends, moreDiags := parseOpenGraph(block, ctx)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				break
			}
			depends = append(depends, moreDepends...)
			mod.OpenGraph = opengraph
		}
	}

	return depends, diags
}

func parseModAttributes(content *hcl.BodyContent, mod *modconfig.Mod, ctx *hcl.EvalContext) ([]hcl.Traversal, hcl.Diagnostics) {

	var diags hcl.Diagnostics
	var depends []hcl.Traversal

	moreDepends, moreDiags := parseAttribute("color", &mod.Color, content, ctx)
	diags = append(diags, moreDiags...)
	depends = append(depends, moreDepends...)

	moreDepends, diags = parseAttribute("description", &mod.Description, content, ctx)
	diags = append(diags, moreDiags...)
	depends = append(depends, moreDepends...)

	moreDepends, diags = parseAttribute("documentation", &mod.Description, content, ctx)
	diags = append(diags, moreDiags...)
	depends = append(depends, moreDepends...)

	moreDepends, diags = parseAttribute("icon", &mod.Description, content, ctx)
	diags = append(diags, moreDiags...)
	depends = append(depends, moreDepends...)

	moreDepends, diags = parseAttribute("labels", &mod.Description, content, ctx)
	diags = append(diags, moreDiags...)
	depends = append(depends, moreDepends...)

	moreDepends, diags = parseAttribute("title", &mod.Description, content, ctx)
	diags = append(diags, moreDiags...)
	depends = append(depends, moreDepends...)

	return depends, diags
}

func parseAttribute(name string, dest interface{}, content *hcl.BodyContent, ctx *hcl.EvalContext) ([]hcl.Traversal, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var dependencies []hcl.Traversal
	if content.Attributes[name] != nil {
		expr := content.Attributes[name].Expr
		dependencies, diags = decodeExpression(expr, dest, ctx)

	}
	return dependencies, diags
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

const unknownVariableError = "Unknown variable"

func IsMissingVariableError(diag *hcl.Diagnostic) bool {
	return diag.Summary == unknownVariableError
}

func parseModVersion(block *hcl.Block) (*modconfig.ModVersion, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var dest = &modconfig.ModVersion{}

	diags = gohcl.DecodeBody(block.Body, nil, dest)
	if diags.HasErrors() {
		return nil, diags
	}

	return dest, nil
}

func parsePluginDependency(block *hcl.Block) (*modconfig.PluginDependency, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var dest = &modconfig.PluginDependency{}

	diags = gohcl.DecodeBody(block.Body, nil, dest)
	if diags.HasErrors() {
		return nil, diags
	}

	return dest, nil
}

func parseOpenGraph(block *hcl.Block, ctx *hcl.EvalContext) (*modconfig.OpenGraph, []hcl.Traversal, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var dest = &modconfig.OpenGraph{}

	diags = gohcl.DecodeBody(block.Body, ctx, dest)
	if diags.HasErrors() {
		return nil, nil, diags
	}

	return dest, nil, nil
}
