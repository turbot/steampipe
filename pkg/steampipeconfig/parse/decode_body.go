package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe/pkg/steampipeconfig/hclhelpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"reflect"
	"strings"
)

func decodeHclBody(body hcl.Body, evalCtx *hcl.EvalContext, resourceProvider modconfig.ResourceMapsProvider, resource any) hcl.Diagnostics {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	var diags hcl.Diagnostics
	diags = gohcl.DecodeBody(body, evalCtx, resource)

	resolveReferences(body, resourceProvider, resource)
	for _, nestedStruct := range getNestedStructVals(resource) {
		moreDiags := decodeHclBody(body, evalCtx, resourceProvider, nestedStruct)
		diags = append(diags, moreDiags...)
	}

	return diags
}

func resolveReferences(body hcl.Body, resourceMapsProvider modconfig.ResourceMapsProvider, val any) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("resolveReferences", r)
		}
	}()
	attributes := body.(*hclsyntax.Body).Attributes
	rv := reflect.ValueOf(val)
	for rv.Type().Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	ty := rv.Type()
	if ty.Kind() != reflect.Struct {
		return
	}

	ct := ty.NumField()
	for i := 0; i < ct; i++ {
		field := ty.Field(i)
		fieldVal := rv.Field(i)
		// get hcl attribute tag (if any) tag
		hclAttribute := getHclAttributeTag(field)
		if hclAttribute == "" {
			continue
		}
		if fieldVal.Type().Kind() == reflect.Pointer && !fieldVal.IsNil() {
			fieldVal = fieldVal.Elem()
		}
		if fieldVal.Kind() == reflect.Struct {
			v := fieldVal.Addr().Interface()
			if _, ok := v.(modconfig.HclResource); ok {
				if hclVal, ok := attributes[hclAttribute]; ok {
					if scopeTraversal, ok := hclVal.Expr.(*hclsyntax.ScopeTraversalExpr); ok {
						path := hclhelpers.TraversalAsString(scopeTraversal.Traversal)
						if parsedName, err := modconfig.ParseResourceName(path); err == nil {
							if r, ok := modconfig.GetResource(resourceMapsProvider, parsedName); ok {
								f := rv.FieldByName(field.Name)
								if f.IsValid() && f.CanSet() {
									targetVal := reflect.ValueOf(r)
									f.Set(targetVal)
								}
							}
						}
					}
				}
			}
		}
	}
}

func getHclAttributeTag(field reflect.StructField) string {
	tag := field.Tag.Get("hcl")
	if tag == "" {
		return ""
	}

	comma := strings.Index(tag, ",")
	var name, kind string
	if comma != -1 {
		name = tag[:comma]
		kind = tag[comma+1:]
	} else {
		name = tag
		kind = "attr"
	}

	switch kind {
	case "attr":
		return name
	default:
		return ""
	}
}

// GetNestedStructVals return a slice of any nested structs within val
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
		fieldVal := rv.Field(i)
		if field.Anonymous && fieldVal.Kind() == reflect.Struct {
			res = append(res, fieldVal.Addr().Interface())
		}
	}
	return res
}
