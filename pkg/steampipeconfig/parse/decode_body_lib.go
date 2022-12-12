package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/gocty"
	"reflect"
	"sort"
	"strings"
)

var victimExpr hcl.Expression
var victimBody hcl.Body

var exprType = reflect.TypeOf(&victimExpr).Elem()
var bodyType = reflect.TypeOf(&victimBody).Elem()
var blockType = reflect.TypeOf((*hcl.Block)(nil))
var attrType = reflect.TypeOf((*hcl.Attribute)(nil))
var attrsType = reflect.TypeOf(hcl.Attributes(nil))

// DecodeBody extracts the configuration within the given body into the given
// value. This value must be a non-nil pointer to either a struct or
// a map, where in the former case the configuration will be decoded using
// struct tags and in the latter case only attributes are allowed and their
// values are decoded into the map.
//
// The given EvalContext is used to resolve any variables or functions in
// expressions encountered while decoding. This may be nil to require only
// constant values, for simple applications that do not support variables or
// functions.
//
// The returned diagnostics should be inspected with its HasErrors method to
// determine if the populated value is valid and complete. If error diagnostics
// are returned then the given value may have been partially-populated but
// may still be accessed by a careful caller for static analysis and editor
// integration use-cases.
func DecodeBody(body hcl.Body, ctx *hcl.EvalContext, val interface{}) hcl.Diagnostics {
	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("target value must be a pointer, not %s", rv.Type().String()))
	}

	return decodeBodyToValue(body, ctx, rv.Elem())
}

func decodeBodyToValue(body hcl.Body, ctx *hcl.EvalContext, val reflect.Value) hcl.Diagnostics {
	et := val.Type()
	switch et.Kind() {
	case reflect.Struct:
		return decodeBodyToStruct(body, ctx, val)
	case reflect.Map:
		return decodeBodyToMap(body, ctx, val)
	default:
		panic(fmt.Sprintf("target value must be pointer to struct or map, not %s", et.String()))
	}
}

func decodeBodyToStruct(body hcl.Body, ctx *hcl.EvalContext, val reflect.Value) hcl.Diagnostics {
	schema, partial := ImpliedBodySchema(val.Interface())

	var content *hcl.BodyContent
	var leftovers hcl.Body
	var diags hcl.Diagnostics
	if partial {
		content, leftovers, diags = body.PartialContent(schema)
	} else {
		content, diags = body.Content(schema)
	}
	if content == nil {
		return diags
	}

	tags := getFieldTags(val.Type())

	if tags.Body != nil {
		fieldIdx := *tags.Body
		field := val.Type().Field(fieldIdx)
		fieldV := val.Field(fieldIdx)
		switch {
		case bodyType.AssignableTo(field.Type):
			fieldV.Set(reflect.ValueOf(body))

		default:
			diags = append(diags, decodeBodyToValue(body, ctx, fieldV)...)
		}
	}

	if tags.Remain != nil {
		fieldIdx := *tags.Remain
		field := val.Type().Field(fieldIdx)
		fieldV := val.Field(fieldIdx)
		switch {
		case bodyType.AssignableTo(field.Type):
			fieldV.Set(reflect.ValueOf(leftovers))
		case attrsType.AssignableTo(field.Type):
			attrs, attrsDiags := leftovers.JustAttributes()
			if len(attrsDiags) > 0 {
				diags = append(diags, attrsDiags...)
			}
			fieldV.Set(reflect.ValueOf(attrs))
		default:
			diags = append(diags, decodeBodyToValue(leftovers, ctx, fieldV)...)
		}
	}

	for name, fieldIdx := range tags.Attributes {
		attr := content.Attributes[name]
		field := val.Type().Field(fieldIdx)
		fieldV := val.Field(fieldIdx)

		if attr == nil {
			if !exprType.AssignableTo(field.Type) {
				continue
			}

			// As a special case, if the target is of type hcl.Expression then
			// we'll assign an actual expression that evalues to a cty null,
			// so the caller can deal with it within the cty realm rather
			// than within the Go realm.
			synthExpr := hcl.StaticExpr(cty.NullVal(cty.DynamicPseudoType), body.MissingItemRange())
			fieldV.Set(reflect.ValueOf(synthExpr))
			continue
		}

		switch {
		case attrType.AssignableTo(field.Type):
			fieldV.Set(reflect.ValueOf(attr))
		case exprType.AssignableTo(field.Type):
			fieldV.Set(reflect.ValueOf(attr.Expr))
		default:
			diags = append(diags, DecodeExpressionNested(attr.Expr, ctx, fieldV.Addr().Interface())...)
		}
	}

	blocksByType := content.Blocks.ByType()

	for typeName, fieldIdx := range tags.Blocks {
		blocks := blocksByType[typeName]
		field := val.Type().Field(fieldIdx)

		ty := field.Type
		isSlice := false
		isPtr := false
		if ty.Kind() == reflect.Slice {
			isSlice = true
			ty = ty.Elem()
		}
		if ty.Kind() == reflect.Ptr {
			isPtr = true
			ty = ty.Elem()
		}

		if len(blocks) > 1 && !isSlice {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Duplicate %s block", typeName),
				Detail: fmt.Sprintf(
					"Only one %s block is allowed. Another was defined at %s.",
					typeName, blocks[0].DefRange.String(),
				),
				Subject: &blocks[1].DefRange,
			})
			continue
		}

		if len(blocks) == 0 {
			if isSlice || isPtr {
				if val.Field(fieldIdx).IsNil() {
					val.Field(fieldIdx).Set(reflect.Zero(field.Type))
				}
			} else {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Missing %s block", typeName),
					Detail:   fmt.Sprintf("A %s block is required.", typeName),
					Subject:  body.MissingItemRange().Ptr(),
				})
			}
			continue
		}

		switch {

		case isSlice:
			elemType := ty
			if isPtr {
				elemType = reflect.PtrTo(ty)
			}
			sli := val.Field(fieldIdx)
			if sli.IsNil() {
				sli = reflect.MakeSlice(reflect.SliceOf(elemType), len(blocks), len(blocks))
			}

			for i, block := range blocks {
				if isPtr {
					if i >= sli.Len() {
						sli = reflect.Append(sli, reflect.New(ty))
					}
					v := sli.Index(i)
					if v.IsNil() {
						v = reflect.New(ty)
					}
					diags = append(diags, decodeBlockToValue(block, ctx, v.Elem())...)
					sli.Index(i).Set(v)
				} else {
					if i >= sli.Len() {
						sli = reflect.Append(sli, reflect.Indirect(reflect.New(ty)))
					}
					diags = append(diags, decodeBlockToValue(block, ctx, sli.Index(i))...)
				}
			}

			if sli.Len() > len(blocks) {
				sli.SetLen(len(blocks))
			}

			val.Field(fieldIdx).Set(sli)

		default:
			block := blocks[0]
			if isPtr {
				v := val.Field(fieldIdx)
				if v.IsNil() {
					v = reflect.New(ty)
				}
				diags = append(diags, decodeBlockToValue(block, ctx, v.Elem())...)
				val.Field(fieldIdx).Set(v)
			} else {
				diags = append(diags, decodeBlockToValue(block, ctx, val.Field(fieldIdx))...)
			}

		}

	}

	return diags
}

func decodeBodyToMap(body hcl.Body, ctx *hcl.EvalContext, v reflect.Value) hcl.Diagnostics {
	attrs, diags := body.JustAttributes()
	if attrs == nil {
		return diags
	}

	mv := reflect.MakeMap(v.Type())

	for k, attr := range attrs {
		switch {
		case attrType.AssignableTo(v.Type().Elem()):
			mv.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(attr))
		case exprType.AssignableTo(v.Type().Elem()):
			mv.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(attr.Expr))
		default:
			ev := reflect.New(v.Type().Elem())
			diags = append(diags, DecodeExpressionNested(attr.Expr, ctx, ev.Interface())...)
			mv.SetMapIndex(reflect.ValueOf(k), ev.Elem())
		}
	}

	v.Set(mv)

	return diags
}

func decodeBlockToValue(block *hcl.Block, ctx *hcl.EvalContext, v reflect.Value) hcl.Diagnostics {
	diags := decodeBodyToValue(block.Body, ctx, v)

	if len(block.Labels) > 0 {
		blockTags := getFieldTags(v.Type())
		for li, lv := range block.Labels {
			lfieldIdx := blockTags.Labels[li].FieldIndex
			v.Field(lfieldIdx).Set(reflect.ValueOf(lv))
		}
	}

	return diags
}

func DecodeExpressionNested(expr hcl.Expression, ctx *hcl.EvalContext, val interface{}) hcl.Diagnostics {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	var diags hcl.Diagnostics

	//var shouldDecode = true
	//if isStruct(val) {
	//	_, shouldDecode = val.(modconfig.CtyValueProvider)
	//}
	//if shouldDecode {
	// TODO use gohcl fxn
	diags = decodeExpression(expr, ctx, val)
	if diags.HasErrors() {
		return diags
	}
	//}

	nested := getNestedStructVals(val)
	for _, v := range nested {
		moreDiags := DecodeExpressionNested(expr, ctx, v)
		diags = append(diags, moreDiags...)
	}
	return diags

}

//func containsTag(val interface{}, tagName string) (bool, reflect.Kind) {
//	// check the struct has the given tag
//	rv := reflect.ValueOf(val)
//	//rt := rv.Type()
//	kind := rv.Kind()
//	// dereference pointers
//	for kind == reflect.Pointer {
//		rv = rv.Elem()
//		kind = rv.Kind()
//		log.Printf("[WARN] containsTag kind %s", rv.Kind())
//	}
//
//	if kind != reflect.Struct {
//		return false, kind
//	}
//	for i := 0; i < rv.Type().NumField(); i++ {
//		field := rv.Type().Field(i)
//		if tag := field.Tag.Get(tagName); tag != "" {
//			return true, kind
//		}
//	}
//
//	return false, kind
//}

func decodeExpression(expr hcl.Expression, ctx *hcl.EvalContext, val interface{}) hcl.Diagnostics {
	var diags hcl.Diagnostics
	exprVal, diags := expr.Value(ctx)
	if diags.HasErrors() {
		return diags
	}
	convTy, err := gocty.ImpliedType(val)
	if err != nil {
		// we may hit this if we try to decode into a base struct with no cty tags
		return nil
		//panic(fmt.Sprintf("unsuitable DecodeExpression target: %s", err))
	}
	exprVal, err = convert.Convert(exprVal, convTy)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unsuitable value type",
			Detail:   fmt.Sprintf("Unsuitable value: %s", err.Error()),
			Subject:  expr.StartRange().Ptr(),
			Context:  expr.Range().Ptr(),
		})
		return diags
	}

	err = gocty.FromCtyValue(exprVal, val)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unsuitable value type",
			Detail:   fmt.Sprintf("Unsuitable value: %s", err.Error()),
			Subject:  expr.StartRange().Ptr(),
			Context:  expr.Range().Ptr(),
		})
	}

	return diags
}

// ImpliedBodySchema produces a hcl.BodySchema derived from the type of the
// given value, which must be a struct value or a pointer to one. If an
// inappropriate value is passed, this function will panic.
//
// The second return argument indicates whether the given struct includes
// a "remain" field, and thus the returned schema is non-exhaustive.
//
// This uses the tags on the fields of the struct to discover how each
// field's value should be expressed within configuration. If an invalid
// mapping is attempted, this function will panic.
func ImpliedBodySchema(val interface{}) (schema *hcl.BodySchema, partial bool) {
	ty := reflect.TypeOf(val)

	if ty.Kind() == reflect.Ptr {
		ty = ty.Elem()
	}

	if ty.Kind() != reflect.Struct {
		panic(fmt.Sprintf("given value must be struct, not %T", val))
	}

	var attrSchemas []hcl.AttributeSchema
	var blockSchemas []hcl.BlockHeaderSchema

	tags := getFieldTags(ty)

	attrNames := make([]string, 0, len(tags.Attributes))
	for n := range tags.Attributes {
		attrNames = append(attrNames, n)
	}
	sort.Strings(attrNames)
	for _, n := range attrNames {
		idx := tags.Attributes[n]
		optional := tags.Optional[n]
		field := ty.Field(idx)

		var required bool

		switch {
		case field.Type.AssignableTo(exprType):
			// If we're decoding to hcl.Expression then absense can be
			// indicated via a null value, so we don't specify that
			// the field is required during decoding.
			required = false
		case field.Type.Kind() != reflect.Ptr && !optional:
			required = true
		default:
			required = false
		}

		attrSchemas = append(attrSchemas, hcl.AttributeSchema{
			Name:     n,
			Required: required,
		})
	}

	blockNames := make([]string, 0, len(tags.Blocks))
	for n := range tags.Blocks {
		blockNames = append(blockNames, n)
	}
	sort.Strings(blockNames)
	for _, n := range blockNames {
		idx := tags.Blocks[n]
		field := ty.Field(idx)
		fty := field.Type
		if fty.Kind() == reflect.Slice {
			fty = fty.Elem()
		}
		if fty.Kind() == reflect.Ptr {
			fty = fty.Elem()
		}
		if fty.Kind() != reflect.Struct {
			panic(fmt.Sprintf(
				"hcl 'block' tag kind cannot be applied to %s field %s: struct required", field.Type.String(), field.Name,
			))
		}
		ftags := getFieldTags(fty)
		var labelNames []string
		if len(ftags.Labels) > 0 {
			labelNames = make([]string, len(ftags.Labels))
			for i, l := range ftags.Labels {
				labelNames[i] = l.Name
			}
		}

		blockSchemas = append(blockSchemas, hcl.BlockHeaderSchema{
			Type:       n,
			LabelNames: labelNames,
		})
	}

	partial = tags.Remain != nil
	schema = &hcl.BodySchema{
		Attributes: attrSchemas,
		Blocks:     blockSchemas,
	}
	return schema, partial
}

type fieldTags struct {
	Attributes map[string]int
	Blocks     map[string]int
	Labels     []labelField
	Remain     *int
	Body       *int
	Optional   map[string]bool
}

type labelField struct {
	FieldIndex int
	Name       string
}

func getFieldTags(ty reflect.Type) *fieldTags {
	ret := &fieldTags{
		Attributes: map[string]int{},
		Blocks:     map[string]int{},
		Optional:   map[string]bool{},
	}

	ct := ty.NumField()
	for i := 0; i < ct; i++ {
		field := ty.Field(i)
		tag := field.Tag.Get("hcl")
		if tag == "" {
			continue
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
			ret.Attributes[name] = i
		case "block":
			ret.Blocks[name] = i
		case "label":
			ret.Labels = append(ret.Labels, labelField{
				FieldIndex: i,
				Name:       name,
			})
		case "remain":
			if ret.Remain != nil {
				panic("only one 'remain' tag is permitted")
			}
			idx := i // copy, because this loop will continue assigning to i
			ret.Remain = &idx
		case "body":
			if ret.Body != nil {
				panic("only one 'body' tag is permitted")
			}
			idx := i // copy, because this loop will continue assigning to i
			ret.Body = &idx
		case "optional":
			ret.Attributes[name] = i
			ret.Optional[name] = true
		default:
			panic(fmt.Sprintf("invalid hcl field tag kind %q on %s %q", kind, field.Type.String(), field.Name))
		}
	}

	return ret
}
