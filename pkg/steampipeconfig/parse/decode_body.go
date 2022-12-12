package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/steampipeconfig/parse/parse_hcl"
)

func decodeHclBody(body hcl.Body, evalCtx *hcl.EvalContext, resource any) hcl.Diagnostics {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	var diags hcl.Diagnostics
	diags = parse_hcl.DecodeBody(body, evalCtx, resource)

	for _, nestedStruct := range parse_hcl.GetNestedStructVals(resource) {
		moreDiags := decodeHclBody(body, evalCtx, nestedStruct)
		diags = append(diags, moreDiags...)
	}

	return diags
}
