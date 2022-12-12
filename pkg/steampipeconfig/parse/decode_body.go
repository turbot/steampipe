package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"reflect"
)
func decodeHclBody(body hcl.Body, evalCtx *hcl.EvalContext, resource any) hcl.Diagnostics {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	var diags hcl.Diagnostics
	diags = DecodeBody(body, evalCtx, resource)

	for _, nestedStruct := range getNestedStructVals(resource) {
		moreDiags := decodeHclBody(body, evalCtx, nestedStruct)
		diags = append(diags, moreDiags...)
	}

	return diags
}

func getNestedStructVals(val any) []any {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("getNestedStructVals", r)
		}
	}()

	rv := reflect.ValueOf(val)
	for rv.Type().Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	ty := rv.Type()
	if ty.Kind() != reflect.Struct {
		return nil
	}
	ct := ty.NumField()
	var res []any
	for i := 0; i < ct; i++ {
		field := ty.Field(i)

		tag := field.Tag.Get("hcl")
		if tag == "" {
			fieldVal := rv.Field(i)
			if field.Anonymous && fieldVal.Kind() == reflect.Struct {
				res = append(res, fieldVal.Addr().Interface())
			}
		}
	}
	return res
}

// TODO use when serialising
func getNestedCtyValueProviders(val any) []any {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	var res []any
	for _, i := range getNestedStructVals(val) {
		if _, ok := i.(modconfig.CtyValueProvider); ok {
			res = append(res, i)
		}
	}
	return res
}

//reflect: call of reflect.Value.Type on zero Value

//func isStruct(val any) bool {
//	defer func() {
//		if r := recover(); r != nil {
//			fmt.Println(r)
//		}
//	}()
//
//	rv := reflect.ValueOf(val)
//	for rv.Type().Kind() == reflect.Pointer {
//		rv = rv.Elem()
//		if rv.IsZero() {
//			return false
//		}
//	}
//	ty := rv.Type()
//	return ty.Kind() == reflect.Struct
//
//}

//
//func getNestedStructValsForCty_refactor(val any) []any {
//	rv := reflect.ValueOf(val)
//	log.Printf("[WARN] getNestedStructValsForCty_refactor : val %v kind  %v, type %s", val, rv.Kind(), reflect.TypeOf(val).String())
//	for rv.Kind() == reflect.Ptr {
//		log.Printf("[WARN] getNestedStructValsForCty_refactor : NOT PTR")
//		//rv = rv.Addr()
//		rv = rv.Elem()
//	}
//	ty := rv.Type()
//	if ty.Kind() != reflect.Struct {
//		log.Printf("[WARN] getNestedStructValsForCty_refactor : NOT STRUCT!!! %v", ty.Kind())
//		return nil
//	}
//	ct := ty.NumField()
//	var res []any
//	for i := 0; i < ct; i++ {
//		field := ty.Field(i)
//
//		tag := field.Tag.Get("hcl")
//		if tag == "" {
//			fieldVal := rv.Field(i)
//			if field.Anonymous && fieldVal.Kind() == reflect.Struct {
//				// ensure the nested struct has cty tags
//				fieldValTy := fieldVal.Type()
//				containsCty := false
//				for i := 0; i < fieldValTy.NumField(); i++ {
//					field := fieldValTy.Field(i)
//					if tag := field.Tag.Get("cty"); tag != "" {
//						containsCty = true
//						break
//					}
//				}
//				if containsCty {
//					res = append(res, fieldVal.Addr().Interface())
//				}
//			}
//		}
//	}
//	return res
//}
