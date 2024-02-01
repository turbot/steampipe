package modconfig

import (
	"log"
	"reflect"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/terraform-components/configs/configschema"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// GetCtyTypes builds a map of cty types for all tagged properties.
// It is used to convert the struct to a cty value
func GetCtyTypes(item interface{}) map[string]cty.Type {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[WARN] GetCtyTypes failed with panic: %v", r)
		}
	}()
	var res = make(map[string]cty.Type)

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

			res[attribute] = ctyType
		}
	}
	return res
}

// GetCtyValue converts the item into a cty value
func GetCtyValue(item interface{}) (cty.Value, error) {
	// TODO  [node_reuse] look at merging base struct schemas

	// build the block schema
	var block = configschema.Block{Attributes: make(map[string]*configschema.Attribute)}

	// get the hcl attributes - these include the cty type
	for attribute, ctyType := range GetCtyTypes(item) {
		// TODO how to determine optional?
		block.Attributes[attribute] = &configschema.Attribute{Optional: true, Type: ctyType}
	}

	// get cty spec
	spec := block.DecoderSpec()
	ty := hcldec.ImpliedType(spec)

	return gocty.ToCtyValue(item, ty)
}
