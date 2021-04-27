package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Local represents a single entry from a "locals" block in a module or file.
// The "locals" block itself is not represented, because it serves only to
// provide context for us to interpret its contents.
type Local struct {
	Name      string
	Expr      hcl.Expression
	DeclRange hcl.Range
}

func decodeLocalsBlock(block *hcl.Block) ([]*Local, hcl.Diagnostics) {
	attrs, diags := block.Body.JustAttributes()
	if len(attrs) == 0 {
		return nil, diags
	}

	locals := make([]*Local, 0, len(attrs))
	for name, attr := range attrs {
		if !hclsyntax.ValidIdentifier(name) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid local value name",
				//Detail:   badIdentifierDetail,
				Subject: &attr.NameRange,
			})
		}

		//diags = append(diags, moreDiags...)

		locals = append(locals, &Local{
			Name:      name,
			Expr:      attr.Expr,
			DeclRange: attr.Range,
		})
	}
	return locals, diags
}
