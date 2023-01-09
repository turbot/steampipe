package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/hclhelpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"reflect"
	"strings"
)

func decodeHclBody(body hcl.Body, evalCtx *hcl.EvalContext, resourceProvider modconfig.ResourceMapsProvider, resource any) hcl.Diagnostics {
	defer func() {
		if r := recover(); r != nil {
			// TODO ADD DIAG
			fmt.Println(r)
		}
	}()
	var diags hcl.Diagnostics

	nestedStructs := getNestedStructValsRecursive(resource)

	// get the schema for this resource
	schema := getResourceSchema(resource, nestedStructs)
	// handle invalid block types
	moreDiags := validateHcl(body.(*hclsyntax.Body), schema)
	diags = append(diags, moreDiags...)

	moreDiags = decodeHclBodyIntoStruct(body, evalCtx, resourceProvider, resource)
	diags = append(diags, moreDiags...)

	for _, nestedStruct := range nestedStructs {
		moreDiags := decodeHclBodyIntoStruct(body, evalCtx, resourceProvider, nestedStruct)
		diags = append(diags, moreDiags...)
	}

	return diags
}

func decodeHclBodyIntoStruct(body hcl.Body, evalCtx *hcl.EvalContext, resourceProvider modconfig.ResourceMapsProvider, resource any) hcl.Diagnostics {
	var diags hcl.Diagnostics
	// call decodeHclBodyIntoStruct to do actual decode
	moreDiags := gohcl.DecodeBody(body, evalCtx, resource)
	diags = append(diags, moreDiags...)

	// TODO WHAT DOES THIS DO?????
	resolveReferences(body, resourceProvider, resource)
	diags = append(diags, moreDiags...)
	return diags
}

func getResourceSchema(resource any, nestedStructs []any) *hcl.BodySchema {
	var schemas []*hcl.BodySchema

	// build schema for top level object
	schemas = append(schemas, getSchemaForStruct(resource))
	for _, nestedStruct := range nestedStructs {
		schemas = append(schemas, getSchemaForStruct(nestedStruct))
	}

	// TODO handle duplicates and required/optional
	// now merge the schemas
	var res = &hcl.BodySchema{}
	for _, s := range schemas {
		for _, b := range s.Blocks {
			res.Blocks = append(res.Blocks, b)
		}
		for _, a := range s.Attributes {
			res.Attributes = append(res.Attributes, a)
		}
	}

	/* special cases for manually parsed attributes and blocks
	mod require block
	*/
	switch resource.(type) {
	case *modconfig.Mod:
		res.Blocks = append(res.Blocks, hcl.BlockHeaderSchema{Type: modconfig.BlockTypeRequire})
	}

	return res
}

func getSchemaForStruct(s any) *hcl.BodySchema {
	v := reflect.TypeOf(helpers.DereferencePointer(s))

	typeName := v.Name()
	if cachedSchema, ok := resourceSchemaCache[typeName]; ok {
		return cachedSchema
	}
	var schema = &hcl.BodySchema{}
	// ensure we cache before returning
	defer func() {
		resourceSchemaCache[typeName] = schema
	}()

	// get all hcl tags
	for i := 0; i < v.NumField(); i++ {
		tag := v.FieldByIndex([]int{i}).Tag.Get("hcl")
		if tag == "" {
			continue
		}
		if idx := strings.LastIndex(tag, ",block"); idx != -1 {
			blockName := tag[:idx]
			schema.Blocks = append(schema.Blocks, hcl.BlockHeaderSchema{Type: blockName})
		} else {
			attributeName := strings.Split(tag, ",")[0]
			if attributeName != "" {
				schema.Attributes = append(schema.Attributes, hcl.AttributeSchema{Name: attributeName})
			}
		}
	}
	return schema
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

func getNestedStructValsRecursive(val any) []any {
	nested := getNestedStructVals(val)
	res := nested

	for _, n := range nested {
		res = append(res, getNestedStructValsRecursive(n)...)
	}
	return res

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
