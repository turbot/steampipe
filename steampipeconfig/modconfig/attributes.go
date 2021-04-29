package modconfig

import (
	"log"
	"reflect"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/turbot/go-kit/helpers"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type AttributeDetails struct {
	Attribute string
	Dest      interface{}
	CtyType   cty.Type
}

// GetAttributeDetails
// all properties parsed from hcl have a cty tag
// return map of property name to the pointer to the destination property
// this is used to populate a control during decoding and build the cty schema
func GetAttributeDetails(item interface{}) map[string]AttributeDetails {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[WARN] GetAttributeDetails failed with panic: %v", r)
		}
	}()
	var res = make(map[string]AttributeDetails)

	t := reflect.TypeOf(helpers.DereferencePointer(item))
	val := reflect.ValueOf(item)
	if val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		structField := t.Field(i)
		attribute, ok := structField.Tag.Lookup("cty")
		if ok && attribute != "-" {
			valField := val.Field(i)
			// get cty type
			ctyType, err := gocty.ImpliedType(valField.Interface())
			if err != nil {
				panic(err)
			}
			// get pointer to property
			dest := valField.Addr().Interface()
			// store as AttributeDetails
			res[attribute] = AttributeDetails{
				Attribute: attribute,
				Dest:      dest,
				CtyType:   ctyType,
			}
		}
	}
	return res
}

// convert the item into a cty value.
func getCtyValue(item interface{}) (cty.Value, error) {
	// build the block schema
	var block = configschema.Block{Attributes: make(map[string]*configschema.Attribute)}

	// get the hcl attributes - these include the cty type
	for attribute, details := range GetAttributeDetails(item) {
		// TODO how to determine optional?
		block.Attributes[attribute] = &configschema.Attribute{Optional: true, Type: details.CtyType}
	}

	// get cty spec
	spec := block.DecoderSpec()
	ty := hcldec.ImpliedType(spec)

	return gocty.ToCtyValue(item, ty)
}
