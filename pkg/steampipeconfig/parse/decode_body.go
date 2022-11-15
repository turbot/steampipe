package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func decodeHclBody(body hcl.Body, evalCtx *hcl.EvalContext, resource modconfig.HclResource) hcl.Diagnostics {
	var diags hcl.Diagnostics
	diags = gohcl.DecodeBody(body, evalCtx, resource)
	// handle any resulting diags, which may specify dependencies
	moreDiags := gohcl.DecodeBody(body, evalCtx, resource.GetHclResourceBase())
	diags = append(diags, moreDiags...)
	// check what other interfaces the resource supports and deserialise into their base objects
	if qp, ok := resource.(modconfig.QueryProvider); ok {
		moreDiags = gohcl.DecodeBody(body, evalCtx, qp.GetQueryProviderBase())
		diags = append(diags, moreDiags...)
	}
	return diags
}
