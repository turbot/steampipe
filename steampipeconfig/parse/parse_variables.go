package parse

import (
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/zclconf/go-cty/cty"
)

func ParseVariables(fileData map[string][]byte) (map[string]cty.Value, error) {
	res := make(map[string]cty.Value)

	body, diags := ParseHclFiles(fileData)
	if diags.HasErrors() {
		return res, plugin.DiagsToError("Failed to parse variables file data", diags)
	}

	attrs, attrDiags := body.JustAttributes()
	diags = append(diags, attrDiags...)
	if attrs == nil {
		return res, diags
	}

	for name, attr := range attrs {
		val, valDiags := attr.Expr.Value(nil)
		diags = append(diags, valDiags...)
		res[name] = val
	}
	return res, nil
}
