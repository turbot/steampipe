package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/go-kit/hcl_helpers"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"reflect"
	"strings"
)

func decodeHclBody(body hcl.Body, evalCtx *hcl.EvalContext, resourceProvider modconfig.ResourceMapsProvider, resource modconfig.HclResource) (diags hcl.Diagnostics) {
	defer func() {
		if r := recover(); r != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "unexpected error in decodeHclBody",
				Detail:   helpers.ToError(r).Error()})
		}
	}()

	nestedStructs, moreDiags := getNestedStructValsRecursive(resource)
	diags = append(diags, moreDiags...)
	// get the schema for this resource
	schema := getResourceSchema(resource, nestedStructs)
	// handle invalid block types
	moreDiags = validateHcl(resource.BlockType(), body.(*hclsyntax.Body), schema)
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

	// resolve any resource references using the resource map, rather than relying on the EvalCtx
	// (which does not work with nested struct vals)
	moreDiags = resolveReferences(body, resourceProvider, resource)
	diags = append(diags, moreDiags...)
	return diags
}

// build the hcl schema for this resource
func getResourceSchema(resource modconfig.HclResource, nestedStructs []any) *hcl.BodySchema {
	t := reflect.TypeOf(helpers.DereferencePointer(resource))
	typeName := t.Name()

	if cachedSchema, ok := resourceSchemaCache[typeName]; ok {
		return cachedSchema
	}
	var res = &hcl.BodySchema{}

	// ensure we cache before returning
	defer func() {
		resourceSchemaCache[typeName] = res
	}()

	var schemas []*hcl.BodySchema

	// build schema for top level object
	schemas = append(schemas, getSchemaForStruct(t))

	// now get schemas for any nested structs (using cache)
	for _, nestedStruct := range nestedStructs {
		t := reflect.TypeOf(helpers.DereferencePointer(nestedStruct))
		typeName := t.Name()

		// is this cached?
		nestedStructSchema, schemaCached := resourceSchemaCache[typeName]
		if !schemaCached {
			nestedStructSchema = getSchemaForStruct(t)
			resourceSchemaCache[typeName] = nestedStructSchema
		}

		// add to our list of schemas
		schemas = append(schemas, nestedStructSchema)
	}

	// TODO handle duplicates and required/optional
	// now merge the schemas
	for _, s := range schemas {
		res.Blocks = append(res.Blocks, s.Blocks...)
		res.Attributes = append(res.Attributes, s.Attributes...)
	}

	// special cases for manually parsed attributes and blocks
	switch resource.BlockType() {
	case modconfig.BlockTypeMod:
		res.Blocks = append(res.Blocks, hcl.BlockHeaderSchema{Type: modconfig.BlockTypeRequire})
	case modconfig.BlockTypeDashboard, modconfig.BlockTypeContainer:
		res.Blocks = append(res.Blocks,
			hcl.BlockHeaderSchema{Type: modconfig.BlockTypeControl},
			hcl.BlockHeaderSchema{Type: modconfig.BlockTypeBenchmark},
			hcl.BlockHeaderSchema{Type: modconfig.BlockTypeCard},
			hcl.BlockHeaderSchema{Type: modconfig.BlockTypeChart},
			hcl.BlockHeaderSchema{Type: modconfig.BlockTypeContainer},
			hcl.BlockHeaderSchema{Type: modconfig.BlockTypeFlow},
			hcl.BlockHeaderSchema{Type: modconfig.BlockTypeGraph},
			hcl.BlockHeaderSchema{Type: modconfig.BlockTypeHierarchy},
			hcl.BlockHeaderSchema{Type: modconfig.BlockTypeImage},
			hcl.BlockHeaderSchema{Type: modconfig.BlockTypeInput},
			hcl.BlockHeaderSchema{Type: modconfig.BlockTypeTable},
			hcl.BlockHeaderSchema{Type: modconfig.BlockTypeText},
			hcl.BlockHeaderSchema{Type: modconfig.BlockTypeWith},
		)
	case modconfig.BlockTypeQuery:
		// remove `Query` from attributes
		var querySchema = &hcl.BodySchema{}
		for _, a := range res.Attributes {
			if a.Name != modconfig.AttributeQuery {
				querySchema.Attributes = append(querySchema.Attributes, a)
			}
		}
		res = querySchema
	}

	if _, ok := resource.(modconfig.QueryProvider); ok {
		res.Blocks = append(res.Blocks, hcl.BlockHeaderSchema{Type: modconfig.BlockTypeParam})
		// if this is NOT query, add args
		if resource.BlockType() != modconfig.BlockTypeQuery {
			res.Attributes = append(res.Attributes, hcl.AttributeSchema{Name: modconfig.AttributeArgs})
		}
	}
	if _, ok := resource.(modconfig.NodeAndEdgeProvider); ok {
		res.Blocks = append(res.Blocks,
			hcl.BlockHeaderSchema{Type: modconfig.BlockTypeCategory},
			hcl.BlockHeaderSchema{Type: modconfig.BlockTypeNode},
			hcl.BlockHeaderSchema{Type: modconfig.BlockTypeEdge})
	}
	if _, ok := resource.(modconfig.WithProvider); ok {
		res.Blocks = append(res.Blocks, hcl.BlockHeaderSchema{Type: modconfig.BlockTypeWith})
	}
	return res
}

func getSchemaForStruct(t reflect.Type) *hcl.BodySchema {
	var schema = &hcl.BodySchema{}
	// get all hcl tags
	for i := 0; i < t.NumField(); i++ {
		tag := t.FieldByIndex([]int{i}).Tag.Get("hcl")
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

// rather than relying on the evaluation context to resolve resource references
// (which has the issue that when deserializing from cty we do not receive all base struct values)
// instead resolve the reference by parsing the resource name and finding the resource in the ResourceMap
// and use this resource to set the target property
func resolveReferences(body hcl.Body, resourceMapsProvider modconfig.ResourceMapsProvider, val any) (diags hcl.Diagnostics) {
	defer func() {
		if r := recover(); r != nil {
			if r := recover(); r != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "unexpected error in resolveReferences",
					Detail:   helpers.ToError(r).Error()})
			}
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
						path := hcl_helpers.TraversalAsString(scopeTraversal.Traversal)
						if parsedName, err := modconfig.ParseResourceName(path); err == nil {
							if r, ok := resourceMapsProvider.GetResource(parsedName); ok {
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
	return nil
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

func getNestedStructValsRecursive(val any) ([]any, hcl.Diagnostics) {
	nested, diags := getNestedStructVals(val)
	res := nested

	for _, n := range nested {
		nestedVals, moreDiags := getNestedStructValsRecursive(n)
		diags = append(diags, moreDiags...)
		res = append(res, nestedVals...)
	}
	return res, diags

}

// GetNestedStructVals return a slice of any nested structs within val
func getNestedStructVals(val any) (_ []any, diags hcl.Diagnostics) {
	defer func() {
		if r := recover(); r != nil {
			if r := recover(); r != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "unexpected error in resolveReferences",
					Detail:   helpers.ToError(r).Error()})
			}
		}
	}()

	rv := reflect.ValueOf(val)
	for rv.Type().Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	ty := rv.Type()
	if ty.Kind() != reflect.Struct {
		return nil, nil
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
	return res, nil
}
