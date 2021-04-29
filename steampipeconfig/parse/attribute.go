package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func parseAttribute(name string, dest interface{}, content *hcl.BodyContent, ctx *hcl.EvalContext) *decodeResult {
	var diags hcl.Diagnostics
	var dependencies []hcl.Traversal
	if content.Attributes[name] != nil {
		expr := content.Attributes[name].Expr
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

const unknownVariableError = "Unknown variable"
const missingMapElement = "Missing map element"

func IsMissingVariableError(diag *hcl.Diagnostic) bool {
	return diag.Summary == unknownVariableError || diag.Summary == missingMapElement
}
