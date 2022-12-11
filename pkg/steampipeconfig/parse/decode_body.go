package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func decodeHclBody(body hcl.Body, evalCtx *hcl.EvalContext, resource modconfig.HclResource) hcl.Diagnostics {
	var diags hcl.Diagnostics
	diags = gohcl.DecodeBody(body, evalCtx, resource)

	moreDiags := gohcl.DecodeBody(body, evalCtx, resource.GetHclResourceBase())
	diags = append(diags, moreDiags...)

	// check what other interfaces the resource supports and deserialise into their base objects
	if qp, ok := resource.(modconfig.QueryProvider); ok {
		moreDiags = gohcl.DecodeBody(body, evalCtx, qp.GetQueryProviderBase())
		diags = append(diags, moreDiags...)
	}
	return diags
	// to

}

//
//func decodeHclBody(body hcl.Body, evalCtx *hcl.EvalContext, resource any) hcl.Diagnostics {
//
//	var diags hcl.Diagnostics
//	diags = gohcl.DecodeBody(body, evalCtx, resource)
//
//	for _, nestedStruct := range getNestedStructVals(resource) {
//		moreDiags := decodeHclBody(body, evalCtx, nestedStruct)
//		diags = append(diags, moreDiags...)
//	}
//
//	return diags
//}
//
//func getNestedStructVals(val any) []any {
//	rv := reflect.ValueOf(val)
//	if rv.Kind() != reflect.Ptr {
//		return nil
//	}
//	rv = rv.Elem()
//	ty := rv.Type()
//	ct := ty.NumField()
//	var res []any
//	for i := 0; i < ct; i++ {
//		field := ty.Field(i)
//
//		tag := field.Tag.Get("hcl")
//		if tag == "" {
//			fieldVal := rv.Field(i)
//			if field.Anonymous && fieldVal.Kind() == reflect.Struct {
//				res = append(res, fieldVal.Addr().Interface())
//			}
//		}
//	}
//	return res
//}
